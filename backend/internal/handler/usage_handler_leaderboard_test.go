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
	start         time.Time
	end           time.Time
	limit         int
	currentUserID int64
	response      *usagestats.UserLeaderboardResponse
}

func (r *userLeaderboardUsageRepo) GetUserLeaderboard(ctx context.Context, startTime, endTime time.Time, limit int, currentUserID int64) (*usagestats.UserLeaderboardResponse, error) {
	r.start = startTime
	r.end = endTime
	r.limit = limit
	r.currentUserID = currentUserID
	if r.response != nil {
		return r.response, nil
	}
	return &usagestats.UserLeaderboardResponse{}, nil
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
	require.Equal(t, 100, repo.limit)
	require.Equal(t, int64(42), repo.currentUserID)

	req = httptest.NewRequest(http.MethodGet, "/usage/dashboard/leaderboard?limit=0", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 20, repo.limit)
}

func TestUsageHandlerDashboardLeaderboardMasksEmailAndDisplayName(t *testing.T) {
	repo := &userLeaderboardUsageRepo{
		response: &usagestats.UserLeaderboardResponse{
			Ranking: []usagestats.UserLeaderboardItem{
				{Rank: 1, UserID: 42, Username: "raw-username@example.com", Email: "alice@example.com", ActualCost: 9.5, Requests: 2, Tokens: 100, IsCurrentUser: true},
			},
			CurrentUserEntry: &usagestats.UserLeaderboardItem{Rank: 1, UserID: 42, Username: "raw-username@example.com", Email: "alice@example.com", ActualCost: 9.5, Requests: 2, Tokens: 100, IsCurrentUser: true},
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
}
