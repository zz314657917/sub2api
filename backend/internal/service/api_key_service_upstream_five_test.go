package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type upstreamFiveAuthRepo struct {
	APIKeyRepository
	calls int
}

func (r *upstreamFiveAuthRepo) GetByKeyForAuth(context.Context, string) (*APIKey, error) {
	r.calls++
	return nil, ErrAPIKeyNotFound
}

type upstreamFiveAuthCache struct {
	APIKeyCache
	calls int
}

func (c *upstreamFiveAuthCache) GetAuthCache(context.Context, string) (*APIKeyAuthCacheEntry, error) {
	c.calls++
	return nil, nil
}

func TestUpstreamFiveInvalidUTF8BeforeAuthCacheAndRepository(t *testing.T) {
	cfg := &config.Config{}
	cfg.APIKeyAuth.L2TTLSeconds = 60
	repo := &upstreamFiveAuthRepo{}
	cache := &upstreamFiveAuthCache{}
	svc := NewAPIKeyService(repo, nil, nil, nil, nil, cache, cfg)
	for _, key := range []string{"sk-\xff\xfe", "sk-\xc0\xaf", "sk-\xed\xa0\x80", "sk-\xe2\x82"} {
		keyResult, err := svc.GetByKey(context.Background(), key)
		require.Nil(t, keyResult)
		require.ErrorIs(t, err, ErrAPIKeyNotFound)
	}
	require.Zero(t, cache.calls)
	require.Zero(t, repo.calls)

	// A populated L1 entry must not turn an invalid credential into a cache hit.
	cfg.APIKeyAuth.L1Size = 1024
	cfg.APIKeyAuth.L1TTLSeconds = 60
	svc = NewAPIKeyService(repo, nil, nil, nil, nil, cache, cfg)
	require.NotNil(t, svc.authCacheL1)
	t.Cleanup(svc.authCacheL1.Close)
	entry := &APIKeyAuthCacheEntry{Snapshot: &APIKeyAuthSnapshot{Version: apiKeyAuthSnapshotVersion, APIKeyID: 42}}
	svc.authCacheL1.Set(svc.authCacheKey("sk-\xff"), entry, 1)
	svc.authCacheL1.Wait()
	_, hit := svc.authCacheL1.Get(svc.authCacheKey("sk-\xff"))
	require.True(t, hit)
	result, err := svc.GetByKey(context.Background(), "sk-\xff")
	require.Nil(t, result)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	require.Zero(t, cache.calls)
	require.Zero(t, repo.calls)
}

func TestUpstreamFiveValidUTF8StillReachesRepository(t *testing.T) {
	repo := &upstreamFiveAuthRepo{}
	svc := NewAPIKeyService(repo, nil, nil, nil, nil, nil, &config.Config{})
	for _, key := range []string{"sk-ascii", "sk-合法UTF8"} {
		_, err := svc.GetByKey(context.Background(), key)
		require.ErrorIs(t, err, ErrAPIKeyNotFound)
	}
	require.Equal(t, 2, repo.calls)
}
