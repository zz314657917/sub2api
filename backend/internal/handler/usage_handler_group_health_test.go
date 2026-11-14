package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type groupHealthUsageRepoStub struct {
	service.UsageLogRepository
	groups             []usagestats.GroupStat
	calls              int
	userID             int64
	startTime, endTime time.Time
}

func (s *groupHealthUsageRepoStub) GetGroupStatsWithFilters(
	ctx context.Context,
	startTime, endTime time.Time,
	userID, apiKeyID, accountID, groupID int64,
	requestType *int16,
	stream *bool,
	billingType *int8,
) ([]usagestats.GroupStat, error) {
	s.calls++
	s.startTime = startTime
	s.endTime = endTime
	s.userID = userID
	return s.groups, nil
}

func TestDashboardGroupHealthCacheSharedAcrossUsers(t *testing.T) {
	repo := &groupHealthUsageRepoStub{groups: []usagestats.GroupStat{
		{GroupID: 7, InputTokens: 1000, CacheReadTokens: 3000},
	}}
	handler := NewUsageHandler(service.NewUsageService(repo, nil, nil, nil), nil, nil, nil)
	for _, userID := range []int64{42, 99} {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: userID})
		c.Request = httptest.NewRequest(http.MethodGet, "/usage/dashboard/group-health?start_date=2026-05-01&end_date=2026-05-07", nil)
		handler.DashboardGroupHealth(c)
		require.Equal(t, http.StatusOK, rec.Code)
	}
	require.Equal(t, 1, repo.calls)

	// A warm shared cache must not bypass authentication.
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/usage/dashboard/group-health?start_date=2026-05-01&end_date=2026-05-07", nil)
	handler.DashboardGroupHealth(c)
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Equal(t, 1, repo.calls)
}

func newGroupHealthTestRouter(repo *groupHealthUsageRepoStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewUsageHandler(service.NewUsageService(repo, nil, nil, nil), nil, nil, nil)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 42})
		c.Next()
	})
	router.GET("/usage/dashboard/group-health", handler.DashboardGroupHealth)
	return router
}

func TestDashboardGroupHealthUsesServerAggregateAndLimitsFields(t *testing.T) {
	repo := &groupHealthUsageRepoStub{
		groups: []usagestats.GroupStat{
			{GroupID: 7, GroupName: "pro max", Requests: 99, Cost: 12.5, InputTokens: 1000, CacheReadTokens: 3000},
			{GroupID: 8, GroupName: "empty", InputTokens: 0, CacheReadTokens: 0},
			{GroupID: 0, GroupName: "ungrouped", InputTokens: 100, CacheReadTokens: 100},
		},
	}
	router := newGroupHealthTestRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/usage/dashboard/group-health?start_date=2026-05-01&end_date=2026-05-08", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(0), repo.userID)
	require.True(t, repo.startTime.Before(repo.endTime))

	var envelope struct {
		Data struct {
			Groups []map[string]any `json:"groups"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Len(t, envelope.Data.Groups, 1)
	require.Equal(t, "pro max", envelope.Data.Groups[0]["group_name"])
	require.Equal(t, float64(1000), envelope.Data.Groups[0]["input_tokens"])
	require.Equal(t, float64(3000), envelope.Data.Groups[0]["cache_read_tokens"])
	_, hasRequests := envelope.Data.Groups[0]["requests"]
	require.False(t, hasRequests)
	_, hasCost := envelope.Data.Groups[0]["cost"]
	require.False(t, hasCost)
}
