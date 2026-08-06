//go:build unit

package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

type openAIQuotaHandlerAccountRepoStub struct {
	account *service.Account
}

func (r *openAIQuotaHandlerAccountRepoStub) GetByID(ctx context.Context, id int64) (*service.Account, error) {
	return r.account, nil
}

func newOpenAIQuotaHandlerForTest(t *testing.T, upstream http.HandlerFunc, account *service.Account) *OpenAIOAuthHandler {
	t.Helper()
	server := httptest.NewServer(upstream)
	t.Cleanup(server.Close)

	oldUsageURL := serviceTestSetOpenAIQuotaURL(t, server.URL)
	t.Cleanup(oldUsageURL)

	tokenProvider := service.NewOpenAITokenProvider(nil, nil, nil)
	quotaService := service.NewOpenAIQuotaServiceForTest(
		&openAIQuotaHandlerAccountRepoStub{account: account},
		nil,
		tokenProvider,
		func(proxyURL string) (*req.Client, error) {
			return req.C().SetTimeout(time.Second), nil
		},
	)
	return NewOpenAIOAuthHandler(nil, nil, quotaService, nil)
}

func serviceTestSetOpenAIQuotaURL(t *testing.T, serverURL string) func() {
	t.Helper()
	return service.SetOpenAIQuotaURLsForTest(serverURL)
}

func TestOpenAIOAuthHandlerQueryQuotaSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	account := &service.Account{
		ID:       12,
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token":       "token",
			"expires_at":         time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
			"chatgpt_account_id": "acct_12",
		},
	}
	handler := newOpenAIQuotaHandlerForTest(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/backend-api/wham/usage", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rate_limit_reset_credits":{"available_count":3}}`))
	}, account)

	router := gin.New()
	router.GET("/admin/openai/accounts/:id/quota", handler.QueryQuota)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/openai/accounts/12/quota", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			RateLimitResetCredits struct {
				AvailableCount int `json:"available_count"`
			} `json:"rate_limit_reset_credits"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, 3, body.Data.RateLimitResetCredits.AvailableCount)
}

func TestOpenAIOAuthHandlerResetQuotaInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewOpenAIOAuthHandler(nil, nil, nil, nil)
	router := gin.New()
	router.POST("/admin/openai/accounts/:id/reset-quota", handler.ResetQuota)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/admin/openai/accounts/not-a-number/reset-quota", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "Invalid account ID")
}

func TestOpenAIOAuthHandlerQueryQuotaServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	account := &service.Account{
		ID:          13,
		Platform:    service.PlatformOpenAI,
		Type:        service.AccountTypeOAuth,
		Credentials: map[string]any{"access_token": "token"},
	}
	handler := newOpenAIQuotaHandlerForTest(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("upstream should not be called")
	}, account)

	router := gin.New()
	router.GET("/admin/openai/accounts/:id/quota", handler.QueryQuota)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/openai/accounts/13/quota", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "OPENAI_QUOTA_MISSING_ACCOUNT_ID")
}
