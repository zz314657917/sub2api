package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

type leaderboardBadgeUsageRepo struct {
	UsageLogRepository
	calls        int
	leaders      *usagestats.UserLeaderboardBadgeLeaders
	callLeaders  []*usagestats.UserLeaderboardBadgeLeaders
	callUserTZs  []string
	callCostEnds []time.Time
}

func (r *leaderboardBadgeUsageRepo) GetUserLeaderboardBadgeLeaders(context.Context, time.Time, time.Time, time.Time, time.Time, time.Time, time.Time, string) (*usagestats.UserLeaderboardBadgeLeaders, error) {
	r.calls++
	if len(r.callLeaders) > 0 {
		leaders := r.callLeaders[0]
		r.callLeaders = r.callLeaders[1:]
		if leaders == nil {
			return &usagestats.UserLeaderboardBadgeLeaders{}, nil
		}
		clone := *leaders
		return &clone, nil
	}
	if r.leaders == nil {
		return &usagestats.UserLeaderboardBadgeLeaders{}, nil
	}
	leaders := *r.leaders
	return &leaders, nil
}

func TestUsageServiceGetUserLeaderboardBadgeLeadersCachesUntilNextMidnight(t *testing.T) {
	repo := &leaderboardBadgeUsageRepo{
		leaders: &usagestats.UserLeaderboardBadgeLeaders{WeeklyTokenKingUserID: 42},
	}
	svc := NewUsageService(repo, nil, nil, nil)
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	first, err := svc.GetUserLeaderboardBadgeLeaders(context.Background(), start, end, start, end, start, end, "Asia/Shanghai")
	require.NoError(t, err)
	require.Equal(t, int64(42), first.WeeklyTokenKingUserID)
	first.WeeklyTokenKingUserID = 99

	second, err := svc.GetUserLeaderboardBadgeLeaders(context.Background(), start, end, start, end, start, end, "Asia/Shanghai")
	require.NoError(t, err)
	require.Equal(t, int64(42), second.WeeklyTokenKingUserID)
	require.Equal(t, 1, repo.calls)
}

func TestUsageServiceGetUserLeaderboardBadgeLeadersDoesNotCacheEmptyResult(t *testing.T) {
	repo := &leaderboardBadgeUsageRepo{
		callLeaders: []*usagestats.UserLeaderboardBadgeLeaders{
			nil,
			{WeeklyTokenKingUserID: 42},
		},
	}
	svc := NewUsageService(repo, nil, nil, nil)
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	first, err := svc.GetUserLeaderboardBadgeLeaders(context.Background(), start, end, start, end, start, end, "Asia/Shanghai")
	require.NoError(t, err)
	require.Equal(t, int64(0), first.WeeklyTokenKingUserID)

	second, err := svc.GetUserLeaderboardBadgeLeaders(context.Background(), start, end, start, end, start, end, "Asia/Shanghai")
	require.NoError(t, err)
	require.Equal(t, int64(42), second.WeeklyTokenKingUserID)
	require.Equal(t, 2, repo.calls)
}

func TestUsageServiceInvalidateUsageCachesClearsLeaderboardBadgeCache(t *testing.T) {
	repo := &leaderboardBadgeUsageRepo{
		callLeaders: []*usagestats.UserLeaderboardBadgeLeaders{
			{WeeklyTokenKingUserID: 42},
			{WeeklyTokenKingUserID: 99},
		},
	}
	svc := NewUsageService(repo, nil, nil, nil)
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	first, err := svc.GetUserLeaderboardBadgeLeaders(context.Background(), start, end, start, end, start, end, "Asia/Shanghai")
	require.NoError(t, err)
	require.Equal(t, int64(42), first.WeeklyTokenKingUserID)

	svc.invalidateUsageCaches(context.Background(), 7, false)

	second, err := svc.GetUserLeaderboardBadgeLeaders(context.Background(), start, end, start, end, start, end, "Asia/Shanghai")
	require.NoError(t, err)
	require.Equal(t, int64(99), second.WeeklyTokenKingUserID)
	require.Equal(t, 2, repo.calls)
}
