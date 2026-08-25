package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIGatewayService_ForwardRetriesRejectedStatusFieldWithFakeHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"gpt-5","stream":false,"input":[{"type":"tool_search_output","status":"completed","call_id":"first"},{"type":"message","status":"completed","content":"keep"},{"type":"tool_search_output","status":"completed","call_id":"second"}]}`)
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		openAIRejectedFieldRetryTestResponse(http.StatusBadRequest, `{"error":{"code":"unknown_parameter","message":"Unknown parameter: 'input[2].status'.","param":"input[2].status"}}`),
		openAIRejectedFieldRetryTestResponse(http.StatusOK, `{"output":[],"usage":{"input_tokens":1,"output_tokens":1,"input_tokens_details":{"cached_tokens":0}}}`),
	}}

	result, err := openAIRejectedFieldRetryTestService(upstream).Forward(context.Background(), openAIRejectedFieldRetryTestContext(body), openAIRejectedFieldRetryTestAccount(), body)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, upstream.bodies, 2)
	require.Equal(t, "completed", gjson.GetBytes(upstream.bodies[0], "input.0.status").String())
	require.False(t, gjson.GetBytes(upstream.bodies[1], "input.0.status").Exists())
	require.Equal(t, "completed", gjson.GetBytes(upstream.bodies[1], "input.1.status").String())
	require.False(t, gjson.GetBytes(upstream.bodies[1], "input.2.status").Exists())
}

func openAIRejectedFieldRetryTestService(upstream *httpUpstreamRecorder) *OpenAIGatewayService {
	return &OpenAIGatewayService{
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
		httpUpstream: upstream,
	}
}

func openAIRejectedFieldRetryTestContext(body []byte) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("User-Agent", "curl/8.0")
	return c
}

func openAIRejectedFieldRetryTestAccount() *Account {
	return &Account{
		ID:          5257,
		Name:        "responses-rejected-field-test",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://compat.example"},
		Extra: map[string]any{
			openai_compat.ExtraKeyResponsesMode:      string(openai_compat.ResponsesSupportModeAuto),
			openai_compat.ExtraKeyResponsesSupported: true,
		},
		Status:      StatusActive,
		Schedulable: true,
	}
}

func openAIRejectedFieldRetryTestResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
	}
}
