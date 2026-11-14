package handler

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

func TestGroupHealthCacheLifetimeAndRanges(t *testing.T) {
	var cache groupHealthCache
	start := time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 7)
	calls := 0
	load := func() ([]usagestats.GroupStat, error) {
		calls++
		return []usagestats.GroupStat{{GroupID: 7, CacheReadTokens: int64(calls)}}, nil
	}
	before := time.Now()
	_, err := cache.getOrLoad(start, end, load)
	require.NoError(t, err)
	require.Len(t, cache.items, 1)
	for _, entry := range cache.items {
		require.False(t, entry.expiresAt.Before(before.Add(time.Hour)))
		require.False(t, entry.expiresAt.After(time.Now().Add(time.Hour)))
	}

	zone := time.FixedZone("UTC+8", 8*60*60)
	groups, err := cache.getOrLoad(start.In(zone), end.In(zone), load)
	require.NoError(t, err)
	require.Equal(t, 1, calls)
	require.Equal(t, int64(1), groups[0].CacheReadTokens)
	_, err = cache.getOrLoad(start.AddDate(0, 0, -1), end, load)
	require.NoError(t, err)
	require.Equal(t, 2, calls)

	// Expiration refreshes on the next request, without waiting an hour in tests.
	for key, entry := range cache.items {
		entry.expiresAt = time.Now().Add(-time.Second)
		cache.items[key] = entry
	}
	groups, err = cache.getOrLoad(start, end, load)
	require.NoError(t, err)
	require.Equal(t, 3, calls)
	require.Equal(t, int64(3), groups[0].CacheReadTokens)
}

func TestGroupHealthCacheRetriesErrorsAndCachesEmptyResults(t *testing.T) {
	var cache groupHealthCache
	start, end := time.Now(), time.Now().Add(time.Hour)
	failure := errors.New("database unavailable")
	_, err := cache.getOrLoad(start, end, func() ([]usagestats.GroupStat, error) { return nil, failure })
	require.ErrorIs(t, err, failure)
	require.Empty(t, cache.items)
	calls := 0
	for i := 0; i < 2; i++ {
		groups, err := cache.getOrLoad(start, end, func() ([]usagestats.GroupStat, error) {
			calls++
			return nil, nil
		})
		require.NoError(t, err)
		require.Empty(t, groups)
	}
	require.Equal(t, 1, calls)
}

func TestGroupHealthCacheCoalescesConcurrentRequests(t *testing.T) {
	var cache groupHealthCache
	var calls atomic.Int32
	start, end := time.Now(), time.Now().Add(time.Hour)
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	load := func() ([]usagestats.GroupStat, error) {
		calls.Add(1)
		once.Do(func() { close(started) })
		<-release
		return []usagestats.GroupStat{{GroupID: 7}}, nil
	}
	type result struct {
		groups []usagestats.GroupStat
		err    error
	}
	results := make(chan result, 16)
	for i := 0; i < cap(results); i++ {
		go func() {
			groups, err := cache.getOrLoad(start, end, load)
			results <- result{groups, err}
		}()
	}
	<-started
	close(release)
	for i := 0; i < cap(results); i++ {
		got := <-results
		require.NoError(t, got.err)
		require.Equal(t, int64(7), got.groups[0].GroupID)
	}
	require.Equal(t, int32(1), calls.Load())
}

func TestGroupHealthCacheBoundsDateRangeEntries(t *testing.T) {
	var cache groupHealthCache
	start := time.Now()
	for i := 0; i < groupHealthCacheMaxEntries+5; i++ {
		_, err := cache.getOrLoad(start.AddDate(0, 0, -i), start, func() ([]usagestats.GroupStat, error) { return nil, nil })
		require.NoError(t, err)
	}
	require.Len(t, cache.items, groupHealthCacheMaxEntries)
}
