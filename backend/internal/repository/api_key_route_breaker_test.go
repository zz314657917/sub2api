package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newAPIKeyRouteBreakerTestCache(t *testing.T) (*apiKeyCache, *miniredis.Miniredis, *time.Time) {
	t.Helper()
	mini := miniredis.RunT(t)
	now := time.Unix(1_700_000_000, 0)
	mini.SetTime(now)
	t.Cleanup(mini.Close)
	return &apiKeyCache{rdb: redis.NewClient(&redis.Options{Addr: mini.Addr()})}, mini, &now
}

func advanceAPIKeyRouteBreakerTime(mini *miniredis.Miniredis, now *time.Time, duration time.Duration) {
	*now = now.Add(duration)
	mini.SetTime(*now)
}

func TestAPIKeyRouteBreakerThresholdBackoffAndHalfOpenProbe(t *testing.T) {
	ctx := context.Background()
	cache, mini, now := newAPIKeyRouteBreakerTestCache(t)
	key, ok := service.NewAPIKeyRouteBreakerKey(11, service.GroupRoutingScopeInference, " GPT-5.6 ")
	require.True(t, ok)

	for range 3 {
		lease, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
		require.NoError(t, err)
		require.NotNil(t, lease)
		require.NoError(t, cache.RecordAPIKeyRouteBreakerFailure(ctx, *lease))
	}

	blocked, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.Nil(t, blocked)
	advanceAPIKeyRouteBreakerTime(mini, now, 29*time.Second)
	blocked, err = cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.Nil(t, blocked)
	advanceAPIKeyRouteBreakerTime(mini, now, time.Second)

	probe, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, probe)
	require.True(t, probe.HalfOpen)
	secondProbe, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.Nil(t, secondProbe)
	require.NoError(t, cache.RecordAPIKeyRouteBreakerFailure(ctx, *probe))

	advanceAPIKeyRouteBreakerTime(mini, now, 119*time.Second)
	blocked, err = cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.Nil(t, blocked)
	advanceAPIKeyRouteBreakerTime(mini, now, time.Second)
	probe, err = cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, probe)
	require.NoError(t, cache.RecordAPIKeyRouteBreakerFailure(ctx, *probe))

	advanceAPIKeyRouteBreakerTime(mini, now, 10*time.Minute-time.Second)
	blocked, err = cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.Nil(t, blocked)
	advanceAPIKeyRouteBreakerTime(mini, now, time.Second)
	probe, err = cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, probe)
	require.NoError(t, cache.RecordAPIKeyRouteBreakerFailure(ctx, *probe))

	advanceAPIKeyRouteBreakerTime(mini, now, 30*time.Minute-time.Second)
	blocked, err = cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.Nil(t, blocked)
	advanceAPIKeyRouteBreakerTime(mini, now, time.Second)
	probe, err = cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, probe)
	require.NoError(t, cache.RecordAPIKeyRouteBreakerSuccess(ctx, *probe))

	closed, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, closed)
	require.False(t, closed.HalfOpen)
	mini.FastForward(30 * time.Minute)
	require.False(t, mini.Exists(apiKeyRouteBreakerKey(key)))
}

func TestAPIKeyRouteBreakerStaleGenerationAndProbeRelease(t *testing.T) {
	ctx := context.Background()
	cache, mini, now := newAPIKeyRouteBreakerTestCache(t)
	key, ok := service.NewAPIKeyRouteBreakerKey(12, service.GroupRoutingScopeInference, "gpt-5.6")
	require.True(t, ok)

	first, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	stale := *first
	require.NoError(t, cache.RecordAPIKeyRouteBreakerFailure(ctx, *first))
	second, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, second)
	require.NoError(t, cache.RecordAPIKeyRouteBreakerSuccess(ctx, *second))
	require.NoError(t, cache.RecordAPIKeyRouteBreakerFailure(ctx, stale))
	for range 3 {
		lease, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
		require.NoError(t, err)
		require.NotNil(t, lease)
		require.NoError(t, cache.RecordAPIKeyRouteBreakerFailure(ctx, *lease))
	}
	advanceAPIKeyRouteBreakerTime(mini, now, 30*time.Second)
	probe, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.True(t, probe.HalfOpen)
	require.NoError(t, cache.ReleaseAPIKeyRouteBreakerProbe(ctx, *probe))
	replacement, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.True(t, replacement.HalfOpen)
	require.NotEqual(t, probe.ProbeToken, replacement.ProbeToken)
	require.NoError(t, cache.RecordAPIKeyRouteBreakerSuccess(ctx, *probe))
	require.NoError(t, cache.RecordAPIKeyRouteBreakerSuccess(ctx, *replacement))
	closed, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, closed)
}

func TestAPIKeyRouteBreakerKeyIsModelAndScopeIsolated(t *testing.T) {
	text, ok := service.NewAPIKeyRouteBreakerKey(19, service.GroupRoutingScopeInference, " GPT-5.6 ")
	require.True(t, ok)
	textNormalized, ok := service.NewAPIKeyRouteBreakerKey(19, service.GroupRoutingScopeInference, "gpt-5.6")
	require.True(t, ok)
	image, ok := service.NewAPIKeyRouteBreakerKey(19, service.GroupRoutingScopeImage, "gpt-5.6")
	require.True(t, ok)
	differentModel, ok := service.NewAPIKeyRouteBreakerKey(19, service.GroupRoutingScopeInference, "gpt-image-2")
	require.True(t, ok)

	require.Equal(t, text, textNormalized)
	require.NotEqual(t, text, image)
	require.NotEqual(t, text, differentModel)
	require.Len(t, text.ModelDigest, 64)
}

func TestAPIKeyRouteBreakerHealthyAcquireDoesNotCreateRedisState(t *testing.T) {
	cache, mini, _ := newAPIKeyRouteBreakerTestCache(t)
	key, ok := service.NewAPIKeyRouteBreakerKey(21, service.GroupRoutingScopeInference, "gpt-5.6")
	require.True(t, ok)
	lease, err := cache.AcquireAPIKeyRouteBreaker(context.Background(), key)
	require.NoError(t, err)
	require.NotNil(t, lease)
	require.Zero(t, lease.Generation)
	require.False(t, mini.Exists(apiKeyRouteBreakerKey(key)))
	require.NoError(t, cache.RecordAPIKeyRouteBreakerSuccess(context.Background(), *lease))
	require.False(t, mini.Exists(apiKeyRouteBreakerKey(key)))
}

func TestAPIKeyRouteBreakerConcurrentFailuresOpenOnce(t *testing.T) {
	ctx := context.Background()
	cache, _, _ := newAPIKeyRouteBreakerTestCache(t)
	key, ok := service.NewAPIKeyRouteBreakerKey(22, service.GroupRoutingScopeInference, "gpt-5.6")
	require.True(t, ok)
	leases := make([]*service.APIKeyRouteBreakerLease, 4)
	for index := range leases {
		lease, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
		require.NoError(t, err)
		require.NotNil(t, lease)
		leases[index] = lease
	}
	var wg sync.WaitGroup
	errs := make(chan error, len(leases))
	for _, lease := range leases {
		wg.Add(1)
		go func(lease *service.APIKeyRouteBreakerLease) {
			defer wg.Done()
			errs <- cache.RecordAPIKeyRouteBreakerFailure(ctx, *lease)
		}(lease)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	blocked, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.Nil(t, blocked)
}

func TestAPIKeyRouteBreakerRetentionSurvivesThirtyOneMinutes(t *testing.T) {
	ctx := context.Background()
	cache, mini, now := newAPIKeyRouteBreakerTestCache(t)
	key, ok := service.NewAPIKeyRouteBreakerKey(23, service.GroupRoutingScopeInference, "gpt-5.6")
	require.True(t, ok)
	for range 3 {
		lease, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
		require.NoError(t, err)
		require.NoError(t, cache.RecordAPIKeyRouteBreakerFailure(ctx, *lease))
	}
	advanceAPIKeyRouteBreakerTime(mini, now, 31*time.Minute)
	probe, err := cache.AcquireAPIKeyRouteBreaker(ctx, key)
	require.NoError(t, err)
	require.NotNil(t, probe)
	require.True(t, probe.HalfOpen)
}
