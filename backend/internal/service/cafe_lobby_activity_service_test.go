package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestCafeLobbyActivityRecordsOnlyAnonymousDailyProjection(t *testing.T) {
	mini := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	service := NewCafeLobbyActivityService(client, &config.Config{Timezone: "Asia/Shanghai", JWT: config.JWTConfig{Secret: "test-only-cafe-secret"}})
	service.now = func() time.Time { return time.Date(2026, 8, 3, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60)) }
	service.RecordPersistedUsage(42, time.Date(2026, 8, 3, 3, 59, 0, 0, time.UTC))
	service.RecordPersistedUsage(42, time.Date(2026, 8, 3, 4, 0, 0, 0, time.UTC))
	service.RecordPersistedUsage(43, time.Date(2026, 8, 3, 4, 1, 0, 0, time.UTC))

	key := cafeLobbyUsersKey("2026-08-03")
	requestKey := cafeLobbyRequestsKey("2026-08-03")
	require.Eventually(t, func() bool {
		return client.ZCard(context.Background(), key).Val() == 2 && client.Get(context.Background(), requestKey).Val() == "3"
	}, time.Second, 10*time.Millisecond)

	activity := service.Snapshot(context.Background())
	require.True(t, activity.Available)
	require.Equal(t, "2026-08-03", activity.Date)
	require.Equal(t, int64(2), activity.UniqueUsers)
	require.Equal(t, int64(3), activity.SuccessfulRequests)
	require.Len(t, activity.Avatars, 2)
	require.NotContains(t, mustJSON(t, activity), "42")
	require.NotContains(t, mustJSON(t, activity), "43")
	for _, avatar := range activity.Avatars {
		require.Len(t, avatar.AvatarSeed, 16)
		require.GreaterOrEqual(t, avatar.SeatIndex, 1)
		require.LessOrEqual(t, avatar.SeatIndex, cafeLobbyDisplayMax)
	}
	require.Greater(t, mini.TTL(key), time.Duration(0))
	require.Greater(t, mini.TTL(requestKey), time.Duration(0))
}

func TestCafeLobbyActivityDegradesWhenRedisUnavailable(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() { _ = client.Close() })
	service := NewCafeLobbyActivityService(client, &config.Config{Timezone: "UTC", JWT: config.JWTConfig{Secret: "test-only-cafe-secret"}})
	activity := service.Snapshot(context.Background())
	require.False(t, activity.Available)
	require.Zero(t, activity.UniqueUsers)
	require.Zero(t, activity.SuccessfulRequests)
	require.Empty(t, activity.Avatars)
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	payload, err := json.Marshal(value)
	require.NoError(t, err)
	return strings.TrimSpace(string(payload))
}
