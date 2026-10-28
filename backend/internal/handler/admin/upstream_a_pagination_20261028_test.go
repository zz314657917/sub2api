package admin

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type upstreamAPaginationOpsRepo struct {
	service.OpsRepository
	calls  int
	filter *service.OpsOpenAITokenStatsFilter
}

func (r *upstreamAPaginationOpsRepo) GetOpenAITokenStats(_ context.Context, filter *service.OpsOpenAITokenStatsFilter) (*service.OpsOpenAITokenStatsResponse, error) {
	r.calls++
	r.filter = filter
	return &service.OpsOpenAITokenStatsResponse{}, nil
}

func TestUpstreamAFixesOpsPaginationRejectsImpossiblePage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &upstreamAPaginationOpsRepo{}
	svc := service.NewOpsService(repo, nil, &config.Config{Ops: config.OpsConfig{Enabled: true}}, nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewOpsHandler(svc)
	router := gin.New()
	router.GET("/stats", handler.GetDashboardOpenAITokenStats)

	for _, query := range []string{
		"page=" + strconv.Itoa(math.MaxInt),
		"page=" + strconv.Itoa(math.MaxInt) + "&page_size=100",
		"page=" + strconv.Itoa(math.MaxInt/20+1) + "&page_size=20",
		"page=" + strconv.Itoa(math.MaxInt) + "0",
		"page=bad", "page=0", "page=-1",
		"page_size=0", "page_size=-1", "page_size=1001",
	} {
		t.Run(query, func(t *testing.T) {
			before := repo.calls
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/stats?"+query, nil)
			router.ServeHTTP(recorder, req)
			require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
			require.Equal(t, before, repo.calls, "invalid pagination must not reach repository")
		})
	}
}

func TestUpstreamAFixesOpsPaginationAcceptsOrdinaryPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &upstreamAPaginationOpsRepo{}
	svc := service.NewOpsService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.GET("/stats", NewOpsHandler(svc).GetDashboardOpenAITokenStats)
	for _, tc := range []struct {
		query            string
		page, size, topN int
	}{
		{"", 1, 20, 0},
		{"page=3&page_size=20", 3, 20, 0},
		{"page=" + strconv.Itoa(math.MaxInt/20) + "&page_size=20", math.MaxInt / 20, 20, 0},
		{"page=" + strconv.Itoa(math.MaxInt) + "&page_size=1", math.MaxInt, 1, 0},
		{"top_n=5", 0, 0, 5},
	} {
		t.Run(tc.query, func(t *testing.T) {
			before := repo.calls
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/stats?"+tc.query, nil))
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			require.Equal(t, before+1, repo.calls)
			require.Equal(t, tc.page, repo.filter.Page)
			require.Equal(t, tc.size, repo.filter.PageSize)
			require.Equal(t, tc.topN, repo.filter.TopN)
		})
	}
}
