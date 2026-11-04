package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUpstreamFiveOpenAIImageProbeOutboundAndSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, model := range []string{"gpt-image-1", "gpt-image-1-mini", "gpt-image-1.5", "gpt-image-2", "gpt-image-2.5"} {
		for _, baseURL := range []string{"", "https://image-upstream.example/v1"} {
			t.Run(model+"/"+baseURL, func(t *testing.T) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/1/test", nil)
				upstream := &httpUpstreamRecorder{resp: &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"data":[{"b64_json":"aGVsbG8=","revised_prompt":"revised cat"},{"b64_json":"d29ybGQ="}]}`)),
				}}
				svc := &AccountTestService{httpUpstream: upstream, cfg: &config.Config{}}
				account := &Account{ID: 54, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
					Credentials: map[string]any{"api_key": "test-image-key", "base_url": baseURL}}
				err := svc.testOpenAIAccountConnection(c, account, model, "draw a cat", AccountTestModeDefault)
				require.NoError(t, err)
				require.NotNil(t, upstream.lastReq)
				url := "https://api.openai.com/v1/images/generations"
				if baseURL != "" {
					url = baseURL + "/images/generations"
				}
				require.Equal(t, url, upstream.lastReq.URL.String())
				require.Equal(t, http.MethodPost, upstream.lastReq.Method)
				require.Equal(t, "Bearer test-image-key", upstream.lastReq.Header.Get("Authorization"))
				require.Equal(t, "application/json", upstream.lastReq.Header.Get("Content-Type"))
				var payload map[string]any
				require.NoError(t, json.Unmarshal(upstream.lastBody, &payload))
				require.Equal(t, map[string]any{"model": model, "prompt": "draw a cat", "n": float64(1)}, payload)
				require.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
				require.Contains(t, w.Body.String(), `"type":"test_start"`)
				require.Contains(t, w.Body.String(), "revised cat")
				require.Contains(t, w.Body.String(), "data:image/png;base64,aGVsbG8=")
				require.Contains(t, w.Body.String(), "data:image/png;base64,d29ybGQ=")
				require.Contains(t, w.Body.String(), `"success":true`)
				require.NotContains(t, w.Body.String(), "test-image-key")
			})
		}
	}
}

func TestUpstreamFiveOpenAIImageProbeErrorsRemainSSE(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body, message string
		status              int
	}{
		{"upstream", `{"error":{"message":"image rejected"}}`, "API returned 400", 400},
		{"malformed_json", `not-json`, "Failed to parse response", 200},
		{"no_images", `{"data":[]}`, "No images returned from API", 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/test", nil)
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: tc.status, Header: http.Header{"Content-Type": []string{"application/json"}},
				Body: io.NopCloser(strings.NewReader(tc.body)),
			}}
			svc := &AccountTestService{httpUpstream: upstream, cfg: &config.Config{}}
			account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "test"}}
			err := svc.testOpenAIImageAPIKey(c, context.Background(), account, "gpt-image-2", "cat")
			require.ErrorContains(t, err, tc.message)
			require.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
			require.Contains(t, w.Body.String(), `"type":"error"`)
			require.Contains(t, w.Body.String(), tc.message)
			require.NotContains(t, w.Body.String(), `"success":true`)
		})
	}
}
