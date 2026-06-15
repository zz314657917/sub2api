package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type userLeaderboardUsageRepo struct {
	service.UsageLogRepository
	start            time.Time
	end              time.Time
	trendStart       time.Time
	trendEnd         time.Time
	trendGranularity string
	limit            int
	currentUserID    int64
	response         *usagestats.UserLeaderboardResponse
	trend            []usagestats.TrendDataPoint
	trendFromWindow  bool
	limits           []int
	currentUserIDs   []int64
	badgeLeaders     *usagestats.UserLeaderboardBadgeLeaders
}

func (r *userLeaderboardUsageRepo) GetUserLeaderboard(ctx context.Context, startTime, endTime time.Time, limit int, currentUserID int64) (*usagestats.UserLeaderboardResponse, error) {
	r.start = startTime
	r.end = endTime
	r.limit = limit
	r.currentUserID = currentUserID
	r.limits = append(r.limits, limit)
	r.currentUserIDs = append(r.currentUserIDs, currentUserID)
	if r.response != nil {
		return r.response, nil
	}
	return &usagestats.UserLeaderboardResponse{}, nil
}

func (r *userLeaderboardUsageRepo) GetUsageTrendWithFilters(ctx context.Context, startTime, endTime time.Time, granularity string, userID, apiKeyID, accountID, groupID int64, model string, requestType *int16, stream *bool, billingType *int8) ([]usagestats.TrendDataPoint, error) {
	r.trendStart = startTime
	r.trendEnd = endTime
	r.trendGranularity = granularity
	if r.trendFromWindow {
		return []usagestats.TrendDataPoint{
			{Date: startTime.AddDate(0, 0, 8).Format("2006-01-02"), TotalTokens: 180},
			{Date: startTime.AddDate(0, 0, 9).Format("2006-01-02"), TotalTokens: 420},
		}, nil
	}
	return r.trend, nil
}

func (r *userLeaderboardUsageRepo) GetUserLeaderboardBadgeLeaders(ctx context.Context, weekStart, weekEnd, monthStart, monthEnd, costStart, costEnd time.Time, userTZ string) (*usagestats.UserLeaderboardBadgeLeaders, error) {
	if r.badgeLeaders != nil {
		return r.badgeLeaders, nil
	}
	return &usagestats.UserLeaderboardBadgeLeaders{}, nil
}

func newUserLeaderboardRouter(repo *userLeaderboardUsageRepo, userID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	usageSvc := service.NewUsageService(repo, nil, nil, nil)
	handler := NewUsageHandler(usageSvc, nil)
	router := gin.New()
	router.GET("/usage/dashboard/leaderboard", func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: userID})
		handler.DashboardLeaderboard(c)
	})
	return router
}

func TestParseDashboardLeaderboardPeriodWeekUsesMondayBoundary(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	now := time.Date(2026, 5, 7, 15, 30, 0, 0, loc)

	period, start, end, startDate, endDate, err := parseDashboardLeaderboardPeriod("week", "Asia/Shanghai", now)
	require.NoError(t, err)
	require.Equal(t, "week", period)
	require.Equal(t, time.Date(2026, 5, 4, 0, 0, 0, 0, loc), start)
	require.Equal(t, time.Date(2026, 5, 11, 0, 0, 0, 0, loc), end)
	require.Equal(t, "2026-05-04", startDate)
	require.Equal(t, "2026-05-10", endDate)
}

func TestUsageHandlerDashboardLeaderboardInvalidPeriod(t *testing.T) {
	router := newUserLeaderboardRouter(&userLeaderboardUsageRepo{}, 42)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard?period=year", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUsageHandlerDashboardLeaderboardLimitClamp(t *testing.T) {
	repo := &userLeaderboardUsageRepo{}
	router := newUserLeaderboardRouter(repo, 42)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard?limit=250", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"daily_rewards"`)
	require.Contains(t, rec.Body.String(), `"enabled":false`)
	require.NotEmpty(t, repo.limits)
	require.Equal(t, 10, repo.limits[0])
	require.NotEmpty(t, repo.currentUserIDs)
	require.Equal(t, int64(42), repo.currentUserIDs[0])

	req = httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard?limit=0", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, repo.limits, 4)
	require.Equal(t, 10, repo.limits[2])
}

func TestUsageHandlerDashboardLeaderboardMasksEmailAndDisplayName(t *testing.T) {
	repo := &userLeaderboardUsageRepo{
		response: &usagestats.UserLeaderboardResponse{
			Ranking: []usagestats.UserLeaderboardItem{
				{Rank: 1, UserID: 42, Username: "raw-username@example.com", Email: "alice@example.com", ActualCost: 9.5, Requests: 2, InputTokens: 80, OutputTokens: 20, Tokens: 100, CostPer1M: 95000, IsCurrentUser: true},
			},
			CurrentUserEntry: &usagestats.UserLeaderboardItem{Rank: 1, UserID: 42, Username: "raw-username@example.com", Email: "alice@example.com", ActualCost: 9.5, Requests: 2, InputTokens: 80, OutputTokens: 20, Tokens: 100, CostPer1M: 95000, IsCurrentUser: true},
			TotalActualCost:  9.5,
			TotalRequests:    2,
			TotalTokens:      100,
		},
	}
	router := newUserLeaderboardRouter(repo, 42)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.NotContains(t, body, "alice@example.com")
	require.NotContains(t, body, "raw-username@example.com")
	require.Contains(t, body, `"email_masked":"a***e@example.com"`)
	require.Contains(t, body, `"display_name":"a***e@example.com"`)
	require.Contains(t, body, `"input_tokens":80`)
	require.Contains(t, body, `"output_tokens":20`)
	require.Contains(t, body, `"cost_per_1m_tokens":95000`)
}

func TestUsageHandlerDashboardLeaderboardIncludesRecentTokenTrend(t *testing.T) {
	repo := &userLeaderboardUsageRepo{
		trendFromWindow: true,
	}
	router := newUserLeaderboardRouter(repo, 42)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard?timezone=Asia/Shanghai", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "day", repo.trendGranularity)
	require.False(t, repo.trendStart.IsZero())
	require.False(t, repo.trendEnd.IsZero())
	firstTrendDate := repo.trendStart.AddDate(0, 0, 8).Format("2006-01-02")
	secondTrendDate := repo.trendStart.AddDate(0, 0, 9).Format("2006-01-02")
	body := rec.Body.String()
	require.Contains(t, body, `"recent_token_trend"`)
	require.Contains(t, body, `{"date":"`+firstTrendDate+`","total_tokens":180}`)
	require.Contains(t, body, `{"date":"`+secondTrendDate+`","total_tokens":420}`)
}

func TestUsageHandlerDashboardLeaderboardMasksPhoneAndQQDisplayName(t *testing.T) {
	repo := &userLeaderboardUsageRepo{
		response: &usagestats.UserLeaderboardResponse{
			Ranking: []usagestats.UserLeaderboardItem{
				{Rank: 1, UserID: 42, Username: "13812345678", ActualCost: 9.5, Requests: 2, Tokens: 100},
				{Rank: 2, UserID: 99, Username: "QQ:1234567890", ActualCost: 1.5, Requests: 1, Tokens: 20, IsCurrentUser: true},
			},
			CurrentUserEntry: &usagestats.UserLeaderboardItem{Rank: 2, UserID: 99, Username: "QQ:1234567890", ActualCost: 1.5, Requests: 1, Tokens: 20, IsCurrentUser: true},
			TotalActualCost:  11,
			TotalRequests:    3,
			TotalTokens:      120,
		},
	}
	router := newUserLeaderboardRouter(repo, 99)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.NotContains(t, body, "13812345678")
	require.NotContains(t, body, "1234567890")
	require.Contains(t, body, `"display_name":"138****5678"`)
	require.Contains(t, body, `"display_name":"QQ:12******90"`)
}

func TestUsageHandlerDashboardLeaderboardMasksWeeklyRewardTopUsers(t *testing.T) {
	repo := &userLeaderboardUsageRepo{
		response: &usagestats.UserLeaderboardResponse{
			Ranking: []usagestats.UserLeaderboardItem{
				{Rank: 1, UserID: 42, Username: "winner@example.com", Email: "winner@example.com", ActualCost: 9.5, Requests: 2, Tokens: 100, IsCurrentUser: true},
				{Rank: 2, UserID: 99, Username: "13812345678", ActualCost: 1.5, Requests: 1, Tokens: 20},
				{Rank: 3, UserID: 100, Username: "third@example.com", Email: "third@example.com", ActualCost: 1.1, Requests: 1, Tokens: 10},
			},
			CurrentUserEntry: &usagestats.UserLeaderboardItem{Rank: 1, UserID: 42, Username: "winner@example.com", Email: "winner@example.com", ActualCost: 9.5, Requests: 2, Tokens: 100, IsCurrentUser: true},
			TotalActualCost:  12.1,
		},
	}
	router := newUserLeaderboardRouter(repo, 42)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, `"top_users"`)
	require.Contains(t, body, `"display_name":"w***r@example.com"`)
	require.Contains(t, body, `"display_name":"138****5678"`)
	require.Contains(t, body, `"display_name":"t***d@example.com"`)
	require.NotContains(t, body, `"top_users":null`)
	require.NotContains(t, body, "winner@example.com")
	require.NotContains(t, body, "third@example.com")
}

func TestFinalizeLeaderboardDailyRewardsHidesRawTopUserNames(t *testing.T) {
	rewards := &usagestats.LeaderboardDailyRewards{
		TopUsers: []usagestats.LeaderboardDailyRewardTopUser{
			{Rank: 1, UserID: 42, Username: "Visible Nickname", Email: "winner@example.com"},
			{Rank: 2, UserID: 99, Username: "AB"},
			{Rank: 3, UserID: 100, Username: "陈小龙"},
		},
	}

	finalizeLeaderboardDailyRewards(rewards)

	require.Equal(t, []usagestats.LeaderboardDailyRewardTopUser{
		{Rank: 1, DisplayName: "V***e", EmailMasked: "w***r@example.com"},
		{Rank: 2, DisplayName: "A*"},
		{Rank: 3, DisplayName: "陈*龙"},
	}, rewards.TopUsers)
}

func TestUsageHandlerDashboardLeaderboardAddsBadges(t *testing.T) {
	repo := &userLeaderboardUsageRepo{
		response: &usagestats.UserLeaderboardResponse{
			Ranking: []usagestats.UserLeaderboardItem{
				{Rank: 1, UserID: 42, Username: "saver", ActualCost: 1, Requests: 2, Tokens: 100},
				{Rank: 2, UserID: 99, Username: "burner", ActualCost: 9, Requests: 1, Tokens: 20, IsCurrentUser: true},
			},
			CurrentUserEntry: &usagestats.UserLeaderboardItem{Rank: 2, UserID: 99, Username: "burner", ActualCost: 9, Requests: 1, Tokens: 20, IsCurrentUser: true},
		},
		badgeLeaders: &usagestats.UserLeaderboardBadgeLeaders{
			WeeklyTokenKingUserID:  42,
			MonthlyTokenKingUserID: 99,
			TotalTokenKingUserID:   42,
			NightOwlUserID:         99,
			BurstTokenKingUserID:   42,
			CheckinKingUserID:      99,
			CostSaverUserID:        42,
			CostBurnerUserID:       99,
		},
	}
	router := newUserLeaderboardRouter(repo, 99)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard?period=week&timezone=Asia/Shanghai", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, `"badges":["weekly_token_king","total_token_king","burst_token_king","cost_saver"]`)
	require.Contains(t, body, `"badges":["monthly_token_king","night_owl","checkin_king","cost_burner"]`)
}

func TestLeaderboardBadgesForUserOrder(t *testing.T) {
	got := leaderboardBadgesForUser(42, &usagestats.UserLeaderboardBadgeLeaders{
		WeeklyTokenKingUserID:  42,
		MonthlyTokenKingUserID: 42,
		TotalTokenKingUserID:   42,
		NightOwlUserID:         42,
		BurstTokenKingUserID:   42,
		CheckinKingUserID:      42,
		CostSaverUserID:        42,
		CostBurnerUserID:       42,
	})

	require.Equal(t, []string{
		usagestats.LeaderboardBadgeWeeklyTokenKing,
		usagestats.LeaderboardBadgeMonthlyTokenKing,
		usagestats.LeaderboardBadgeTotalTokenKing,
		usagestats.LeaderboardBadgeNightOwl,
		usagestats.LeaderboardBadgeBurstTokenKing,
		usagestats.LeaderboardBadgeCheckinKing,
		usagestats.LeaderboardBadgeCostSaver,
		usagestats.LeaderboardBadgeCostBurner,
	}, got)
}
