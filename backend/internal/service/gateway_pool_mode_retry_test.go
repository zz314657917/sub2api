package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGatewayCompatPoolModeRetry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		path       string
		body       []byte
		poolMode   bool
		retryCodes []any
		expect     bool
		call       func(*GatewayService, context.Context, *gin.Context, *Account, []byte) (*ForwardResult, error)
	}{
		{
			name:     "chat completions pool mode uses same account retry",
			path:     "/v1/chat/completions",
			body:     []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hello"}]}`),
			poolMode: true,
			expect:   true,
			call: func(svc *GatewayService, ctx context.Context, c *gin.Context, account *Account, body []byte) (*ForwardResult, error) {
				return svc.ForwardAsChatCompletions(ctx, c, account, body, nil)
			},
		},
		{
			name:     "responses pool mode uses same account retry",
			path:     "/v1/responses",
			body:     []byte(`{"model":"claude-sonnet-4-5","input":"hello"}`),
			poolMode: true,
			expect:   true,
			call: func(svc *GatewayService, ctx context.Context, c *gin.Context, account *Account, body []byte) (*ForwardResult, error) {
				return svc.ForwardAsResponses(ctx, c, account, body, nil)
			},
		},
		{
			name:   "chat completions non pool account does not retry",
			path:   "/v1/chat/completions",
			body:   []byte(`{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"hello"}]}`),
			expect: false,
			call: func(svc *GatewayService, ctx context.Context, c *gin.Context, account *Account, body []byte) (*ForwardResult, error) {
				return svc.ForwardAsChatCompletions(ctx, c, account, body, nil)
			},
		},
		{
			name:       "responses empty retry code list does not retry",
			path:       "/v1/responses",
			body:       []byte(`{"model":"claude-sonnet-4-5","input":"hello"}`),
			poolMode:   true,
			retryCodes: []any{},
			expect:     false,
			call: func(svc *GatewayService, ctx context.Context, c *gin.Context, account *Account, body []byte) (*ForwardResult, error) {
				return svc.ForwardAsResponses(ctx, c, account, body, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstream := &queuedHTTPUpstreamStub{responses: []*http.Response{{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"X-Request-Id": []string{"pool-429"}},
				Body:       io.NopCloser(http.NoBody),
			}}}
			svc := &GatewayService{
				cfg:                 &config.Config{},
				httpUpstream:        upstream,
				tlsFPProfileService: &TLSFingerprintProfileService{},
			}
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodPost, tt.path, nil)
			credentials := map[string]any{
				"api_key":   "test-key",
				"pool_mode": tt.poolMode,
			}
			if tt.retryCodes != nil {
				credentials["pool_mode_retry_status_codes"] = tt.retryCodes
			}
			account := &Account{
				ID:          1,
				Name:        "pool-account",
				Platform:    PlatformAnthropic,
				Type:        AccountTypeAPIKey,
				Credentials: credentials,
			}

			result, err := tt.call(svc, context.Background(), c, account, tt.body)
			require.Nil(t, result)
			var failoverErr *UpstreamFailoverError
			require.ErrorAs(t, err, &failoverErr)
			require.Equal(t, http.StatusTooManyRequests, failoverErr.StatusCode)
			require.Equal(t, tt.expect, failoverErr.RetryableOnSameAccount)
			require.Equal(t, 1, upstream.callCount)
			require.Empty(t, recorder.Body.String())
		})
	}
}
