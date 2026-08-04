package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

type leaderboardParticipationUserRepoStub struct {
	UserRepository
	user *User
}

func (s *leaderboardParticipationUserRepoStub) GetByID(context.Context, int64) (*User, error) {
	return s.user, nil
}

func (s *leaderboardParticipationUserRepoStub) Update(_ context.Context, user *User, _ UserUpdateFields) error {
	clone := *user
	s.user = &clone
	return nil
}

type leaderboardCacheInvalidatorRecorder struct {
	calls int
}

func (r *leaderboardCacheInvalidatorRecorder) InvalidateLeaderboardCaches() {
	r.calls++
}

func TestAdminServiceUpdateUserLeaderboardExclusionInvalidatesCache(t *testing.T) {
	repo := &leaderboardParticipationUserRepoStub{
		user: &User{ID: 42, Email: "leaderboard@example.com", Concurrency: 1},
	}
	cache := &leaderboardCacheInvalidatorRecorder{}
	svc := &adminServiceImpl{userRepo: repo, leaderboardCacheInvalidator: cache}

	excluded := true
	updated, err := svc.UpdateUser(context.Background(), 42, &UpdateUserInput{ExcludeFromLeaderboard: &excluded})
	require.NoError(t, err)
	require.True(t, updated.ExcludeFromLeaderboard)
	require.True(t, repo.user.ExcludeFromLeaderboard)
	require.Equal(t, 1, cache.calls)

	_, err = svc.UpdateUser(context.Background(), 42, &UpdateUserInput{})
	require.NoError(t, err)
	require.True(t, repo.user.ExcludeFromLeaderboard, "omitted field must preserve the stored setting")
	require.Equal(t, 1, cache.calls)

	included := false
	_, err = svc.UpdateUser(context.Background(), 42, &UpdateUserInput{ExcludeFromLeaderboard: &included})
	require.NoError(t, err)
	require.False(t, repo.user.ExcludeFromLeaderboard)
	require.Equal(t, 2, cache.calls)
}

func TestUsageServiceInvalidateLeaderboardCaches(t *testing.T) {
	svc := &UsageService{
		badgeCache: map[string]leaderboardBadgeCacheEntry{
			"badges": {leaders: &usagestats.UserLeaderboardBadgeLeaders{WeeklyTokenKingUserID: 42}},
		},
		userLeaderboardCache: map[string]leaderboardUserCacheEntry{
			"users": {leaderboard: &usagestats.UserLeaderboardResponse{}},
		},
		modelRankingCache: map[string]leaderboardModelRankingCacheEntry{
			"models": {},
		},
		recentTrendCache: map[string]leaderboardRecentTrendCacheEntry{
			"trend": {},
		},
		dailyChampionsCache: map[string]leaderboardDailyChampionsCacheEntry{
			"champions": {},
		},
	}

	svc.InvalidateLeaderboardCaches()

	require.Empty(t, svc.badgeCache)
	require.Empty(t, svc.userLeaderboardCache)
	require.Empty(t, svc.modelRankingCache)
	require.Empty(t, svc.recentTrendCache)
	require.Empty(t, svc.dailyChampionsCache)
}
