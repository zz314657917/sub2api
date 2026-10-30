package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const modelUnavailable404 = `{"error":{"message":"Model 'gpt-6-luna' is not supported by this account","type":"invalid_request_error"}}`
const modelUnavailablePlan = `{"detail":"The 'gpt-5.4' model is not supported when using Codex with a ChatGPT account."}`

func modelUnavailableContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/v1/responses", nil)
	SetOpenAIClientTransport(c, OpenAIClientTransportHTTP)
	return c, rec
}

func TestModelUnavailableAuthoritativeFields(t *testing.T) {
	a := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	tests := []struct {
		body string
		want bool
	}{
		{modelUnavailable404, true},
		{`{"error":{"code":"model_not_found"}}`, true},
		{`{"code":"unsupported_model"}`, true},
		{`{"error":{"message":"unknown model: gpt-6-luna"}}`, true},
		{`{"error":{"message":"unknown model"}}`, true},
		{`{"error":{"message":"unknown provider for model: gpt-6-luna"}}`, true},
		{`{"detail":"The model 'gpt-6-luna' does not exist"}`, true},
		{`{"message":"Model 'gpt-6-luna' is not available for this account"}`, true},
		{`{"error":{"message":"route not found"},"debug":{"model":"x","message":"model not found"}}`, false},
		{`{"request":{"model":"x","message":"model not found"}}`, false},
		{`{"metadata":{"message":"unknown model: x"}}`, false},
		{`{"error":{"code":"invalid_value","message":"model not found"}}`, false},
		{`{"code":"context_length_exceeded","error":{"code":"model_not_found"}}`, false},
		{`{"error":{"code":"invalid_value","message":"{\"error\":{\"code\":\"model_not_found\"}}"}}`, false},
		{`{"error":{"message":"model does not support response_format"}}`, false},
		{`{"error":{"message":"Model 'gpt-6-luna' is not supported for streaming"}}`, false},
		{`{"error":{"message":"Model 'gpt-6-luna' is not supported for tools"}}`, false},
		{`{"error":{"message":"The 'gpt-6-luna' model is not supported for streaming"}}`, false},
		{`{"error":{"message":"The 'gpt-6-luna' model is not supported for tools"}}`, false},
		{`{"error":{"message":"Model 'gpt-6-luna' is not supported for schema"}}`, false},
		{`{"error":{"message":"Model 'gpt-6-luna' is not supported for parameter temperature"}}`, false},
		{`{"error":{"message":"Model 'gpt-6-luna' is not supported for tools on this account"}}`, false},
		{`{"error":{"message":"Model 'gpt-6-luna' is not supported for this account"}}`, true},
		{`{"error":{"message":"Model 'gpt-6-luna' is not supported for your provider"}}`, true},
		{`{"error":{"message":"The 'gpt-6-luna' model is not supported for the current plan."}}`, true},
		{`{"error":{"message":"Model 'gpt-6-luna' is not supported for your subscription"}}`, true},
		{`{"error":{"code":"invalid_value","message":"unknown provider for model: gpt-6-luna"}}`, false},
		{`{"error":{"code":"context_length_exceeded","message":"unknown model"}}`, false},
		{`{"error":{"message":"route not found"},"debug":{"message":"unknown provider for model: gpt-6-luna"}}`, false},
		{`{"request":{"message":"unknown model"}}`, false},
		{`{"error":{"message":"Unsupported parameter: model"}}`, false},
		{`{"error":{"message":"route not found for model x"}}`, false},
		{`{"message":"maximum context length exceeded for model x"}`, false},
		{`{"error":{"message":"{\"request\":{\"message\":\"model not found\"}}"}}`, false},
		{`["model not found"]`, false},
		{`{"error": "model not found"`, false},
		{"Model 'gpt-6-luna' is not supported by this account", true},
		{"route not found", false},
	}
	for i, tt := range tests {
		for _, status := range []int{400, 404} {
			require.Equal(t, tt.want, openAIModelUnavailableError(a, status, []byte(tt.body)), "case=%d status=%d", i, status)
		}
		require.False(t, openAIModelUnavailableError(a, 500, []byte(tt.body)))
	}
	for _, typ := range []string{AccountTypeAPIKey, AccountTypeOAuth, AccountTypeSetupToken} {
		a.Type = typ
		require.Equal(t, typ == AccountTypeOAuth, openAIModelUnavailableError(a, 400, []byte(modelUnavailablePlan)))
		require.False(t, openAIModelUnavailableError(a, 404, []byte(modelUnavailablePlan)))
	}
}

func TestModelUnavailableHTTPScopesAndReplayBarrier(t *testing.T) {
	for _, scope := range []string{"http", "ws", "unknown-transport", "anthropic", "grok", "protocol", "image", "committed", "semantic", "prelude"} {
		t.Run(scope, func(t *testing.T) {
			c, rec := modelUnavailableContext()
			a := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
			switch scope {
			case "ws":
				SetOpenAIClientTransport(c, OpenAIClientTransportWS)
			case "unknown-transport":
				c.Set(openAIClientTransportContextKey, "unexpected")
			case "anthropic":
				a.Platform = PlatformAnthropic
			case "grok":
				a.Platform = PlatformGrok
			case "protocol":
				SetActualOpenAIUpstreamEndpoint(c, "/v1/messages")
			case "image":
				c.Request = httptest.NewRequest("POST", "/v1/images/generations", nil)
			case "committed":
				_, _ = c.Writer.WriteString("visible")
			case "semantic":
				MarkResponseCommitted(c)
			case "prelude":
				c.Writer = &streamBillingOpenAIAttemptWriter{ResponseWriter: c.Writer, size: 12}
			}
			err := (&OpenAIGatewayService{}).failoverOpenAIModelUnavailable(context.Background(), c, a,
				&http.Response{StatusCode: 404, Header: http.Header{}}, []byte(modelUnavailable404), "gpt-6-luna")
			var f *UpstreamFailoverError
			if scope == "http" || scope == "prelude" {
				require.ErrorAs(t, err, &f)
				require.False(t, f.RetryableOnSameAccount)
				require.False(t, f.RequestScopedTransient)
			} else if scope == "committed" || scope == "semantic" {
				require.Error(t, err)
				require.False(t, errors.As(err, &f))
			} else {
				require.NoError(t, err)
			}
			require.NotContains(t, rec.Body.String(), "model")
		})
	}
}

func TestModelUnavailableErrorOwnersAndPolicies(t *testing.T) {
	for _, owner := range []string{"responses", "compat", "passthrough"} {
		for _, policy := range []string{"default", "included", "excluded", "pool", "pool-included", "pool-excluded", "write-failure"} {
			t.Run(owner+"/"+policy, func(t *testing.T) {
				c, rec := modelUnavailableContext()
				a := modelUnavailableAccount(1)
				if strings.Contains(policy, "pool") {
					a.Credentials["pool_mode"] = true
					a.Credentials["pool_mode_retry_status_codes"] = []any{float64(400), float64(404)}
				}
				if strings.Contains(policy, "included") || strings.Contains(policy, "excluded") {
					a.Credentials["custom_error_codes_enabled"] = true
					status := 404
					if strings.Contains(policy, "excluded") {
						status = 429
					}
					a.Credentials["custom_error_codes"] = []any{float64(status)}
				}
				repo := &modelUnavailableRepo{}
				if policy == "write-failure" {
					repo.writeErr = errors.New("unavailable")
				}
				s := &OpenAIGatewayService{rateLimitService: &RateLimitService{accountRepo: repo}}
				resp := &http.Response{StatusCode: 404, Header: http.Header{"X-Request-Id": {"reject-a"}},
					Body: io.NopCloser(strings.NewReader(modelUnavailable404))}
				var err error
				switch owner {
				case "responses":
					_, err = s.handleErrorResponse(context.Background(), resp, c, &a, []byte(`{"model":"gpt-6-luna"}`), "gpt-6-luna")
				case "compat":
					_, err = s.handleChatCompletionsErrorResponse(resp, c, &a, "gpt-6-luna")
				case "passthrough":
					err = s.handleErrorResponsePassthrough(context.Background(), resp, c, &a, []byte(`{"model":"gpt-6-luna"}`))
				}
				var f *UpstreamFailoverError
				require.ErrorAs(t, err, &f)
				require.False(t, f.RetryableOnSameAccount)
				require.Zero(t, f.SameAccountRetryLimit)
				require.Equal(t, "reject-a", f.ResponseHeaders.Get("X-Request-Id"))
				require.False(t, c.Writer.Written())
				require.Empty(t, rec.Body.String())
				want := 1
				if policy == "pool" || strings.Contains(policy, "excluded") {
					want = 0
				}
				require.Len(t, repo.keys, want)
			})
		}
	}
}

func TestModelUnavailableJSONEchoDoesNotReachLegacyCooldown(t *testing.T) {
	c, _ := modelUnavailableContext()
	repo := &modelUnavailableRepo{}
	s := &OpenAIGatewayService{rateLimitService: &RateLimitService{accountRepo: repo}}
	a := modelUnavailableAccount(1)
	for _, payload := range []any{
		map[string]any{"error": map[string]any{"message": "route not found"}, "debug": map[string]any{"message": "model not found"}},
		map[string]any{"error": map[string]any{"code": "invalid_value", "message": "model not found"}},
	} {
		body, err := json.Marshal(payload)
		require.NoError(t, err)
		require.False(t, s.handleOpenAIHTTPAccountUpstreamError(context.Background(), c, &a, 404, nil, body, "x"))
	}
	require.Empty(t, repo.keys)
}
