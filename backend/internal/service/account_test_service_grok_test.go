//go:build unit

package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/xai"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestAccountTestService_TestAccountConnection_GrokUsesXAIResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	account := &Account{
		ID:          13,
		Name:        "grok-oauth",
		Platform:    PlatformGrok,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: map[string]any{
			"access_token": "grok-access-token",
			"expires_at":   time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
			"model_mapping": map[string]any{
				"grok": "grok-4.3",
			},
		},
	}
	repo := &mockAccountRepoForGemini{
		accountsByID: map[int64]*Account{account.ID: account},
	}
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(
			"data: {\"type\":\"response.output_text.delta\",\"delta\":\"ok\"}\n\n" +
				"data: {\"type\":\"response.completed\"}\n\n",
		)),
	}}
	svc := &AccountTestService{
		accountRepo:       repo,
		grokTokenProvider: NewGrokTokenProvider(repo, nil),
		httpUpstream:      upstream,
	}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/13/test", nil)

	err := svc.TestAccountConnection(c, account.ID, "grok", "", AccountTestModeDefault)
	require.NoError(t, err)

	require.Equal(t, "https://api.x.ai/v1/responses", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer grok-access-token", upstream.lastReq.Header.Get("Authorization"))
	require.Equal(t, grokOfficialOAuthUserAgent, upstream.lastReq.UserAgent())
	require.Equal(t, "grok-4.3", gjson.GetBytes(upstream.lastBody, "model").String())
	require.NotContains(t, rec.Body.String(), "claude")
	require.Contains(t, rec.Body.String(), `"model":"grok-4.3"`)
	require.Contains(t, rec.Body.String(), `"type":"test_complete"`)
}

func TestAccountTestService_TestAccountConnection_GrokAPIKeyUsesXAIResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	account := &Account{
		ID:       14,
		Platform: PlatformGrok,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key": "  xai-fixture  ",
		},
	}
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(
			"data: {\"type\":\"response.output_text.delta\",\"delta\":\"ok\"}\n\n" +
				"data: {\"type\":\"response.completed\"}\n\n",
		)),
	}}
	svc := &AccountTestService{
		accountRepo:  &mockAccountRepoForGemini{accountsByID: map[int64]*Account{account.ID: account}},
		httpUpstream: upstream,
	}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/14/test", nil)

	require.NoError(t, svc.TestAccountConnection(c, account.ID, "", "", AccountTestModeDefault))
	require.NotNil(t, upstream.lastReq)
	require.Equal(t, "https://api.x.ai/v1/responses", upstream.lastReq.URL.String())
	require.Equal(t, "Bearer xai-fixture", upstream.lastReq.Header.Get("Authorization"))
	require.Empty(t, upstream.lastReq.UserAgent())
	require.Equal(t, "grok-4.3", gjson.GetBytes(upstream.lastBody, "model").String())
	require.Contains(t, rec.Header().Get("Content-Type"), "text/event-stream")
	require.Contains(t, rec.Body.String(), `"type":"test_start"`)
	require.Contains(t, rec.Body.String(), `"type":"content","text":"ok"`)
	require.Contains(t, rec.Body.String(), `"type":"test_complete","success":true`)
}

func TestAccountTestService_TestAccountConnection_GrokAPIKeyRejectsInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv(xai.EnvAllowUnsafeURLOverrides, "false")
	for _, tc := range []struct {
		name, key, baseURL, wantError string
	}{
		{"blank key", "  ", xai.DefaultBaseURL, "Grok API key is missing"},
		{"invalid host", "xai-fixture", "https://example.invalid/v1", "Invalid Grok base URL"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account := &Account{
				ID:          14,
				Platform:    PlatformGrok,
				Type:        AccountTypeAPIKey,
				Credentials: map[string]any{"api_key": tc.key, "base_url": tc.baseURL},
			}
			upstream := &httpUpstreamRecorder{}
			svc := &AccountTestService{
				accountRepo:  &mockAccountRepoForGemini{accountsByID: map[int64]*Account{account.ID: account}},
				httpUpstream: upstream,
			}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/14/test", nil)

			err := svc.TestAccountConnection(c, account.ID, "", "", AccountTestModeDefault)
			require.ErrorContains(t, err, tc.wantError)
			require.Nil(t, upstream.lastReq)
			require.Contains(t, rec.Body.String(), `"type":"error"`)
		})
	}
}
