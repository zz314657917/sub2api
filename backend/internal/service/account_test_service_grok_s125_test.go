package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAccountTestServiceGrokOAuthPaymentRequiredTemporarilyUnschedulesAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	account := &Account{
		ID:          125,
		Platform:    PlatformGrok,
		Type:        AccountTypeOAuth,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token": "grok-access-token",
			"expires_at":   time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
		},
	}
	repo := &grokS111AccountRepo{}
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusPaymentRequired,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"code":"personal-team-blocked:spending-limit"}`)),
	}}
	svc := &AccountTestService{
		accountRepo:       repo,
		grokTokenProvider: NewGrokTokenProvider(repo, nil),
		httpUpstream:      upstream,
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/125/test", nil)
	before := time.Now()

	err := svc.testGrokAccountConnection(c, account, "grok")

	require.Error(t, err)
	require.Equal(t, 1, repo.tempUnschedCalls)
	require.Equal(t, account.ID, repo.accountID)
	require.Equal(t, "grok payment required", repo.reason)
	require.WithinDuration(t, before.Add(30*time.Minute), repo.until, time.Second)
	require.Contains(t, recorder.Body.String(), `"type":"error"`)
	require.Contains(t, recorder.Body.String(), "Grok Responses API returned 402")
}
