package handler

import (
	"context"
	"errors"
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
	modelStart       time.Time
	modelEnd         time.Time
	modelLimit       int
	modelRanking     []usagestats.UserLeaderboardModelItem
	totalModels      int64
	modelErr         error
	championStart    time.Time
	championEnd      time.Time
	dailyChampions   []usagestats.UserLeaderboardDailyChampion
}

type leaderboardHandlerUserRepo struct {
	service.UserRepository
	createdAt time.Time
	getCalls  int
}

func (r *leaderboardHandlerUserRepo) GetByID(context.Context, int64) (*service.User, error) {
	r.getCalls++
	return &service.User{ID: 42, CreatedAt: r.createdAt}, nil
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

func (r *userLeaderboardUsageRepo) GetLeaderboardModelRanking(ctx context.Context, startTime, endTime time.Time, limit int) ([]usagestats.UserLeaderboardModelItem, int64, error) {
	r.modelStart = startTime
	r.modelEnd = endTime
	r.modelLimit = limit
	if r.modelErr != nil {
		return nil, 0, r.modelErr
	}
	if r.modelRanking != nil {
		return r.modelRanking, r.totalModels, nil
	}
	return []usagestats.UserLeaderboardModelItem{}, 0, nil
}

func (r *userLeaderboardUsageRepo) GetLeaderboardDailyChampions(ctx context.Context, startTime, endTime time.Time) ([]usagestats.UserLeaderboardDailyChampion, error) {
	r.championStart = startTime
	r.championEnd = endTime
	if r.dailyChampions != nil {
		return r.dailyChampions, nil
	}
	return []usagestats.UserLeaderboardDailyChampion{}, nil
}

func newUserLeaderboardRouter(repo *userLeaderboardUsageRepo, userID int64) *gin.Engine {
	return newUserLeaderboardRouterWithCreatedAt(repo, userID, time.Now().Add(-8*24*time.Hour))
}

func newUserLeaderboardRouterWithCreatedAt(repo *userLeaderboardUsageRepo, userID int64, createdAt time.Time) *gin.Engine {
	gin.SetMode(gin.TestMode)
	userRepo := &leaderboardHandlerUserRepo{createdAt: createdAt}
	usageSvc := service.NewUsageService(repo, userRepo, nil, nil)
	handler := NewUsageHandler(usageSvc, nil)
	router := gin.New()
	router.GET("/usage/dashboard/leaderboard", func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: userID})
		handler.DashboardLeaderboard(c)
	})
	router.POST("/usage/dashboard/leaderboard/daily-reward/claim", func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: userID})
		handler.ClaimDashboardLeaderboardDailyReward(c)
	})
	return router
}

func TestUsageHandlerDashboardLeaderboardAccountAgeGate(t *testing.T) {
	t.Run("rejects account younger than seven days before ranking query", func(t *testing.T) {
		repo := &userLeaderboardUsageRepo{}
		router := newUserLeaderboardRouterWithCreatedAt(repo, 42, time.Now().Add(-6*24*time.Hour))

		req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
		require.Contains(t, rec.Body.String(), `"reason":"LEADERBOARD_ACCOUNT_TOO_NEW"`)
		require.Empty(t, repo.limits)
	})

	t.Run("allows account at seven-day boundary", func(t *testing.T) {
		repo := &userLeaderboardUsageRepo{}
		router := newUserLeaderboardRouterWithCreatedAt(repo, 42, time.Now().Add(-7*24*time.Hour))

		req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		require.NotEmpty(t, repo.limits)
	})
}

func TestUsageHandlerClaimDashboardLeaderboardAccountAgeGate(t *testing.T) {
	repo := &userLeaderboardUsageRepo{}
	router := newUserLeaderboardRouterWithCreatedAt(repo, 42, time.Now().Add(-6*24*time.Hour))

	req := httptest.NewRequest(http.MethodPost, "/usage/dashboard/leaderboard/daily-reward/claim", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusForbidden, rec.Code)
	require.Contains(t, rec.Body.String(), `"reason":"LEADERBOARD_ACCOUNT_TOO_NEW"`)
	require.Empty(t, repo.limits)
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

	req = httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard?period=week&limit=0", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, repo.limits, 3)
	require.Equal(t, 10, repo.limits[2])
}

func TestUsageHandlerDashboardLeaderboardMasksEmailAndDisplayName(t *testing.T) {
	rankChange := int64(1)
	repo := &userLeaderboardUsageRepo{
		response: &usagestats.UserLeaderboardResponse{
			Ranking: []usagestats.UserLeaderboardItem{
				{Rank: 1, UserID: 42, Username: "raw-username@example.com", Email: "alice@example.com", ActualCost: 9.5, Requests: 2, InputTokens: 80, OutputTokens: 20, Tokens: 100, CostPer1M: 95000, RankChange: &rankChange, RankNew: true, IsCurrentUser: true},
			},
			CurrentUserEntry: &usagestats.UserLeaderboardItem{Rank: 1, UserID: 42, Username: "raw-username@example.com", Email: "alice@example.com", ActualCost: 9.5, Requests: 2, InputTokens: 80, OutputTokens: 20, Tokens: 100, CostPer1M: 95000, RankChange: &rankChange, RankNew: true, IsCurrentUser: true},
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
	require.Contains(t, body, `"rank_change":1`)
	require.Contains(t, body, `"rank_new":true`)
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

func TestUsageHandlerDashboardLeaderboardIncludesModelRanking(t *testing.T) {
	t.Setenv("SUB2API_LEADERBOARD_SAMPLE_MODELS", "false")
	growth := 87.3
	rankChange := int64(1)
	repo := &userLeaderboardUsageRepo{
		modelRanking: []usagestats.UserLeaderboardModelItem{
			{Rank: 1, Model: "gpt-5.5", Requests: 12, InputTokens: 700, OutputTokens: 300, Tokens: 1000, GrowthPercent: &growth, RankChange: &rankChange},
			{Rank: 2, Model: "claude-opus-4-8", Requests: 5, InputTokens: 400, OutputTokens: 100, Tokens: 500},
		},
		totalModels: 8,
	}
	router := newUserLeaderboardRouter(repo, 42)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard?period=week&timezone=Asia/Shanghai", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 10, repo.modelLimit)
	body := rec.Body.String()
	require.Contains(t, body, `"total_models":8`)
	require.Contains(t, body, `"model_ranking":[{"rank":1,"model":"gpt-5.5","requests":12,"input_tokens":700,"output_tokens":300,"tokens":1000,"growth_percent":87.3,"rank_change":1}`)
	require.Contains(t, body, `{"rank":2,"model":"claude-opus-4-8","requests":5,"input_tokens":400,"output_tokens":100,"tokens":500}`)
}

func TestUsageHandlerDashboardLeaderboardReturnsRankingWhenModelRankingFails(t *testing.T) {
	t.Setenv("SUB2API_LEADERBOARD_SAMPLE_MODELS", "false")
	repo := &userLeaderboardUsageRepo{
		response: &usagestats.UserLeaderboardResponse{
			Ranking: []usagestats.UserLeaderboardItem{
				{Rank: 1, UserID: 42, Username: "winner", ActualCost: 9.5, Requests: 2, Tokens: 100, IsCurrentUser: true},
			},
			CurrentUserEntry: &usagestats.UserLeaderboardItem{Rank: 1, UserID: 42, Username: "winner", ActualCost: 9.5, Requests: 2, Tokens: 100, IsCurrentUser: true},
			TotalActualCost:  9.5,
			TotalRequests:    2,
			TotalTokens:      100,
		},
		modelErr: errors.New("model ranking timeout"),
	}
	router := newUserLeaderboardRouter(repo, 42)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard?period=month&timezone=Asia/Shanghai", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, `"ranking":[{"rank":1`)
	require.Contains(t, body, `"display_name":"w***r"`)
	require.Contains(t, body, `"model_ranking":[]`)
	require.Contains(t, body, `"total_models":0`)
}

func TestUsageHandlerDashboardLeaderboardIncludesMaskedDailyChampions(t *testing.T) {
	avatarURL := "https://cdn.example.com/champion.png"
	repo := &userLeaderboardUsageRepo{
		dailyChampions: []usagestats.UserLeaderboardDailyChampion{
			{
				Date:      "2026-06-10",
				UserID:    42,
				Username:  "raw-winner@example.com",
				Email:     "winner@example.com",
				AvatarURL: &avatarURL,
				Tokens:    48_163_000,
			},
			{
				Date:     "2026-06-11",
				UserID:   99,
				Username: "13812345678",
				Tokens:   12_000,
			},
		},
	}
	router := newUserLeaderboardRouter(repo, 42)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard?timezone=Asia/Shanghai", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.False(t, repo.championStart.IsZero())
	require.False(t, repo.championEnd.IsZero())
	body := rec.Body.String()
	require.Contains(t, body, `"daily_champions"`)
	require.Contains(t, body, `"date":"2026-06-10"`)
	require.Contains(t, body, `"display_name":"w***r@example.com"`)
	require.Contains(t, body, `"email_masked":"w***r@example.com"`)
	require.Contains(t, body, `"avatar_url":"https://cdn.example.com/champion.png"`)
	require.Contains(t, body, `"tokens":48163000`)
	require.Contains(t, body, `"date":"2026-06-11"`)
	require.Contains(t, body, `"display_name":"138****5678"`)
	require.NotContains(t, body, "winner@example.com")
	require.NotContains(t, body, "raw-winner@example.com")
	require.NotContains(t, body, "13812345678")
}

func TestUsageHandlerDashboardLeaderboardUsesSampleModelRankingWhenEnabled(t *testing.T) {
	t.Setenv("SUB2API_LEADERBOARD_SAMPLE_MODELS", "true")
	repo := &userLeaderboardUsageRepo{}
	router := newUserLeaderboardRouter(repo, 42)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard?period=day", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	require.Contains(t, body, `"total_models":3`)
	require.Contains(t, body, `"model":"gpt-5.5"`)
	require.Contains(t, body, `"growth_percent":-77.7`)
	require.Contains(t, body, `"rank_change":1`)
	require.Contains(t, body, `"model":"claude-opus-4-8"`)
	require.Contains(t, body, `"rank_change":-1`)
	require.Contains(t, body, `"model":"gpt-5.4"`)
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
