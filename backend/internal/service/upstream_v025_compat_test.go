package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestUpstreamV025RedisRuntime(t *testing.T) {
	address := os.Getenv("UPSTREAM_V025_REDIS_ADDR")
	if address == "" {
		t.Skip("UPSTREAM_V025_REDIS_ADDR is not set")
	}
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{Addr: address, DB: 0})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	require.NoError(t, client.Ping(ctx).Err())

	cache := repository.NewGeminiTokenCache(client)
	invalidator := service.NewCompositeTokenCacheInvalidator(cache)
	first := &service.Account{ID: 902501, Platform: service.PlatformAntigravity, Type: service.AccountTypeOAuth, Credentials: map[string]any{"project_id": "upstream-v025-shared-project"}}
	second := &service.Account{ID: 902502, Platform: service.PlatformAntigravity, Type: service.AccountTypeOAuth, Credentials: map[string]any{"project_id": "upstream-v025-shared-project"}}
	legacyKey := "ag:upstream-v025-shared-project"
	noProject := &service.Account{ID: 902503, Platform: service.PlatformAntigravity, Type: service.AccountTypeOAuth}
	keys := []string{service.AntigravityTokenCacheKey(first), service.AntigravityTokenCacheKey(second), legacyKey, service.AntigravityTokenCacheKey(noProject)}
	for _, key := range keys {
		_ = cache.DeleteAccessToken(ctx, key)
	}

	require.NotEqual(t, service.AntigravityTokenCacheKey(first), service.AntigravityTokenCacheKey(second))
	require.NoError(t, cache.SetAccessToken(ctx, service.AntigravityTokenCacheKey(first), "first", time.Minute))
	require.NoError(t, cache.SetAccessToken(ctx, service.AntigravityTokenCacheKey(second), "second", time.Minute))
	require.NoError(t, cache.SetAccessToken(ctx, legacyKey, "legacy", time.Minute))
	got, err := cache.GetAccessToken(ctx, service.AntigravityTokenCacheKey(second))
	require.NoError(t, err)
	require.Equal(t, "second", got)

	require.NoError(t, invalidator.InvalidateToken(ctx, first))
	_, err = cache.GetAccessToken(ctx, service.AntigravityTokenCacheKey(first))
	require.ErrorIs(t, err, redis.Nil)
	_, err = cache.GetAccessToken(ctx, legacyKey)
	require.ErrorIs(t, err, redis.Nil)
	got, err = cache.GetAccessToken(ctx, service.AntigravityTokenCacheKey(second))
	require.NoError(t, err)
	require.Equal(t, "second", got)

	require.NoError(t, cache.SetAccessToken(ctx, service.AntigravityTokenCacheKey(noProject), "no-project", time.Minute))
	require.NoError(t, invalidator.InvalidateToken(ctx, noProject))
	_, err = cache.GetAccessToken(ctx, service.AntigravityTokenCacheKey(noProject))
	require.ErrorIs(t, err, redis.Nil)
}
