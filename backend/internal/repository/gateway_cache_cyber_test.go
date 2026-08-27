package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestGatewayCacheCyberSessionBlockTTL(t *testing.T) {
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	store, ok := NewGatewayCache(client).(service.CyberSessionBlockStore)
	require.True(t, ok)
	ctx := context.Background()

	require.NoError(t, store.SetCyberSessionBlocked(ctx, "key-a", 30*time.Second))
	blocked, err := store.IsCyberSessionBlocked(ctx, "key-a")
	require.NoError(t, err)
	require.True(t, blocked)

	otherBlocked, err := store.IsCyberSessionBlocked(ctx, "key-b")
	require.NoError(t, err)
	require.False(t, otherBlocked)

	redisServer.FastForward(31 * time.Second)
	blocked, err = store.IsCyberSessionBlocked(ctx, "key-a")
	require.NoError(t, err)
	require.False(t, blocked)
}
