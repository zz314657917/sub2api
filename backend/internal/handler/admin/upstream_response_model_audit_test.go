package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAdminUsageHandlersPropagateUpstreamModelMismatch(t *testing.T) {
	for _, tc := range []struct {
		name, path string
		wantStatus int
		want       *bool
	}{
		{"list true", "/admin/usage?upstream_model_mismatch=true", http.StatusOK, boolPtr(true)},
		{"list false", "/admin/usage?upstream_model_mismatch=false", http.StatusOK, boolPtr(false)},
		{"list unset", "/admin/usage", http.StatusOK, nil},
		{"list invalid", "/admin/usage?upstream_model_mismatch=maybe", http.StatusBadRequest, nil},
		{"stats true", "/admin/usage/stats?upstream_model_mismatch=true", http.StatusOK, boolPtr(true)},
		{"stats false", "/admin/usage/stats?upstream_model_mismatch=false", http.StatusOK, boolPtr(false)},
		{"stats unset", "/admin/usage/stats", http.StatusOK, nil},
		{"stats invalid", "/admin/usage/stats?upstream_model_mismatch=maybe", http.StatusBadRequest, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &adminUsageRepoCapture{}
			router := newAdminUsageRequestTypeTestRouter(repo)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			require.Equal(t, tc.wantStatus, rec.Code)
			if tc.wantStatus != http.StatusOK {
				return
			}
			got := repo.listFilters.UpstreamModelMismatch
			if strings.HasPrefix(tc.path, "/admin/usage/stats") {
				got = repo.statsFilters.UpstreamModelMismatch
			}
			if tc.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *tc.want, *got)
		})
	}
}

type upstreamResponseModelAuditUsageRepo struct{ service.UsageLogRepository }

func (r *upstreamResponseModelAuditUsageRepo) ListWithFilters(context.Context, pagination.PaginationParams, usagestats.UsageLogFilters) ([]service.UsageLog, *pagination.PaginationResult, error) {
	responseModel := "gpt-5-upstream"
	mismatch := true
	return []service.UsageLog{{Model: "gpt-5", RequestedModel: "gpt-5", UpstreamResponseModel: &responseModel, UpstreamModelMismatch: &mismatch}}, &pagination.PaginationResult{Total: 1, Page: 1, PageSize: 20, Pages: 1}, nil
}

func TestAdminUsageResponseModelAuditFieldsAreRawAndTriState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &upstreamResponseModelAuditUsageRepo{}
	handler := NewUsageHandler(service.NewUsageService(repo, nil, nil, nil), nil, nil, nil)
	router := gin.New()
	router.GET("/admin/usage", handler.List)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/admin/usage", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"upstream_response_model":"gpt-5-upstream"`)
	require.Contains(t, rec.Body.String(), `"upstream_model_mismatch":true`)
}
