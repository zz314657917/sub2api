package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

type apiKeyRouteRetryCache struct {
	APIKeyCache
	acquires int
	releases int
	cooling  bool
	acquire  func(context.Context, APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error)
	release  func(context.Context)
}

func (c *apiKeyRouteRetryCache) IsRouteGroupCooling(context.Context, int64, int64) (bool, error) {
	return c.cooling, nil
}

func (c *apiKeyRouteRetryCache) AcquireAPIKeyRouteBreaker(ctx context.Context, key APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error) {
	c.acquires++
	return c.acquire(ctx, key)
}

func (c *apiKeyRouteRetryCache) ReleaseAPIKeyRouteBreakerProbe(ctx context.Context, _ APIKeyRouteBreakerLease) error {
	c.releases++
	if c.release != nil {
		c.release(ctx)
	}
	return nil
}

func (c *apiKeyRouteRetryCache) RecordAPIKeyRouteBreakerSuccess(context.Context, APIKeyRouteBreakerLease) error {
	panic("selection must not record success")
}

func (c *apiKeyRouteRetryCache) RecordAPIKeyRouteBreakerFailure(context.Context, APIKeyRouteBreakerLease) error {
	panic("selection must not record failure")
}

func apiKeyRouteRetryKey() *APIKey {
	id := int64(1)
	return &APIKey{
		ID: 71, Key: "route-retry", GroupID: &id,
		User: &User{ID: 3, Status: StatusActive, Balance: 10},
		Group: &Group{
			ID: id, Status: StatusActive, Hydrated: true, Platform: PlatformOpenAI,
			ModelMatchPatterns: []string{"gpt-5.6"},
		},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{{GroupID: id, Enabled: true}},
	}
}

func TestAPIKeyRouteRetryBackoffAndAttemptCap(t *testing.T) {
	for recovery := 0; recovery <= 4; recovery++ {
		t.Run(string(rune('0'+recovery)), func(t *testing.T) {
			cache := &apiKeyRouteRetryCache{}
			cache.acquire = func(_ context.Context, key APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error) {
				if cache.acquires <= recovery {
					return nil, nil
				}
				return &APIKeyRouteBreakerLease{Key: key, Generation: 2, HalfOpen: true, ProbeToken: 8}, nil
			}
			svc := NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)
			var delays []time.Duration
			key := apiKeyRouteRetryKey()
			resolved, err := svc.resolveForModelRequestWithRetry(context.Background(), key, "/v1/responses", "", "gpt-5.6", false, func(ctx context.Context, delay time.Duration) error {
				require.NoError(t, ctx.Err())
				delays = append(delays, delay)
				return nil
			})
			require.Equal(t, min(recovery+1, 4), cache.acquires, "classifier must not acquire leases")
			expected := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second}[:min(recovery, 3)]
			require.Equal(t, expected, append([]time.Duration{}, delays...))
			require.LessOrEqual(t, sumAPIKeyRouteRetryDelays(delays), 7*time.Second)
			if recovery == 4 {
				require.Nil(t, resolved)
				require.ErrorIs(t, err, ErrAPIKeyRouteTemporarilyUnavailable)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resolved)
				require.NotNil(t, resolved.RouteBreakerLease)
			}
			require.Zero(t, cache.releases)
			require.Nil(t, key.RouteBreakerLease, "selection must not mutate the auth snapshot")
		})
	}
}

func sumAPIKeyRouteRetryDelays(delays []time.Duration) time.Duration {
	var total time.Duration
	for _, delay := range delays {
		total += delay
	}
	return total
}

func TestAPIKeyRouteRetryPermanentEligibilityMatrix(t *testing.T) {
	cases := []struct {
		name   string
		change func(*APIKey)
		model  string
		image  bool
	}{
		{"model_exact", func(*APIKey) {}, "gpt-5.6-extra", false},
		{"scope", func(k *APIKey) { k.Group.RoutingScope = GroupRoutingScopeImage }, "gpt-5.6", false},
		{"image_denied", func(k *APIKey) { k.Group.RoutingScope = GroupRoutingScopeImage }, "gpt-5.6", true},
		{"image_only", func(k *APIKey) { k.MultiGroupRoutes[0].ImageOnly = true }, "gpt-5.6", false},
		{"text_only", func(k *APIKey) {
			k.MultiGroupRoutes[0].TextOnly = true
			k.Group.RoutingScope = GroupRoutingScopeImage
			k.Group.AllowImageGeneration = true
		}, "gpt-5.6", true},
		{"disabled_group", func(k *APIKey) { k.Group.Status = "disabled" }, "gpt-5.6", false},
		{"missing_group", func(k *APIKey) { k.Group = nil }, "gpt-5.6", false},
		{"disabled_route", func(k *APIKey) {
			k.MultiGroupRoutes[0].Enabled = false
			k.Group.ModelMatchPatterns = nil
		}, "gpt-5.6", false},
		{"unavailable_subscription", func(k *APIKey) {
			k.Group.SubscriptionType = SubscriptionTypeSubscription
			k.UnavailableRouteGroupIDs = map[int64]struct{}{1: {}}
		}, "gpt-5.6", false},
		{"pinned_mismatch", func(k *APIKey) { k.PinnedAccountID = 42 }, "claude-sonnet", false},
		{"pinned_platform", func(k *APIKey) {
			k.PinnedAccountID = 42
			k.Group.Platform = PlatformAnthropic
		}, "gpt-5.6", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := apiKeyRouteRetryKey()
			tc.change(key)
			cache := &apiKeyRouteRetryCache{acquire: func(context.Context, APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error) { return nil, nil }}
			svc := NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)
			resolved, err := svc.resolveForModelRequestWithRetry(context.Background(), key, "/v1/responses", "", tc.model, tc.image, func(context.Context, time.Duration) error {
				t.Fatal("permanent mismatch must not wait")
				return nil
			})
			require.NoError(t, err)
			require.Nil(t, resolved)
		})
	}
}

func TestAPIKeyRouteRetryUnrelatedAndExclusiveCandidates(t *testing.T) {
	for _, eligible := range []string{"unrelated", "exclusive_denied", "exclusive_allowed", "subscription"} {
		t.Run(eligible, func(t *testing.T) {
			key := apiKeyRouteRetryKey()
			key.Group.ModelMatchPatterns = []string{"other"}
			candidate := &Group{ID: 2, Status: StatusActive, Hydrated: true, Platform: PlatformOpenAI, ModelMatchPatterns: []string{"gpt-5.6"}, IsExclusive: true}
			key.MultiGroupRoutes = append(key.MultiGroupRoutes, domain.APIKeyMultiGroupRoute{GroupID: 2, Enabled: true})
			key.MultiGroupRouteGroups = []*Group{candidate}
			switch eligible {
			case "unrelated":
				candidate.ModelMatchPatterns = []string{"gpt-5.6-unrelated"}
			case "exclusive_allowed":
				key.User.AllowedGroups = []int64{2}
			case "subscription":
				candidate.SubscriptionType = SubscriptionTypeSubscription
			}
			wantEligible := eligible == "exclusive_allowed" || eligible == "subscription"
			require.Equal(t, wantEligible, hasEligibleAPIKeyModelRoute(key, "/v1/responses", "", "gpt-5.6", false))
			cache := &apiKeyRouteRetryCache{cooling: true}
			svc := NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)
			waits := 0
			_, err := svc.resolveForModelRequestWithRetry(context.Background(), key, "/v1/responses", "", "gpt-5.6", false, func(context.Context, time.Duration) error {
				waits++
				return nil
			})
			if wantEligible {
				require.ErrorIs(t, err, ErrAPIKeyRouteTemporarilyUnavailable)
				require.Equal(t, 3, waits)
			} else {
				require.NoError(t, err)
				require.Zero(t, waits)
			}
			require.Zero(t, cache.acquires)
		})
	}
}

func TestAPIKeyRouteRetryLiveCooldownAndLegacy(t *testing.T) {
	cache := &apiKeyRouteRetryCache{acquire: func(_ context.Context, key APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error) {
		return &APIKeyRouteBreakerLease{Key: key}, nil
	}}
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)
	key := apiKeyRouteRetryKey()
	svc.routeCooldownL1.Store(svc.routeCooldownKey(key.ID, 1), time.Now().Add(time.Hour))
	waits := 0
	resolved, err := svc.resolveForModelRequestWithRetry(context.Background(), key, "/v1/responses", "", "gpt-5.6", false, func(context.Context, time.Duration) error {
		waits++
		svc.routeCooldownL1.Store(svc.routeCooldownKey(key.ID, 1), time.Now().Add(-time.Second))
		return nil
	})
	require.NoError(t, err)
	require.NotNil(t, resolved)
	require.Equal(t, 1, waits)
	require.Equal(t, 1, cache.acquires)
	key.MultiGroupRoutes = nil
	key.Group.ModelMatchPatterns = nil
	resolved, err = svc.ResolveForModelRequestWithRetry(context.Background(), key, "/v1/responses", "anthropic", "unknown-model", true)
	require.NoError(t, err)
	require.Same(t, key, resolved)
	require.Equal(t, 1, cache.acquires)
}

func TestAPIKeyRouteRetryContextTimerCancellation(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		t.Run(map[bool]string{false: "cancel", true: "deadline"}[deadline], func(t *testing.T) {
			var ctx context.Context
			var cancel context.CancelFunc
			if deadline {
				ctx, cancel = context.WithTimeout(context.Background(), 25*time.Millisecond)
			} else {
				ctx, cancel = context.WithCancel(context.Background())
				timer := time.AfterFunc(25*time.Millisecond, cancel)
				defer timer.Stop()
			}
			defer cancel()
			cache := &apiKeyRouteRetryCache{acquire: func(context.Context, APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error) { return nil, nil }}
			svc := NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)
			start := time.Now()
			resolved, err := svc.ResolveForModelRequestWithRetry(ctx, apiKeyRouteRetryKey(), "/v1/responses", "", "gpt-5.6", false)
			require.Nil(t, resolved)
			require.True(t, errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
			require.Less(t, time.Since(start), 500*time.Millisecond)
			require.Equal(t, 1, cache.acquires)
		})
	}
}

func TestAPIKeyRouteRetryCancellationAfterLease(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache := &apiKeyRouteRetryCache{}
	cache.acquire = func(_ context.Context, key APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error) {
		cancel()
		return &APIKeyRouteBreakerLease{Key: key, HalfOpen: true, ProbeToken: 9}, nil
	}
	cache.release = func(cleanup context.Context) {
		require.NoError(t, cleanup.Err())
		until, ok := cleanup.Deadline()
		require.True(t, ok)
		require.WithinDuration(t, time.Now().Add(2*time.Second), until, 100*time.Millisecond)
	}
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)
	resolved, err := svc.ResolveForModelRequestWithRetry(ctx, apiKeyRouteRetryKey(), "/v1/responses", "", "gpt-5.6", false)
	require.Nil(t, resolved)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, cache.acquires)
	require.Equal(t, 1, cache.releases)
}

func TestAPIKeyRouteRetryUnselectedLeaseCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	key := apiKeyRouteRetryKey()
	key.Group.ModelMatchPatterns = []string{"unrelated"}
	key.MultiGroupRouteGroups = []*Group{{ID: 2, Status: StatusActive, Hydrated: true, Platform: PlatformOpenAI, ModelMatchPatterns: []string{"gpt-5.6"}}}
	key.MultiGroupRoutes = append(key.MultiGroupRoutes, domain.APIKeyMultiGroupRoute{GroupID: 2, Enabled: true})
	cache := &apiKeyRouteRetryCache{}
	cache.acquire = func(_ context.Context, breakerKey APIKeyRouteBreakerKey) (*APIKeyRouteBreakerLease, error) {
		cancel()
		return &APIKeyRouteBreakerLease{Key: breakerKey, HalfOpen: true, ProbeToken: 3}, nil
	}
	cache.release = func(cleanup context.Context) {
		require.NoError(t, cleanup.Err())
		_, bounded := cleanup.Deadline()
		require.True(t, bounded)
	}
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, cache, nil)
	resolved, err := svc.ResolveForModelRequestWithRetry(ctx, key, "/v1/responses", "", "gpt-5.6", false)
	require.Nil(t, resolved)
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, cache.acquires, "cancellation must prevent acquiring another group")
	require.Equal(t, 1, cache.releases, "unselected probe must also be released exactly once")
}
