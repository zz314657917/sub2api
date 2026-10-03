package repository

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type createCounterServiceRepo struct {
	service.APIKeyRepository
	created atomic.Int64
}

func (r *createCounterServiceRepo) Create(_ context.Context, key *service.APIKey) error {
	key.ID = r.created.Add(1)
	return nil
}
func (*createCounterServiceRepo) ExistsByKey(context.Context, string) (bool, error) {
	return false, nil
}

type createCounterServiceUser struct{ service.UserRepository }

func (*createCounterServiceUser) GetByID(_ context.Context, id int64) (*service.User, error) {
	return &service.User{ID: id, Status: service.StatusActive}, nil
}

func TestAPIKeyCreateCountConcurrentRealService(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	repo := &createCounterServiceRepo{}
	svc := service.NewAPIKeyService(repo, &createCounterServiceUser{}, nil, nil, nil, NewAPIKeyCache(rdb),
		&config.Config{Default: config.DefaultConfig{APIKeyPrefix: "sk-"},
			APIKeyCreate: config.APIKeyCreateConfig{MaxPerUserPerHour: 60}})
	errs := make(chan error, 100)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req := service.CreateAPIKeyRequest{}
			if i%2 == 0 {
				key := fmt.Sprintf("custom-concurrent-key-%d", i)
				req.CustomKey = &key
			}
			_, err := svc.Create(context.Background(), 10, req)
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	success, denied := 0, 0
	for err := range errs {
		if err == nil {
			success++
		} else {
			require.ErrorIs(t, err, service.ErrAPIKeyCreateRateLimited)
			denied++
		}
	}
	require.Equal(t, 60, success)
	require.Equal(t, 40, denied)
	require.EqualValues(t, 60, repo.created.Load())
	require.Equal(t, time.Hour, mr.TTL(fmt.Sprintf("%s10", apiKeyCreateCountPrefix)))
}

func TestAPIKeyCreateCountFixedTTLAndIsolation(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	cache := NewAPIKeyCache(rdb)
	counter, ok := cache.(service.APIKeyCreateCounter)
	require.True(t, ok, "production constructor must expose create counter")
	ctx := context.Background()
	n, err := counter.IncrementAPIKeyCreateCount(ctx, 7)
	require.NoError(t, err)
	require.EqualValues(t, 1, n)
	key := fmt.Sprintf("%s7", apiKeyCreateCountPrefix)
	require.Equal(t, time.Hour, mr.TTL(key))
	mr.FastForward(20 * time.Minute)
	n, err = counter.IncrementAPIKeyCreateCount(ctx, 7)
	require.NoError(t, err)
	require.EqualValues(t, 2, n)
	require.Equal(t, 40*time.Minute, mr.TTL(key))
	n, err = counter.IncrementAPIKeyCreateCount(ctx, 8)
	require.NoError(t, err)
	require.EqualValues(t, 1, n)
	require.NoError(t, cache.IncrementCreateAttemptCount(ctx, 7))
	require.NoError(t, cache.DeleteCreateAttemptCount(ctx, 7))
	value, err := mr.Get(key)
	require.NoError(t, err)
	require.Equal(t, "2", value)
	mr.FastForward(40 * time.Minute)
	n, err = counter.IncrementAPIKeyCreateCount(ctx, 7)
	require.NoError(t, err)
	require.EqualValues(t, 1, n)
	require.Equal(t, time.Hour, mr.TTL(key))
}

func TestAPIKeyCreateCountConcurrentAtomicAdmissions(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	counter := NewAPIKeyCache(rdb).(service.APIKeyCreateCounter)
	results := make(chan int64, 100)
	errs := make(chan error, 100)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := counter.IncrementAPIKeyCreateCount(context.Background(), 9)
			results <- n
			errs <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	seen, admitted := map[int64]bool{}, 0
	for n := range results {
		require.False(t, seen[n])
		seen[n] = true
		if n <= 60 {
			admitted++
		}
	}
	require.Len(t, seen, 100)
	require.Equal(t, 60, admitted)
	require.Equal(t, time.Hour, mr.TTL(fmt.Sprintf("%s9", apiKeyCreateCountPrefix)))
}

func TestAPIKeyCreateCountUnavailable(t *testing.T) {
	counter := NewAPIKeyCache(nil).(service.APIKeyCreateCounter)
	_, err := counter.IncrementAPIKeyCreateCount(context.Background(), 1)
	require.Error(t, err)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: -1})
	t.Cleanup(func() { _ = rdb.Close() })
	mr.Close()
	_, err = NewAPIKeyCache(rdb).(service.APIKeyCreateCounter).IncrementAPIKeyCreateCount(context.Background(), 1)
	require.Error(t, err)
}
