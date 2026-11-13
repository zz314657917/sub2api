package repository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newRefreshTokenConsumeFixture(t *testing.T) (*miniredis.Miniredis, *redis.Client, service.RefreshTokenCache, service.RefreshTokenConsumer) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: -1})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	cache := NewRefreshTokenCache(client)
	consumer, ok := cache.(service.RefreshTokenConsumer)
	require.True(t, ok, "production repository must implement atomic consumption")
	return mr, client, cache, consumer
}

func TestRefreshTokenConsumeConcurrentSingleWinner(t *testing.T) {
	_, _, cache, consumer := newRefreshTokenConsumeFixture(t)
	ctx := context.Background()
	data := &service.RefreshTokenData{
		UserID: 41, TokenVersion: 4, FamilyID: "family", BindingHash: "binding",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		ExpiresAt: time.Now().UTC().Add(time.Hour).Truncate(time.Second),
	}
	require.NoError(t, cache.StoreRefreshToken(ctx, "opaque-token-hash", data, time.Hour))
	const callers = 32
	type outcome struct {
		data *service.RefreshTokenData
		err  error
	}
	results := make(chan outcome, callers)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			got, err := consumer.ConsumeRefreshToken(ctx, "opaque-token-hash")
			results <- outcome{got, err}
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	winners := 0
	for result := range results {
		if result.err == nil {
			winners++
			require.Equal(t, data, result.data)
		} else {
			require.ErrorIs(t, result.err, service.ErrRefreshTokenNotFound)
			require.Nil(t, result.data)
		}
	}
	require.Equal(t, 1, winners)
	got, err := cache.GetRefreshToken(ctx, "opaque-token-hash")
	require.Nil(t, got)
	require.ErrorIs(t, err, service.ErrRefreshTokenNotFound)
	got, err = consumer.ConsumeRefreshToken(ctx, "opaque-token-hash")
	require.Nil(t, got)
	require.ErrorIs(t, err, service.ErrRefreshTokenNotFound)
}

func TestRefreshTokenConsumeMissingExpiredCorruptAndServerError(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		_, _, _, consumer := newRefreshTokenConsumeFixture(t)
		got, err := consumer.ConsumeRefreshToken(context.Background(), "missing")
		require.Nil(t, got)
		require.ErrorIs(t, err, service.ErrRefreshTokenNotFound)
	})
	t.Run("expired", func(t *testing.T) {
		mr, _, cache, consumer := newRefreshTokenConsumeFixture(t)
		require.NoError(t, cache.StoreRefreshToken(context.Background(), "expired", &service.RefreshTokenData{UserID: 41}, time.Second))
		mr.FastForward(2 * time.Second)
		got, err := consumer.ConsumeRefreshToken(context.Background(), "expired")
		require.Nil(t, got)
		require.ErrorIs(t, err, service.ErrRefreshTokenNotFound)
	})
	for _, body := range []string{"{", "null", "[]", "\"not an object\""} {
		t.Run("corrupt/"+body, func(t *testing.T) {
			mr, _, _, consumer := newRefreshTokenConsumeFixture(t)
			require.NoError(t, mr.Set(refreshTokenKey("corrupt"), body))
			got, err := consumer.ConsumeRefreshToken(context.Background(), "corrupt")
			require.Nil(t, got)
			require.Error(t, err)
			require.False(t, errors.Is(err, service.ErrRefreshTokenNotFound))
			require.False(t, mr.Exists(refreshTokenKey("corrupt")))
			got, err = consumer.ConsumeRefreshToken(context.Background(), "corrupt")
			require.Nil(t, got)
			require.ErrorIs(t, err, service.ErrRefreshTokenNotFound)
		})
	}
	t.Run("wrong Redis type", func(t *testing.T) {
		mr, _, _, consumer := newRefreshTokenConsumeFixture(t)
		_, err := mr.SAdd(refreshTokenKey("wrong-type"), "member")
		require.NoError(t, err)
		got, err := consumer.ConsumeRefreshToken(context.Background(), "wrong-type")
		require.Nil(t, got)
		require.Error(t, err)
		require.False(t, errors.Is(err, service.ErrRefreshTokenNotFound))
		require.True(t, mr.Exists(refreshTokenKey("wrong-type")))
	})
	t.Run("server error", func(t *testing.T) {
		mr, _, cache, consumer := newRefreshTokenConsumeFixture(t)
		require.NoError(t, cache.StoreRefreshToken(context.Background(), "server-error", &service.RefreshTokenData{UserID: 41}, time.Hour))
		mr.SetError("ERR injected consume failure")
		got, err := consumer.ConsumeRefreshToken(context.Background(), "server-error")
		require.Nil(t, got)
		require.ErrorContains(t, err, "injected consume failure")
		require.False(t, errors.Is(err, service.ErrRefreshTokenNotFound))
		mr.SetError("")
		got, err = consumer.ConsumeRefreshToken(context.Background(), "server-error")
		require.NoError(t, err)
		require.Equal(t, int64(41), got.UserID)
	})
}
