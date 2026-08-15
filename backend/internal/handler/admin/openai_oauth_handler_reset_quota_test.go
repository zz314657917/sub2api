package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type openAIQuotaResetFake struct {
	order       *[]string
	cancel      context.CancelFunc
	postCtxErrs []error
	cacheErr    error
}

func (f *openAIQuotaResetFake) ResetCredit(context.Context, int64) (*service.OpenAIQuotaResetResult, error) {
	*f.order = append(*f.order, "reset")
	f.cancel()
	return &service.OpenAIQuotaResetResult{Code: "success", WindowsReset: 1}, nil
}

func (f *openAIQuotaResetFake) QueryUsage(ctx context.Context, _ int64) (*service.OpenAIQuotaUsage, error) {
	*f.order = append(*f.order, "query")
	f.postCtxErrs = append(f.postCtxErrs, ctx.Err())
	return &service.OpenAIQuotaUsage{RateLimitResetCredits: &service.OpenAIRateLimitResetCredits{
		AvailableCount: 1,
		Credits:        []service.OpenAIRateLimitResetCreditDetail{{ExpiresAt: "2026-08-07T00:00:00Z"}},
	}}, nil
}

func (f *openAIQuotaResetFake) CacheResetCreditsSnapshot(ctx context.Context, _ int64, _ *service.OpenAIRateLimitResetCredits) error {
	*f.order = append(*f.order, "cache")
	f.postCtxErrs = append(f.postCtxErrs, ctx.Err())
	return f.cacheErr
}

type openAIQuotaResetRecoveryFake struct {
	order       *[]string
	postCtxErr  error
	hasDeadline bool
}

func (f *openAIQuotaResetRecoveryFake) RecoverAccountState(ctx context.Context, _ int64, options service.AccountRecoveryOptions) (*service.SuccessfulTestRecoveryResult, error) {
	*f.order = append(*f.order, "recover")
	f.postCtxErr = ctx.Err()
	_, f.hasDeadline = ctx.Deadline()
	if !options.InvalidateToken {
		return nil, context.Canceled
	}
	return &service.SuccessfulTestRecoveryResult{ClearedRateLimit: true}, nil
}

func TestOpenAIOAuthHandlerResetQuota_PostProcessingSurvivesClientCancel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	order := make([]string, 0, 2)
	requestCtx, cancelRequest := context.WithCancel(context.Background())
	defer cancelRequest()
	quota := &openAIQuotaResetFake{order: &order, cancel: cancelRequest}
	recovery := &openAIQuotaResetRecoveryFake{order: &order}
	handler := &OpenAIOAuthHandler{
		quotaService:     quota,
		rateLimitService: recovery,
		adminService:     newStubAdminService(),
	}

	router := gin.New()
	router.POST("/admin/openai/accounts/:id/reset-quota", handler.ResetQuota)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/openai/accounts/42/reset-quota", nil).WithContext(requestCtx)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, []string{"reset", "recover"}, order)
	require.NoError(t, recovery.postCtxErr)
	require.True(t, recovery.hasDeadline)
	require.Empty(t, quota.postCtxErrs)

	var body struct {
		Code int `json:"code"`
		Data struct {
			CacheRefreshed        bool   `json:"cache_refreshed"`
			AccountStateRecovered bool   `json:"account_state_recovered"`
			WarningCode           string `json:"warning_code"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.False(t, body.Data.CacheRefreshed)
	require.True(t, body.Data.AccountStateRecovered)
	require.Empty(t, body.Data.WarningCode)
}

func TestOpenAIOAuthHandlerResetQuota_CacheFailureKeepsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	order := make([]string, 0, 2)
	quota := &openAIQuotaResetFake{
		order:    &order,
		cancel:   func() {},
		cacheErr: errors.New("cache unavailable"),
	}
	handler := &OpenAIOAuthHandler{
		quotaService:     quota,
		rateLimitService: &openAIQuotaResetRecoveryFake{order: &order},
		adminService:     newStubAdminService(),
	}

	router := gin.New()
	router.POST("/admin/openai/accounts/:id/reset-quota", handler.ResetQuota)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/admin/openai/accounts/42/reset-quota", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, []string{"reset", "recover"}, order)
	var body struct {
		Code int `json:"code"`
		Data struct {
			CacheRefreshed        bool   `json:"cache_refreshed"`
			AccountStateRecovered bool   `json:"account_state_recovered"`
			WarningCode           string `json:"warning_code"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.False(t, body.Data.CacheRefreshed)
	require.True(t, body.Data.AccountStateRecovered)
	require.Empty(t, body.Data.WarningCode)
}

func TestOpenAIOAuthHandlerRefreshQuota_PersistsSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	order := make([]string, 0, 2)
	quota := &openAIQuotaResetFake{order: &order, cancel: func() {}}
	h := &OpenAIOAuthHandler{quotaService: quota}
	router := gin.New()
	router.POST("/admin/openai/accounts/:id/quota/refresh", h.RefreshQuota)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/admin/openai/accounts/42/quota/refresh", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, []string{"query", "cache"}, order)
}
