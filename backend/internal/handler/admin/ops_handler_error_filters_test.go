package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type opsErrorRepoCapture struct {
	service.OpsRepository
	filter *service.OpsErrorLogFilter
}

func (r *opsErrorRepoCapture) ListErrorLogs(_ context.Context, filter *service.OpsErrorLogFilter) (*service.OpsErrorLogList, error) {
	copy := *filter
	r.filter = &copy
	return &service.OpsErrorLogList{Errors: []*service.OpsErrorLog{}, Page: filter.Page, PageSize: filter.PageSize}, nil
}

func newOpsErrorFilterRouter(repo *opsErrorRepoCapture) *gin.Engine {
	gin.SetMode(gin.TestMode)
	svc := service.NewOpsService(repo, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.GET("/errors", NewOpsHandler(svc).GetErrorLogs)
	return router
}

func TestOpsErrorHandler_PropagatesUsageFiltersAndSort(t *testing.T) {
	repo := &opsErrorRepoCapture{}
	router := newOpsErrorFilterRouter(repo)
	req := httptest.NewRequest(http.MethodGet, "/errors?time_range=24h&user_id=12&api_key_id=34&model=gpt-5.6&category=auth&sort_by=status_code&sort_order=asc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, repo.filter)
	require.Equal(t, int64(12), *repo.filter.UserID)
	require.Equal(t, int64(34), *repo.filter.APIKeyID)
	require.Equal(t, "gpt-5.6", repo.filter.Model)
	require.Equal(t, []string{"auth"}, repo.filter.ErrorPhasesAny)
	require.Equal(t, "status_code", repo.filter.SortBy)
	require.Equal(t, "asc", repo.filter.SortOrder)
}

func TestOpsErrorHandler_RejectsInvalidUsageIdentityFilter(t *testing.T) {
	for _, query := range []string{
		"user_id=invalid",
		"api_key_id=0",
		"category=other",
		"sort_by=id%3Bdrop+table+ops_error_logs",
		"sort_order=sideways",
	} {
		t.Run(query, func(t *testing.T) {
			repo := &opsErrorRepoCapture{}
			router := newOpsErrorFilterRouter(repo)
			req := httptest.NewRequest(http.MethodGet, "/errors?"+query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			require.Equal(t, http.StatusBadRequest, w.Code)
			require.Nil(t, repo.filter)
		})
	}
}
