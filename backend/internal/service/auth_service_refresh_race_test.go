//go:build unit

package service_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type refreshRaceCache struct {
	*emailBindRefreshTokenCacheStub
	stores          atomic.Int32
	consumes        atomic.Int32
	deletes         atomic.Int32
	familyDeletes   atomic.Int32
	readOverride    func(context.Context, string) (*service.RefreshTokenData, error)
	consumeOverride func(context.Context, string) (*service.RefreshTokenData, error)
	barrierToken    string
	readBarrier     chan struct{}
	readCount       atomic.Int32
}

func (c *refreshRaceCache) StoreRefreshToken(ctx context.Context, hash string, data *service.RefreshTokenData, ttl time.Duration) error {
	c.stores.Add(1)
	return c.emailBindRefreshTokenCacheStub.StoreRefreshToken(ctx, hash, data, ttl)
}

func (c *refreshRaceCache) GetRefreshToken(ctx context.Context, hash string) (*service.RefreshTokenData, error) {
	if c.readOverride != nil {
		return c.readOverride(ctx, hash)
	}
	data, err := c.emailBindRefreshTokenCacheStub.GetRefreshToken(ctx, hash)
	if c.readBarrier != nil && hash == c.barrierToken {
		if c.readCount.Add(1) == 2 {
			close(c.readBarrier)
		}
		select {
		case <-c.readBarrier:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return data, err
}

func (c *refreshRaceCache) ConsumeRefreshToken(ctx context.Context, hash string) (*service.RefreshTokenData, error) {
	c.consumes.Add(1)
	if c.consumeOverride != nil {
		return c.consumeOverride(ctx, hash)
	}
	return c.emailBindRefreshTokenCacheStub.ConsumeRefreshToken(ctx, hash)
}

func (c *refreshRaceCache) DeleteRefreshToken(ctx context.Context, hash string) error {
	c.deletes.Add(1)
	return c.emailBindRefreshTokenCacheStub.DeleteRefreshToken(ctx, hash)
}

func (c *refreshRaceCache) DeleteTokenFamily(ctx context.Context, family string) error {
	c.familyDeletes.Add(1)
	return c.emailBindRefreshTokenCacheStub.DeleteTokenFamily(ctx, family)
}

// The static interface intentionally hides the concrete cache's consumer method.
type refreshInterfaceOnlyCache struct {
	service.RefreshTokenCache
}

func refreshRaceHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func newRefreshRaceService(cache service.RefreshTokenCache, repo service.UserRepository) *service.AuthService {
	cfg := &config.Config{JWT: config.JWTConfig{
		Secret: "refresh-race-test-secret", ExpireHour: 1,
		AccessTokenExpireMinutes: 60, RefreshTokenExpireDays: 7,
	}}
	return service.NewAuthService(nil, repo, nil, cache, cfg, nil, nil, nil, nil, nil, nil, nil)
}

func newRefreshRaceFixture(t *testing.T) (*service.AuthService, *refreshRaceCache, *emailBindUserRepoStub, *service.User, string) {
	t.Helper()
	user := &service.User{ID: 41, Email: "refresh@example.com", Role: service.RoleUser, Status: service.StatusActive, TokenVersion: 4}
	repo := newEmailBindUserRepoStub(user)
	cache := &refreshRaceCache{emailBindRefreshTokenCacheStub: newEmailBindRefreshTokenCacheStub()}
	svc := newRefreshRaceService(cache, repo)
	pair, err := svc.GenerateTokenPair(context.Background(), user, "")
	require.NoError(t, err)
	return svc, cache, repo, user, pair.RefreshToken
}

func TestAuthServiceRefreshConcurrentSingleWinner(t *testing.T) {
	svc, cache, _, _, token := newRefreshRaceFixture(t)
	cache.barrierToken = refreshRaceHash(token)
	cache.readBarrier = make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	type outcome struct {
		pair *service.TokenPairWithUser
		err  error
	}
	results := make(chan outcome, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pair, err := svc.RefreshTokenPair(ctx, token)
			results <- outcome{pair, err}
		}()
	}
	wg.Wait()
	close(results)
	wins, losses := 0, 0
	for result := range results {
		if result.err == nil {
			wins++
			require.NotNil(t, result.pair)
			require.NotEmpty(t, result.pair.AccessToken)
			require.NotEqual(t, token, result.pair.RefreshToken)
		} else {
			losses++
			require.ErrorIs(t, result.err, service.ErrRefreshTokenInvalid)
			require.Nil(t, result.pair)
		}
	}
	require.Equal(t, 1, wins)
	require.Equal(t, 1, losses)
	require.Equal(t, int32(2), cache.readCount.Load(), "both requests must validate the preconsume value")
	require.Equal(t, int32(2), cache.consumes.Load())
	require.Equal(t, int32(2), cache.stores.Load(), "initial issuance plus exactly one replacement")
	require.Zero(t, cache.deletes.Load(), "rotation must not fall back to separate deletion")
}

func TestAuthServiceRefreshSequentialRotationAndReuse(t *testing.T) {
	svc, cache, _, _, original := newRefreshRaceFixture(t)
	ctx := context.Background()
	firstData, err := cache.GetRefreshToken(ctx, refreshRaceHash(original))
	require.NoError(t, err)
	first, err := svc.RefreshTokenPair(ctx, original)
	require.NoError(t, err)
	require.NotEmpty(t, first.AccessToken)
	second, err := svc.RefreshTokenPair(ctx, first.RefreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, second.AccessToken)
	require.NotEqual(t, original, first.RefreshToken)
	require.NotEqual(t, first.RefreshToken, second.RefreshToken)
	data, err := cache.GetRefreshToken(ctx, refreshRaceHash(second.RefreshToken))
	require.NoError(t, err)
	require.Equal(t, firstData.FamilyID, data.FamilyID)
	require.Equal(t, firstData.UserID, data.UserID)
	require.Equal(t, firstData.TokenVersion, data.TokenVersion)
	for _, used := range []string{original, first.RefreshToken} {
		pair, err := svc.RefreshTokenPair(ctx, used)
		require.Nil(t, pair)
		require.ErrorIs(t, err, service.ErrRefreshTokenInvalid)
	}
	require.Equal(t, int32(3), cache.stores.Load())
	require.Equal(t, int32(2), cache.consumes.Load())
	require.Zero(t, cache.deletes.Load())
}

func TestAuthServiceRefreshConsumeFailuresNeverIssue(t *testing.T) {
	injected := errors.New("injected Redis execution failure")
	for _, tc := range []struct {
		name string
		data bool
		err  error
		want error
	}{
		{"missing", false, service.ErrRefreshTokenNotFound, service.ErrRefreshTokenInvalid},
		{"nil result", false, nil, service.ErrRefreshTokenInvalid},
		{"execution error", false, injected, service.ErrServiceUnavailable},
		{"data with execution error", true, injected, service.ErrServiceUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, cache, _, _, token := newRefreshRaceFixture(t)
			data, err := cache.GetRefreshToken(context.Background(), refreshRaceHash(token))
			require.NoError(t, err)
			cache.consumeOverride = func(context.Context, string) (*service.RefreshTokenData, error) {
				if tc.data {
					return data, tc.err
				}
				return nil, tc.err
			}
			pair, err := svc.RefreshTokenPair(context.Background(), token)
			require.Nil(t, pair)
			require.ErrorIs(t, err, tc.want)
			require.Equal(t, int32(1), cache.stores.Load())
			require.Equal(t, int32(1), cache.consumes.Load())
			require.Zero(t, cache.deletes.Load())
		})
	}
}

func TestAuthServiceRefreshReadFailuresNeverConsumeOrIssue(t *testing.T) {
	injected := errors.New("injected read failure")
	for _, tc := range []struct {
		name string
		data bool
		err  error
		want error
	}{
		{"missing", false, service.ErrRefreshTokenNotFound, service.ErrRefreshTokenInvalid},
		{"nil result", false, nil, service.ErrRefreshTokenInvalid},
		{"execution error", false, injected, service.ErrServiceUnavailable},
		{"data with execution error", true, injected, service.ErrServiceUnavailable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, cache, _, _, token := newRefreshRaceFixture(t)
			data, err := cache.GetRefreshToken(context.Background(), refreshRaceHash(token))
			require.NoError(t, err)
			cache.readOverride = func(context.Context, string) (*service.RefreshTokenData, error) {
				if tc.data {
					return data, tc.err
				}
				return nil, tc.err
			}
			pair, err := svc.RefreshTokenPair(context.Background(), token)
			require.Nil(t, pair)
			require.ErrorIs(t, err, tc.want)
			require.Equal(t, int32(1), cache.stores.Load())
			require.Zero(t, cache.consumes.Load())
			require.Zero(t, cache.deletes.Load())
		})
	}
}

func TestAuthServiceRefreshAbsentConsumerAndNilCache(t *testing.T) {
	_, cache, repo, user, token := newRefreshRaceFixture(t)
	legacy := newRefreshRaceService(&refreshInterfaceOnlyCache{RefreshTokenCache: cache}, repo)
	// Interface-only injected caches still support issuance; rotation fails closed.
	issued, err := legacy.GenerateTokenPair(context.Background(), user, "")
	require.NoError(t, err)
	require.NotEmpty(t, issued.RefreshToken)
	before := cache.stores.Load()
	pair, err := legacy.RefreshTokenPair(context.Background(), token)
	require.Nil(t, pair)
	require.ErrorIs(t, err, service.ErrServiceUnavailable)
	require.Equal(t, before, cache.stores.Load())
	require.Zero(t, cache.consumes.Load())
	require.Zero(t, cache.deletes.Load())
	pair, err = newRefreshRaceService(nil, repo).RefreshTokenPair(context.Background(), token)
	require.Nil(t, pair)
	require.ErrorIs(t, err, service.ErrRefreshTokenInvalid)
}

func TestAuthServiceRefreshEnabledSessionBinding(t *testing.T) {
	for _, tc := range []struct {
		name    string
		binding *service.SessionBinding
		matches bool
	}{
		{name: "matching legitimate refresh", binding: &service.SessionBinding{IP: "192.0.2.10", UserAgent: "refresh-test-agent"}, matches: true},
		{name: "changed IP", binding: &service.SessionBinding{IP: "192.0.2.11", UserAgent: "refresh-test-agent"}},
		{name: "changed user agent", binding: &service.SessionBinding{IP: "192.0.2.10", UserAgent: "different-agent"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{JWT: config.JWTConfig{
				Secret: "refresh-race-test-secret", ExpireHour: 1,
				AccessTokenExpireMinutes: 60, RefreshTokenExpireDays: 7,
			}}
			settings := service.NewSettingService(&emailBindSettingRepoStub{values: map[string]string{
				service.SettingKeySessionBindingEnabled: "true",
			}}, cfg)
			user := &service.User{ID: 41, Email: "refresh@example.com", Role: service.RoleUser, Status: service.StatusActive, TokenVersion: 4}
			repo := newEmailBindUserRepoStub(user)
			cache := &refreshRaceCache{emailBindRefreshTokenCacheStub: newEmailBindRefreshTokenCacheStub()}
			svc := service.NewAuthService(nil, repo, nil, cache, cfg, settings, nil, nil, nil, nil, nil, nil)
			originalBinding := &service.SessionBinding{IP: "192.0.2.10", UserAgent: "refresh-test-agent"}
			issueCtx := service.WithSessionBinding(context.Background(), originalBinding)
			issued, err := svc.GenerateTokenPair(issueCtx, user, "")
			require.NoError(t, err)
			originalHash := refreshRaceHash(issued.RefreshToken)
			originalData, err := cache.GetRefreshToken(issueCtx, originalHash)
			require.NoError(t, err)
			require.NotNil(t, originalData)
			require.NotEmpty(t, originalData.FamilyID)
			require.Equal(t, originalBinding.Hash(), originalData.BindingHash)
			refreshCtx := service.WithSessionBinding(context.Background(), tc.binding)
			pair, err := svc.RefreshTokenPair(refreshCtx, issued.RefreshToken)
			if tc.matches {
				require.NoError(t, err)
				require.NotNil(t, pair)
				require.NotEmpty(t, pair.AccessToken)
				require.NotEqual(t, issued.RefreshToken, pair.RefreshToken)
				replacement, err := cache.GetRefreshToken(refreshCtx, refreshRaceHash(pair.RefreshToken))
				require.NoError(t, err)
				require.NotNil(t, replacement)
				require.Equal(t, originalData.FamilyID, replacement.FamilyID)
				require.Equal(t, originalBinding.Hash(), replacement.BindingHash)
				require.Equal(t, int32(2), cache.stores.Load())
				require.Equal(t, int32(1), cache.consumes.Load())
				require.Zero(t, cache.familyDeletes.Load())
			} else {
				require.ErrorIs(t, err, service.ErrSessionBindingMismatch)
				require.Nil(t, pair)
				require.Equal(t, int32(1), cache.stores.Load())
				require.Zero(t, cache.consumes.Load())
				require.Equal(t, int32(1), cache.familyDeletes.Load())
				hashes, err := cache.GetFamilyTokenHashes(refreshCtx, originalData.FamilyID)
				require.NoError(t, err)
				require.Empty(t, hashes)
				data, err := cache.GetRefreshToken(refreshCtx, originalHash)
				require.ErrorIs(t, err, service.ErrRefreshTokenNotFound)
				require.Nil(t, data)
			}
			require.Zero(t, cache.deletes.Load())
		})
	}
}

func TestAuthServiceRefreshValidationBeforeConsume(t *testing.T) {
	for _, tc := range []struct {
		name          string
		change        func(*refreshRaceCache, *emailBindUserRepoStub, string)
		want          error
		familyDeletes int32
		deletes       int32
	}{
		{
			name: "expired", want: service.ErrRefreshTokenExpired, deletes: 1,
			change: func(c *refreshRaceCache, _ *emailBindUserRepoStub, hash string) {
				c.mu.Lock()
				defer c.mu.Unlock()
				c.tokens[hash].ExpiresAt = time.Now().Add(-time.Minute)
			},
		},
		{
			name: "deleted user", want: service.ErrRefreshTokenInvalid, familyDeletes: 1,
			change: func(_ *refreshRaceCache, repo *emailBindUserRepoStub, _ string) {
				repo.mu.Lock()
				defer repo.mu.Unlock()
				delete(repo.usersByID, 41)
			},
		},
		{
			name: "inactive user", want: service.ErrUserNotActive, familyDeletes: 1,
			change: func(_ *refreshRaceCache, repo *emailBindUserRepoStub, _ string) {
				repo.mu.Lock()
				defer repo.mu.Unlock()
				repo.usersByID[41].Status = "disabled"
			},
		},
		{
			name: "revoked version", want: service.ErrTokenRevoked, familyDeletes: 1,
			change: func(_ *refreshRaceCache, repo *emailBindUserRepoStub, _ string) {
				repo.mu.Lock()
				defer repo.mu.Unlock()
				repo.usersByID[41].TokenVersion++
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, cache, repo, _, token := newRefreshRaceFixture(t)
			tc.change(cache, repo, refreshRaceHash(token))
			pair, err := svc.RefreshTokenPair(context.Background(), token)
			require.Nil(t, pair)
			require.ErrorIs(t, err, tc.want)
			require.Equal(t, int32(1), cache.stores.Load())
			require.Zero(t, cache.consumes.Load())
			require.Equal(t, tc.familyDeletes, cache.familyDeletes.Load())
			require.Equal(t, tc.deletes, cache.deletes.Load())
		})
	}
	t.Run("malformed", func(t *testing.T) {
		svc, cache, _, _, _ := newRefreshRaceFixture(t)
		cache.readOverride = func(context.Context, string) (*service.RefreshTokenData, error) {
			t.Error("malformed token must not reach cache lookup")
			return nil, nil
		}
		for _, token := range []string{"", "not-a-refresh-token"} {
			pair, err := svc.RefreshTokenPair(context.Background(), token)
			require.Nil(t, pair)
			require.ErrorIs(t, err, service.ErrRefreshTokenInvalid)
		}
		require.Zero(t, cache.consumes.Load())
		require.Equal(t, int32(1), cache.stores.Load())
	})
	t.Run("binding disabled retains existing exception", func(t *testing.T) {
		svc, cache, _, _, token := newRefreshRaceFixture(t)
		cache.mu.Lock()
		cache.tokens[refreshRaceHash(token)].BindingHash = "stored-network-binding"
		cache.mu.Unlock()
		pair, err := svc.RefreshTokenPair(context.Background(), token)
		require.NoError(t, err)
		require.NotNil(t, pair)
		require.Equal(t, int32(1), cache.consumes.Load())
		require.Zero(t, cache.familyDeletes.Load())
	})
}
