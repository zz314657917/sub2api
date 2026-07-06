package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/stretchr/testify/require"
)

type leaderboardStatsCacheUsageRepo struct {
	UsageLogRepository
	userCalls          int
	userResponses      []*usagestats.UserLeaderboardResponse
	modelCalls         int
	modelItems         []usagestats.UserLeaderboardModelItem
	modelTotal         int64
	trendCalls         int
	trend              []usagestats.TrendDataPoint
	dailyChampionCalls int
	dailyChampions     []usagestats.UserLeaderboardDailyChampion
	createdUsageLogs   []*UsageLog
}

func (r *leaderboardStatsCacheUsageRepo) GetUserLeaderboard(context.Context, time.Time, time.Time, int, int64) (*usagestats.UserLeaderboardResponse, error) {
	r.userCalls++
	if len(r.userResponses) == 0 {
		return &usagestats.UserLeaderboardResponse{}, nil
	}
	index := r.userCalls - 1
	if index >= len(r.userResponses) {
		index = len(r.userResponses) - 1
	}
	return cloneUserLeaderboardResponse(r.userResponses[index]), nil
}

func (r *leaderboardStatsCacheUsageRepo) GetLeaderboardModelRanking(context.Context, time.Time, time.Time, int) ([]usagestats.UserLeaderboardModelItem, int64, error) {
	r.modelCalls++
	return cloneUserLeaderboardModelItems(r.modelItems), r.modelTotal, nil
}

func (r *leaderboardStatsCacheUsageRepo) GetUsageTrendWithFilters(context.Context, time.Time, time.Time, string, int64, int64, int64, int64, string, *int16, *bool, *int8) ([]usagestats.TrendDataPoint, error) {
	r.trendCalls++
	points := make([]usagestats.TrendDataPoint, len(r.trend))
	copy(points, r.trend)
	return points, nil
}

func (r *leaderboardStatsCacheUsageRepo) GetLeaderboardDailyChampions(context.Context, time.Time, time.Time) ([]usagestats.UserLeaderboardDailyChampion, error) {
	r.dailyChampionCalls++
	return cloneLeaderboardDailyChampions(r.dailyChampions), nil
}

func (r *leaderboardStatsCacheUsageRepo) Create(_ context.Context, log *UsageLog) (bool, error) {
	r.createdUsageLogs = append(r.createdUsageLogs, log)
	return true, nil
}

func TestUsageServiceGetUserLeaderboardCachesStatsAndReturnsClones(t *testing.T) {
	avatarURL := "https://example.test/avatar.png"
	repo := &leaderboardStatsCacheUsageRepo{
		userResponses: []*usagestats.UserLeaderboardResponse{
			{
				TotalTokens: 100,
				Ranking: []usagestats.UserLeaderboardItem{
					{Rank: 1, UserID: 42, Username: "alice", Email: "alice@example.com", Tokens: 100, AvatarURL: &avatarURL, Badges: []string{usagestats.LeaderboardBadgeWeeklyTokenKing}},
				},
				CurrentUserEntry: &usagestats.UserLeaderboardItem{Rank: 1, UserID: 42, Tokens: 100},
			},
		},
	}
	svc := NewUsageService(repo, nil, nil, nil)
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	first, err := svc.GetUserLeaderboard(context.Background(), start, end, 10, 42)
	require.NoError(t, err)
	require.Len(t, first.Ranking, 1)
	first.Ranking[0].Tokens = 999
	first.Ranking[0].Badges[0] = "mutated"
	*first.Ranking[0].AvatarURL = "mutated"
	first.CurrentUserEntry.Tokens = 999

	second, err := svc.GetUserLeaderboard(context.Background(), start, end, 10, 42)
	require.NoError(t, err)
	require.Equal(t, 1, repo.userCalls)
	require.Equal(t, int64(100), second.Ranking[0].Tokens)
	require.Equal(t, []string{usagestats.LeaderboardBadgeWeeklyTokenKing}, second.Ranking[0].Badges)
	require.Equal(t, avatarURL, *second.Ranking[0].AvatarURL)
	require.Equal(t, int64(100), second.CurrentUserEntry.Tokens)
}

func TestUsageServiceGetUserLeaderboardCacheKeyIncludesCurrentUser(t *testing.T) {
	repo := &leaderboardStatsCacheUsageRepo{
		userResponses: []*usagestats.UserLeaderboardResponse{
			{CurrentUserEntry: &usagestats.UserLeaderboardItem{UserID: 42, Tokens: 100}},
			{CurrentUserEntry: &usagestats.UserLeaderboardItem{UserID: 99, Tokens: 200}},
		},
	}
	svc := NewUsageService(repo, nil, nil, nil)
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	first, err := svc.GetUserLeaderboard(context.Background(), start, end, 10, 42)
	require.NoError(t, err)
	second, err := svc.GetUserLeaderboard(context.Background(), start, end, 10, 99)
	require.NoError(t, err)

	require.Equal(t, 2, repo.userCalls)
	require.Equal(t, int64(42), first.CurrentUserEntry.UserID)
	require.Equal(t, int64(99), second.CurrentUserEntry.UserID)
}

func TestUsageServiceGetLeaderboardModelRankingAndRecentTrendUseShortCache(t *testing.T) {
	growth := 12.5
	rankChange := int64(1)
	repo := &leaderboardStatsCacheUsageRepo{
		modelItems: []usagestats.UserLeaderboardModelItem{
			{Rank: 1, Model: "gpt-5.5", Tokens: 300, GrowthPercent: &growth, RankChange: &rankChange},
		},
		modelTotal: 1,
		trend: []usagestats.TrendDataPoint{
			{Date: "2026-05-01", TotalTokens: 300},
		},
	}
	svc := NewUsageService(repo, nil, nil, nil)
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	models, total, err := svc.GetLeaderboardModelRanking(context.Background(), start, end, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	*models[0].GrowthPercent = 99
	*models[0].RankChange = 99

	modelsAgain, totalAgain, err := svc.GetLeaderboardModelRanking(context.Background(), start, end, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), totalAgain)
	require.Equal(t, 1, repo.modelCalls)
	require.Equal(t, 12.5, *modelsAgain[0].GrowthPercent)
	require.Equal(t, int64(1), *modelsAgain[0].RankChange)

	trend, err := svc.GetLeaderboardRecentTokenTrend(context.Background(), start, end)
	require.NoError(t, err)
	trend[0].TotalTokens = 999

	trendAgain, err := svc.GetLeaderboardRecentTokenTrend(context.Background(), start, end)
	require.NoError(t, err)
	require.Equal(t, 1, repo.trendCalls)
	require.Equal(t, int64(300), trendAgain[0].TotalTokens)
}

func TestUsageServiceGetLeaderboardDailyChampionsCachesAndReturnsClones(t *testing.T) {
	avatarURL := "https://example.test/champion.png"
	repo := &leaderboardStatsCacheUsageRepo{
		dailyChampions: []usagestats.UserLeaderboardDailyChampion{
			{
				Date:        "2026-06-10",
				UserID:      42,
				DisplayName: "Champion",
				EmailMasked: "c***@example.com",
				AvatarURL:   &avatarURL,
				Tokens:      48_163_000,
			},
		},
	}
	svc := NewUsageService(repo, nil, nil, nil)
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	first, err := svc.GetLeaderboardDailyChampions(context.Background(), start, end)
	require.NoError(t, err)
	require.Len(t, first, 1)
	first[0].Tokens = 1
	*first[0].AvatarURL = "mutated"

	second, err := svc.GetLeaderboardDailyChampions(context.Background(), start, end)
	require.NoError(t, err)
	require.Equal(t, 1, repo.dailyChampionCalls)
	require.Equal(t, int64(48_163_000), second[0].Tokens)
	require.Equal(t, avatarURL, *second[0].AvatarURL)

	second[0].Tokens = 2
	*second[0].AvatarURL = "mutated-again"
	third, err := svc.GetLeaderboardDailyChampions(context.Background(), start, end)
	require.NoError(t, err)
	require.Equal(t, 1, repo.dailyChampionCalls)
	require.Equal(t, int64(48_163_000), third[0].Tokens)
	require.Equal(t, avatarURL, *third[0].AvatarURL)
}

func TestUsageServiceGetUserLeaderboardCacheKeyIncludesTimeRange(t *testing.T) {
	repo := &leaderboardStatsCacheUsageRepo{
		userResponses: []*usagestats.UserLeaderboardResponse{
			{TotalTokens: 100},
			{TotalTokens: 200},
		},
	}
	svc := NewUsageService(repo, nil, nil, nil)
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	first, err := svc.GetUserLeaderboard(context.Background(), start, end, 10, 42)
	require.NoError(t, err)
	require.Equal(t, int64(100), first.TotalTokens)

	second, err := svc.GetUserLeaderboard(context.Background(), start.Add(24*time.Hour), end.Add(24*time.Hour), 10, 42)
	require.NoError(t, err)

	require.Equal(t, int64(200), second.TotalTokens)
	require.Equal(t, 2, repo.userCalls)
}
