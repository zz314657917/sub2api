package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type apiKeyRouteRetryMiddlewareCache struct {
	service.APIKeyCache
	acquires, releases, successes, failures, cooldownSets, cooldownDeletes int
	acquire                                                                func(context.Context, service.APIKeyRouteBreakerKey) (*service.APIKeyRouteBreakerLease, error)
	cooling                                                                func(int64) bool
	release                                                                func(context.Context)
}

func (s *apiKeyRouteRetryMiddlewareCache) GetAuthCache(context.Context, string) (*service.APIKeyAuthCacheEntry, error) {
	return nil, nil
}

func (s *apiKeyRouteRetryMiddlewareCache) SetAuthCache(context.Context, string, *service.APIKeyAuthCacheEntry, time.Duration) error {
	return nil
}

func (s *apiKeyRouteRetryMiddlewareCache) IsRouteGroupCooling(_ context.Context, _, groupID int64) (bool, error) {
	return s.cooling != nil && s.cooling(groupID), nil
}

func (s *apiKeyRouteRetryMiddlewareCache) AcquireAPIKeyRouteBreaker(ctx context.Context, key service.APIKeyRouteBreakerKey) (*service.APIKeyRouteBreakerLease, error) {
	s.acquires++
	if s.acquire != nil {
		return s.acquire(ctx, key)
	}
	return &service.APIKeyRouteBreakerLease{Key: key, HalfOpen: true, Generation: 1, ProbeToken: int64(s.acquires)}, nil
}

func (s *apiKeyRouteRetryMiddlewareCache) SetRouteGroupCooldown(context.Context, int64, int64, time.Duration) error {
	s.cooldownSets++
	return nil
}

func (s *apiKeyRouteRetryMiddlewareCache) DeleteRouteGroupCooldown(context.Context, int64, int64) error {
	s.cooldownDeletes++
	return nil
}

func (s *apiKeyRouteRetryMiddlewareCache) RecordAPIKeyRouteBreakerSuccess(context.Context, service.APIKeyRouteBreakerLease) error {
	s.successes++
	return nil
}

func (s *apiKeyRouteRetryMiddlewareCache) RecordAPIKeyRouteBreakerFailure(context.Context, service.APIKeyRouteBreakerLease) error {
	s.failures++
	return nil
}

func (s *apiKeyRouteRetryMiddlewareCache) ReleaseAPIKeyRouteBreakerProbe(ctx context.Context, _ service.APIKeyRouteBreakerLease) error {
	s.releases++
	if s.release != nil {
		s.release(ctx)
	}
	return nil
}

func apiKeyRouteRetryMiddlewareKey() *service.APIKey {
	id := int64(1)
	return &service.APIKey{
		ID: 81, Key: "retry-middleware", UserID: 3, Status: service.StatusActive,
		User:    &service.User{ID: 3, Status: service.StatusActive, Balance: 10},
		GroupID: &id,
		Group: &service.Group{ID: id, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI,
			ModelMatchPatterns: []string{"other"}},
		MultiGroupRoutes: []domain.APIKeyMultiGroupRoute{
			{GroupID: 1, Enabled: true, Priority: 1},
			{GroupID: 2, Enabled: true, Priority: 2},
		},
		MultiGroupRouteGroups: []*service.Group{{
			ID: 2, Status: service.StatusActive, Hydrated: true, Platform: service.PlatformOpenAI,
			ModelMatchPatterns: []string{"gpt-5.6"},
		}},
		AccountPoolStrategy: service.AccountPoolStrategySharedOnly,
	}
}

func apiKeyRouteRetryMiddlewareService(key *service.APIKey, cache *apiKeyRouteRetryMiddlewareCache, cfg *config.Config) *service.APIKeyService {
	repo := fakeAPIKeyRepo{
		getByKey: func(context.Context, string) (*service.APIKey, error) {
			clone := *key
			return &clone, nil
		},
		updateLastUsed: func(context.Context, int64, time.Time) error { return nil },
	}
	return service.NewAPIKeyService(repo, nil, nil, nil, nil, cache, cfg)
}

func apiKeyRouteRetryAssertNoHealthWrites(t *testing.T, cache *apiKeyRouteRetryMiddlewareCache) {
	t.Helper()
	require.Zero(t, cache.successes)
	require.Zero(t, cache.failures)
	require.Zero(t, cache.cooldownSets)
	require.Zero(t, cache.cooldownDeletes)
}

func TestAPIKeyRouteRetryAuthFinalizerTerminalOutcomes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, mode := range []string{config.RunModeSimple, config.RunModeStandard} {
		for _, outcome := range []string{"exhausted", "permanent", "cancelled", "cancelled_unwritten", "committed", "aborted"} {
			t.Run(mode+"/"+outcome, func(t *testing.T) {
				key := apiKeyRouteRetryMiddlewareKey()
				cache := &apiKeyRouteRetryMiddlewareCache{}
				cache.acquire = func(_ context.Context, breakerKey service.APIKeyRouteBreakerKey) (*service.APIKeyRouteBreakerLease, error) {
					if breakerKey.GroupID == 2 {
						return nil, nil
					}
					return &service.APIKeyRouteBreakerLease{Key: breakerKey}, nil
				}
				if outcome == "permanent" {
					key.MultiGroupRouteGroups[0].ModelMatchPatterns = []string{"other"}
				}
				cfg := &config.Config{RunMode: mode}
				svc := apiKeyRouteRetryMiddlewareService(key, cache, cfg)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				calls := 0
				router := gin.New()
				router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(svc, nil, cfg)))
				router.POST("/v1/responses", func(c *gin.Context) {
					switch outcome {
					case "cancelled":
						cancel()
					case "cancelled_unwritten":
						cancel()
						c.Abort()
					case "committed":
						c.String(http.StatusAccepted, "already written")
					case "aborted":
						c.Abort()
					}
					authKey, ok := GetAPIKeyFromContext(c)
					require.True(t, ok)
					_, ok = ResolveAPIKeyForModelRequest(c, svc, authKey, "gpt-5.6", false)
					require.False(t, ok)
					_, memoized := c.Get(contextKeyAPIKeyModelResolution)
					require.False(t, memoized, "local failures cannot be memoized as success")
					_, ok = ResolveAPIKeyForModelRequest(c, svc, authKey, "gpt-5.6", false)
					require.False(t, ok, "already terminated requests cannot resume")
				}, func(c *gin.Context) { calls++; c.Status(http.StatusOK) })
				req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(`{"model":"gpt-5.6"}`)).WithContext(ctx)
				req.Header.Set("Authorization", "Bearer "+key.Key)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				require.Zero(t, calls)
				apiKeyRouteRetryAssertNoHealthWrites(t, cache)
				require.Zero(t, cache.releases)
				switch outcome {
				case "exhausted":
					require.Equal(t, http.StatusServiceUnavailable, w.Code)
					require.Contains(t, w.Body.String(), "GROUP_ROUTE_TEMPORARILY_UNAVAILABLE")
					require.Equal(t, "5", w.Header().Get("Retry-After"))
					require.Equal(t, 8, cache.acquires, "four decisions, one acquire per group per decision")
				case "permanent":
					require.Equal(t, http.StatusForbidden, w.Code)
					require.Contains(t, w.Body.String(), "NO_MATCHING_GROUP_ROUTE")
					require.Empty(t, w.Header().Get("Retry-After"))
					require.Equal(t, 2, cache.acquires)
				case "cancelled":
					require.Equal(t, http.StatusRequestTimeout, w.Code)
					require.Contains(t, w.Body.String(), "REQUEST_TIMEOUT")
					require.Zero(t, cache.acquires)
				case "committed":
					require.Equal(t, http.StatusAccepted, w.Code)
					require.Equal(t, "already written", w.Body.String())
					require.Zero(t, cache.acquires)
				default:
					require.Empty(t, w.Body.String())
					require.Zero(t, cache.acquires)
				}
			})
		}
	}
}

func TestAPIKeyRouteRetryCancelledLeaseAndValidationRelease(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, outcome := range []string{"cancel_after_lease", "cancel_unwritten_after_lease", "exclusive_denied", "billing_denied"} {
		t.Run(outcome, func(t *testing.T) {
			key := apiKeyRouteRetryMiddlewareKey()
			if outcome == "exclusive_denied" {
				key.MultiGroupRouteGroups[0].IsExclusive = true
			}
			if outcome == "billing_denied" {
				key.User.Balance = 0
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var activeContext *gin.Context
			cache := &apiKeyRouteRetryMiddlewareCache{}
			cache.acquire = func(_ context.Context, breakerKey service.APIKeyRouteBreakerKey) (*service.APIKeyRouteBreakerLease, error) {
				lease := &service.APIKeyRouteBreakerLease{Key: breakerKey}
				if breakerKey.GroupID == 2 {
					lease.HalfOpen = true
					lease.ProbeToken = 9
					if strings.HasPrefix(outcome, "cancel") {
						cancel()
						if outcome == "cancel_unwritten_after_lease" {
							activeContext.Abort()
						}
					}
				}
				return lease, nil
			}
			cache.release = func(cleanup context.Context) {
				require.NoError(t, cleanup.Err())
				deadline, ok := cleanup.Deadline()
				require.True(t, ok)
				require.WithinDuration(t, time.Now().Add(2*time.Second), deadline, 100*time.Millisecond)
			}
			cfg := &config.Config{RunMode: config.RunModeStandard}
			svc := apiKeyRouteRetryMiddlewareService(key, cache, cfg)
			router := gin.New()
			router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(svc, nil, cfg)))
			calls := 0
			router.POST("/v1/responses", func(c *gin.Context) {
				activeContext = c
				authKey, _ := GetAPIKeyFromContext(c)
				_, ok := ResolveAPIKeyForModelRequest(c, svc, authKey, "gpt-5.6", false)
				require.False(t, ok)
				_, exists := c.Get(contextKeyAPIKeyModelResolution)
				require.False(t, exists)
				if outcome == "exclusive_denied" {
					current, _ := GetAPIKeyFromContext(c)
					require.Equal(t, int64(1), current.Group.ID, "authorization precedes context and billing")
					require.NotEqual(t, int64(2), c.Request.Context().Value(ctxkey.Group).(*service.Group).ID)
				}
			}, func(c *gin.Context) { calls++; c.Status(http.StatusOK) })
			req := httptest.NewRequest(http.MethodPost, "/v1/responses", nil).WithContext(ctx)
			req.Header.Set("Authorization", "Bearer "+key.Key)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Zero(t, calls)
			apiKeyRouteRetryAssertNoHealthWrites(t, cache)
			require.Equal(t, 1, cache.releases)
			require.Equal(t, 2, cache.acquires)
			switch outcome {
			case "exclusive_denied":
				require.Equal(t, http.StatusForbidden, w.Code)
				require.Contains(t, w.Body.String(), "GROUP_NOT_ALLOWED")
			case "billing_denied":
				require.Equal(t, http.StatusForbidden, w.Code)
				require.Contains(t, w.Body.String(), "INSUFFICIENT_BALANCE")
			case "cancel_unwritten_after_lease":
				require.Empty(t, w.Body.String())
			default:
				require.Equal(t, http.StatusRequestTimeout, w.Code)
			}
		})
	}
}

func TestAPIKeyRouteRetryAuthFinalizerCancellationDuringWait(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key := apiKeyRouteRetryMiddlewareKey()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache := &apiKeyRouteRetryMiddlewareCache{}
	cache.acquire = func(_ context.Context, breakerKey service.APIKeyRouteBreakerKey) (*service.APIKeyRouteBreakerLease, error) {
		if breakerKey.GroupID == 2 {
			time.AfterFunc(25*time.Millisecond, cancel)
			return nil, nil
		}
		return &service.APIKeyRouteBreakerLease{Key: breakerKey}, nil
	}
	cfg := &config.Config{RunMode: config.RunModeStandard}
	svc := apiKeyRouteRetryMiddlewareService(key, cache, cfg)
	router := gin.New()
	router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(svc, nil, cfg)))
	calls := 0
	router.POST("/v1/responses", func(c *gin.Context) {
		authKey, _ := GetAPIKeyFromContext(c)
		resolved, ok := ResolveAPIKeyForModelRequest(c, svc, authKey, "gpt-5.6", false)
		require.Nil(t, resolved)
		require.False(t, ok)
		_, memoized := c.Get(contextKeyAPIKeyModelResolution)
		require.False(t, memoized)
	}, func(c *gin.Context) { calls++; c.Status(http.StatusOK) })
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", nil).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+key.Key)
	w := httptest.NewRecorder()
	start := time.Now()
	router.ServeHTTP(w, req)
	require.Less(t, time.Since(start), 500*time.Millisecond)
	require.Equal(t, http.StatusRequestTimeout, w.Code)
	require.Contains(t, w.Body.String(), "REQUEST_TIMEOUT")
	require.Zero(t, calls)
	require.Equal(t, 2, cache.acquires, "cancelled wait must not begin another attempt")
	apiKeyRouteRetryAssertNoHealthWrites(t, cache)
}

func TestAPIKeyRouteRetryStoppedMemoizedSelectionReleasesOnce(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, stopped := range []string{"cancelled", "aborted", "committed"} {
		t.Run(stopped, func(t *testing.T) {
			key := apiKeyRouteRetryMiddlewareKey()
			cache := &apiKeyRouteRetryMiddlewareCache{}
			cache.acquire = func(_ context.Context, breakerKey service.APIKeyRouteBreakerKey) (*service.APIKeyRouteBreakerLease, error) {
				return &service.APIKeyRouteBreakerLease{Key: breakerKey, HalfOpen: breakerKey.GroupID == 2, ProbeToken: 8}, nil
			}
			cache.release = func(cleanup context.Context) {
				require.NoError(t, cleanup.Err())
				_, bounded := cleanup.Deadline()
				require.True(t, bounded)
			}
			cfg := &config.Config{RunMode: config.RunModeStandard}
			svc := apiKeyRouteRetryMiddlewareService(key, cache, cfg)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			router := gin.New()
			router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(svc, nil, cfg)))
			calls := 0
			router.POST("/v1/responses", func(c *gin.Context) {
				authKey, _ := GetAPIKeyFromContext(c)
				first, ok := ResolveAPIKeyForModelRequest(c, svc, authKey, "gpt-5.6", false)
				require.True(t, ok)
				require.NotNil(t, first.RouteBreakerLease)
				switch stopped {
				case "cancelled":
					cancel()
				case "aborted":
					c.Abort()
				case "committed":
					c.String(http.StatusAccepted, "already written")
				}
				for i := 0; i < 3; i++ {
					// The route wrapper may still hold the original auth snapshot,
					// while the protocol handler carries the selected snapshot.
					input := authKey
					if i == 1 {
						input = first
					}
					resolved, ok := ResolveAPIKeyForModelRequest(c, svc, input, "gpt-5.6", false)
					require.False(t, ok)
					require.Nil(t, resolved)
				}
			}, func(c *gin.Context) { calls++; c.Status(http.StatusOK) })
			req := httptest.NewRequest(http.MethodPost, "/v1/responses", nil).WithContext(ctx)
			req.Header.Set("Authorization", "Bearer "+key.Key)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Zero(t, calls)
			apiKeyRouteRetryAssertNoHealthWrites(t, cache)
			require.Equal(t, 2, cache.acquires)
			require.Equal(t, 1, cache.releases)
			if stopped == "committed" {
				require.Equal(t, "already written", w.Body.String())
				require.Equal(t, http.StatusAccepted, w.Code)
			}
		})
	}
}

func TestAPIKeyRouteRetryExclusiveAuthorizationMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, outcome := range []string{"cooled_denied", "allowed", "subscription", "recovered_denied"} {
		t.Run(outcome, func(t *testing.T) {
			key := apiKeyRouteRetryMiddlewareKey()
			group := key.MultiGroupRouteGroups[0]
			group.IsExclusive = true
			switch outcome {
			case "allowed":
				key.User.AllowedGroups = []int64{2}
			case "subscription":
				group.SubscriptionType = service.SubscriptionTypeSubscription
			case "recovered_denied":
				// Eligibility can be established only by another authorized route.
				public := *group
				public.ID, public.IsExclusive = 3, false
				key.MultiGroupRouteGroups = append(key.MultiGroupRouteGroups, &public)
				key.MultiGroupRoutes = append(key.MultiGroupRoutes, domain.APIKeyMultiGroupRoute{GroupID: 3, Enabled: true, Priority: 3})
			}
			cache := &apiKeyRouteRetryMiddlewareCache{}
			selectedAttempts := 0
			cache.cooling = func(id int64) bool { return outcome == "cooled_denied" && id == 2 }
			cache.acquire = func(_ context.Context, breakerKey service.APIKeyRouteBreakerKey) (*service.APIKeyRouteBreakerLease, error) {
				if breakerKey.GroupID == 3 {
					return nil, nil
				}
				if breakerKey.GroupID == 2 {
					selectedAttempts++
					if outcome == "recovered_denied" && selectedAttempts == 1 {
						return nil, nil
					}
				}
				return &service.APIKeyRouteBreakerLease{Key: breakerKey, HalfOpen: breakerKey.GroupID == 2, ProbeToken: 9}, nil
			}
			cfg := &config.Config{RunMode: config.RunModeSimple}
			svc := apiKeyRouteRetryMiddlewareService(key, cache, cfg)
			router := gin.New()
			router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(svc, nil, cfg)))
			calls := 0
			router.POST("/v1/responses", func(c *gin.Context) {
				authKey, _ := GetAPIKeyFromContext(c)
				if _, ok := ResolveAPIKeyForModelRequest(c, svc, authKey, "gpt-5.6", false); !ok {
					return
				}
			}, func(c *gin.Context) { calls++; c.Status(http.StatusOK) })
			req := httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
			req.Header.Set("Authorization", "Bearer "+key.Key)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if outcome == "allowed" || outcome == "subscription" {
				require.Equal(t, 1, calls)
				require.Equal(t, http.StatusOK, w.Code)
				require.Equal(t, 1, cache.successes)
			} else {
				require.Zero(t, calls)
				require.Equal(t, http.StatusForbidden, w.Code)
				apiKeyRouteRetryAssertNoHealthWrites(t, cache)
				if outcome == "cooled_denied" {
					require.Contains(t, w.Body.String(), "NO_MATCHING_GROUP_ROUTE")
					require.Equal(t, 1, cache.acquires)
				} else {
					require.Contains(t, w.Body.String(), "GROUP_NOT_ALLOWED")
					require.Equal(t, 2, selectedAttempts)
					require.Equal(t, 1, cache.releases)
				}
			}
		})
	}
}

func TestAPIKeyRouteRetryLocalhostHTTPRecoveredAndFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, outcome := range []string{"recovered", "permanent", "terminal"} {
		t.Run(outcome, func(t *testing.T) {
			key := apiKeyRouteRetryMiddlewareKey()
			cache := &apiKeyRouteRetryMiddlewareCache{}
			selections := 0
			cache.acquire = func(_ context.Context, breakerKey service.APIKeyRouteBreakerKey) (*service.APIKeyRouteBreakerLease, error) {
				if breakerKey.GroupID == 2 {
					selections++
					if selections == 1 || outcome == "terminal" {
						return nil, nil
					}
				}
				return &service.APIKeyRouteBreakerLease{Key: breakerKey, HalfOpen: breakerKey.GroupID == 2, ProbeToken: 7}, nil
			}
			if outcome == "permanent" {
				key.MultiGroupRouteGroups[0].ModelMatchPatterns = []string{"other"}
			}
			cfg := &config.Config{RunMode: config.RunModeStandard}
			svc := apiKeyRouteRetryMiddlewareService(key, cache, cfg)
			var calls atomic.Int32
			router := gin.New()
			router.Use(gin.HandlerFunc(NewAPIKeyAuthMiddleware(svc, nil, cfg)))
			router.POST("/v1/responses", func(c *gin.Context) {
				authKey, _ := GetAPIKeyFromContext(c)
				first, ok := ResolveAPIKeyForModelRequest(c, svc, authKey, "gpt-5.6", false)
				if !ok {
					return
				}
				second, ok := ResolveAPIKeyForModelRequest(c, svc, first, "gpt-5.6", false)
				require.True(t, ok)
				require.Same(t, first, second)
				require.Equal(t, int64(2), c.Request.Context().Value(ctxkey.Group).(*service.Group).ID)
				require.Equal(t, service.AccountPoolStrategySharedOnly, c.Request.Context().Value(ctxkey.APIKeyAccountPoolStrategy))
				body, err := io.ReadAll(c.Request.Body)
				require.NoError(t, err)
				require.JSONEq(t, `{"model":"gpt-5.6"}`, string(body), "selection must not reread the body")
			}, func(c *gin.Context) {
				calls.Add(1)
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})
			server := httptest.NewServer(router)
			defer server.Close()
			req, err := http.NewRequest(http.MethodPost, server.URL+"/v1/responses", strings.NewReader(`{"model":"gpt-5.6"}`))
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+key.Key)
			client := &http.Client{Timeout: 12 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			switch outcome {
			case "recovered":
				require.Equal(t, http.StatusOK, resp.StatusCode)
				require.Equal(t, int32(1), calls.Load())
				require.Equal(t, 2, selections, "memoization must avoid an extra lease")
				require.Equal(t, 1, cache.successes)
				require.Zero(t, cache.failures)
			case "permanent":
				require.Equal(t, http.StatusForbidden, resp.StatusCode)
				require.Contains(t, string(body), "NO_MATCHING_GROUP_ROUTE")
				require.Zero(t, calls.Load())
				apiKeyRouteRetryAssertNoHealthWrites(t, cache)
			default:
				require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
				require.Equal(t, "5", resp.Header.Get("Retry-After"))
				require.Contains(t, string(body), "GROUP_ROUTE_TEMPORARILY_UNAVAILABLE")
				require.Equal(t, 4, selections)
				require.Zero(t, calls.Load())
				apiKeyRouteRetryAssertNoHealthWrites(t, cache)
			}
		})
	}
}
