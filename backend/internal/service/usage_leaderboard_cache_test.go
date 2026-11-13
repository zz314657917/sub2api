package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

type refreshingLeaderboardRepo struct {
	UsageLogRepository
	query func(context.Context) (int64, error)
}

func (r *refreshingLeaderboardRepo) GetUserLeaderboard(ctx context.Context, _, _ time.Time, _ int, _ int64) (*usagestats.UserLeaderboardResponse, error) {
	n, err := r.query(ctx)
	return &usagestats.UserLeaderboardResponse{TotalTokens: n, Ranking: []usagestats.UserLeaderboardItem{{UserID: n, Tokens: n}}}, err
}

func (r *refreshingLeaderboardRepo) GetLeaderboardModelRanking(ctx context.Context, _, _ time.Time, _ int) ([]usagestats.UserLeaderboardModelItem, int64, error) {
	n, err := r.query(ctx)
	return []usagestats.UserLeaderboardModelItem{{Tokens: n}}, n, err
}

func (r *refreshingLeaderboardRepo) GetLeaderboardRecentTokenTrend(ctx context.Context, _, _ time.Time) ([]usagestats.UserLeaderboardTokenTrendPoint, error) {
	n, err := r.query(ctx)
	return []usagestats.UserLeaderboardTokenTrendPoint{{TotalTokens: n}}, err
}

func (r *refreshingLeaderboardRepo) GetLeaderboardDailyChampions(ctx context.Context, _, _ time.Time) ([]usagestats.UserLeaderboardDailyChampion, error) {
	n, err := r.query(ctx)
	return []usagestats.UserLeaderboardDailyChampion{{Tokens: n}}, err
}

func (r *refreshingLeaderboardRepo) GetUserLeaderboardBadgeLeaders(ctx context.Context, _, _, _, _, _, _ time.Time, _ string) (*usagestats.UserLeaderboardBadgeLeaders, error) {
	n, err := r.query(ctx)
	return &usagestats.UserLeaderboardBadgeLeaders{WeeklyTokenKingUserID: n}, err
}

var leaderboardCacheKinds = []string{"user", "model", "trend", "daily-champions", "badge"}

func leaderboardCacheScenario(s *UsageService, kind string, start time.Time) (func(time.Time, int64), func(context.Context) (int64, error)) {
	end := start.AddDate(0, 0, 1)
	key := leaderboardStatsCacheKey(start, end, 0)
	switch kind {
	case "user":
		return func(now time.Time, n int64) {
				s.setCachedUserLeaderboard(key, &usagestats.UserLeaderboardResponse{TotalTokens: n, GeneratedAt: now.UTC().Format(time.RFC3339)}, now)
			}, func(ctx context.Context) (int64, error) {
				v, err := s.GetUserLeaderboard(ctx, start, end, 10, 0)
				if err != nil {
					return 0, err
				}
				return v.TotalTokens, nil
			}
	case "model":
		key = leaderboardStatsCacheKey(start, end, 10)
		return func(now time.Time, n int64) {
				s.setCachedLeaderboardModelRanking(key, []usagestats.UserLeaderboardModelItem{{Tokens: n}}, n, now)
			}, func(ctx context.Context) (int64, error) {
				v, _, err := s.GetLeaderboardModelRanking(ctx, start, end, 10)
				if err != nil {
					return 0, err
				}
				return v[0].Tokens, nil
			}
	case "trend":
		return func(now time.Time, n int64) {
				s.setCachedLeaderboardRecentTrend(key, []usagestats.UserLeaderboardTokenTrendPoint{{TotalTokens: n}}, now)
			}, func(ctx context.Context) (int64, error) {
				v, err := s.GetLeaderboardRecentTokenTrend(ctx, start, end)
				if err != nil {
					return 0, err
				}
				return v[0].TotalTokens, nil
			}
	case "daily-champions":
		return func(now time.Time, n int64) {
				s.setCachedLeaderboardDailyChampions(key, []usagestats.UserLeaderboardDailyChampion{{Tokens: n}}, now)
			}, func(ctx context.Context) (int64, error) {
				v, err := s.GetLeaderboardDailyChampions(ctx, start, end)
				if err != nil {
					return 0, err
				}
				return v[0].Tokens, nil
			}
	default:
		key = leaderboardBadgeCacheKey(start, end, start, end, start, end, "UTC")
		return func(now time.Time, n int64) {
				s.setCachedLeaderboardBadgeLeaders(key, &usagestats.UserLeaderboardBadgeLeaders{WeeklyTokenKingUserID: n}, now.Add(leaderboardStatsCacheTTL), now)
			}, func(ctx context.Context) (int64, error) {
				v, err := s.GetUserLeaderboardBadgeLeaders(ctx, start, end, start, end, start, end, "UTC")
				if err != nil {
					return 0, err
				}
				return v.WeeklyTokenKingUserID, nil
			}
	}
}

func TestLeaderboardCacheStaleRefreshFailureAndRecovery(t *testing.T) {
	for _, kind := range leaderboardCacheKinds {
		t.Run(kind, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var calls atomic.Int32
				var fail atomic.Bool
				fail.Store(true)
				release := make(chan struct{})
				repo := &refreshingLeaderboardRepo{query: func(ctx context.Context) (int64, error) {
					calls.Add(1)
					<-release
					if fail.Load() {
						return 0, errors.New("database unavailable")
					}
					return 22, nil
				}}
				svc := NewUsageService(repo, nil, nil, nil)
				start := time.Now()
				seed, fetch := leaderboardCacheScenario(svc, kind, start)
				seed(start.Add(-leaderboardStatsCacheTTL-time.Minute), 11)
				// A write to another window must not evict this stale fallback.
				seedOther, _ := leaderboardCacheScenario(svc, kind, start.AddDate(0, 0, -1))
				seedOther(start, 33)
				for range 20 {
					v, err := fetch(context.Background())
					require.NoError(t, err)
					require.Equal(t, int64(11), v)
				}
				synctest.Wait()
				require.Equal(t, int32(1), calls.Load())
				close(release)
				synctest.Wait()
				v, err := fetch(context.Background())
				require.NoError(t, err)
				require.Equal(t, int64(11), v)
				require.Equal(t, int32(1), calls.Load(), "failed refresh must back off")
				fail.Store(false)
				time.Sleep(leaderboardRetryInitial)
				_, err = fetch(context.Background())
				require.NoError(t, err)
				synctest.Wait()
				v, err = fetch(context.Background())
				require.NoError(t, err)
				require.Equal(t, int64(22), v)
				require.Equal(t, int32(2), calls.Load())
			})
		})
	}
}

func TestLeaderboardCacheColdCoalescingAndCallerCancellation(t *testing.T) {
	for _, kind := range leaderboardCacheKinds {
		t.Run(kind, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var calls atomic.Int32
				release := make(chan struct{})
				var queryErr error
				repo := &refreshingLeaderboardRepo{query: func(ctx context.Context) (int64, error) {
					calls.Add(1)
					<-release
					queryErr = ctx.Err()
					return 42, queryErr
				}}
				svc := NewUsageService(repo, nil, nil, nil)
				_, fetch := leaderboardCacheScenario(svc, kind, time.Now())
				ctx, cancel := context.WithCancel(context.Background())
				cancelled := make(chan error, 1)
				go func() { _, err := fetch(ctx); cancelled <- err }()
				synctest.Wait()
				results := make(chan int64, 16)
				for range 16 {
					go func() { v, _ := fetch(context.Background()); results <- v }()
				}
				synctest.Wait()
				cancel()
				require.ErrorIs(t, <-cancelled, context.Canceled)
				close(release)
				for range 16 {
					require.Equal(t, int64(42), <-results)
				}
				require.Equal(t, int32(1), calls.Load())
				require.NoError(t, queryErr)
			})
		})
	}
}

func TestLeaderboardCacheInvalidationDiscardsInflightResult(t *testing.T) {
	for _, kind := range leaderboardCacheKinds {
		t.Run(kind, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var calls atomic.Int32
				oldQuery := make(chan struct{})
				repo := &refreshingLeaderboardRepo{query: func(context.Context) (int64, error) {
					if calls.Add(1) == 1 {
						<-oldQuery
						return 11, nil
					}
					return 22, nil
				}}
				svc := NewUsageService(repo, nil, nil, nil)
				seed, fetch := leaderboardCacheScenario(svc, kind, time.Now())
				seed(time.Now().Add(-2*time.Hour), 1)
				_, err := fetch(context.Background())
				require.NoError(t, err)
				synctest.Wait()
				svc.InvalidateLeaderboardCaches()
				v, err := fetch(context.Background())
				require.NoError(t, err)
				require.Equal(t, int64(22), v)
				close(oldQuery)
				synctest.Wait()
				v, err = fetch(context.Background())
				require.NoError(t, err)
				require.Equal(t, int64(22), v, "late refresh must not restore invalidated users")
			})
		})
	}
}

func TestLeaderboardCacheBoundedStalenessAndBackoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int32
		failure := errors.New("query failed")
		repo := &refreshingLeaderboardRepo{query: func(context.Context) (int64, error) { calls.Add(1); return 0, failure }}
		svc := NewUsageService(repo, nil, nil, nil)
		seed, fetch := leaderboardCacheScenario(svc, "user", time.Now())
		seed(time.Now().Add(-leaderboardStatsCacheTTL-leaderboardStaleRetention-time.Second), 11)
		delay := leaderboardRetryInitial
		for i := int32(1); i <= 9; i++ {
			_, err := fetch(context.Background())
			require.ErrorIs(t, err, failure, "over-age snapshot must not be served")
			require.Equal(t, i, calls.Load())
			time.Sleep(delay - time.Nanosecond)
			_, err = fetch(context.Background())
			require.ErrorIs(t, err, failure)
			require.Equal(t, i, calls.Load())
			time.Sleep(time.Nanosecond)
			delay = min(delay*2, leaderboardRetryMaximum)
		}
	})
}

func TestLeaderboardCacheRewardPathsRejectStaleRanking(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		failure := errors.New("query failed")
		svc := NewUsageService(&refreshingLeaderboardRepo{query: func(context.Context) (int64, error) { return 0, failure }}, nil, nil, nil)
		now := time.Now()
		start, end, _, _, _ := leaderboardRewardWindow(now)
		svc.setCachedUserLeaderboard(leaderboardStatsCacheKey(start, end, 0), &usagestats.UserLeaderboardResponse{TotalTokens: 11}, now.Add(-2*time.Hour))
		_, err := svc.getLeaderboardDailyRewards(context.Background(), 1, now)
		require.ErrorIs(t, err, failure)
		_, _, _, err = svc.prepareLeaderboardRewardClaim(context.Background(), 1, now)
		require.ErrorIs(t, err, failure)
	})
}

func TestLeaderboardCacheRefreshTimeoutAndOriginalTimestamp(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var calls atomic.Int32
		repo := &refreshingLeaderboardRepo{query: func(ctx context.Context) (int64, error) {
			calls.Add(1)
			<-ctx.Done()
			return 0, ctx.Err()
		}}
		svc := NewUsageService(repo, nil, nil, nil)
		start := time.Now()
		end := start.AddDate(0, 0, 1)
		seed, _ := leaderboardCacheScenario(svc, "user", start)
		oldTime := start.Add(-2 * time.Hour)
		seed(oldTime, 11)
		view, err := svc.GetUserLeaderboard(context.Background(), start, end, 10, 0)
		require.NoError(t, err)
		require.Equal(t, oldTime.UTC().Format(time.RFC3339), view.GeneratedAt)
		synctest.Wait()
		time.Sleep(leaderboardRefreshTimeout)
		synctest.Wait()
		_, err = svc.getUserLeaderboard(context.Background(), start, end, 10, 0, false)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Equal(t, int32(1), calls.Load())
		view, err = svc.GetUserLeaderboard(context.Background(), start, end, 10, 0)
		require.NoError(t, err)
		require.Equal(t, oldTime.UTC().Format(time.RFC3339), view.GeneratedAt)
	})
}
