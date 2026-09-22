package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type v027DeepSeekUpstream struct {
	resp     *http.Response
	lastBody []byte
}

func (u *v027DeepSeekUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	if req != nil && req.Body != nil {
		u.lastBody, _ = io.ReadAll(req.Body)
	}
	return u.resp, nil
}

func (u *v027DeepSeekUpstream) DoWithTLS(req *http.Request, proxyURL string, accountID int64, accountConcurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return u.Do(req, proxyURL, accountID, accountConcurrency)
}

func v027DeepSeekTestAccount(baseURL string) *Account {
	return &Account{
		ID:          927,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{"api_key": "test-key", "base_url": baseURL},
	}
}

func v027DeepSeekTestConfig() *config.Config {
	return &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}}
}

func v027AssistantReasoning(body []byte) string {
	for _, message := range gjson.GetBytes(body, "messages").Array() {
		if message.Get("role").String() == "assistant" {
			return message.Get("reasoning_content").String()
		}
	}
	return ""
}

func TestV027DeepSeekReasoning_FallbackNonStreamingPreservesPlaintextAndStrictHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name     string
		baseURL  string
		platform string
		expected string
	}{
		{name: "exact deepseek host", baseURL: "https://api.deepseek.com/v1", platform: PlatformOpenAI, expected: "visible reasoning"},
		{name: "userinfo lookalike host", baseURL: "https://api.deepseek.com@evil.example/v1", platform: PlatformOpenAI, expected: ""},
		{name: "suffix lookalike host", baseURL: "https://api.deepseek.com.example/v1", platform: PlatformOpenAI, expected: ""},
		{name: "deepseek platform custom host", baseURL: "https://custom.example/v1", platform: PlatformDeepseek, expected: "visible reasoning"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := []byte(`{"model":"deepseek-reasoner","input":[{"type":"reasoning","summary":[{"type":"summary_text","text":"visible reasoning"}]},{"type":"message","role":"assistant","content":"prior answer"},{"type":"message","role":"user","content":"continue"}]}`)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
			upstream := &v027DeepSeekUpstream{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": {"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"id":"chatcmpl","object":"chat.completion","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`)),
			}}
			svc := &OpenAIGatewayService{cfg: v027DeepSeekTestConfig(), httpUpstream: upstream}
			account := v027DeepSeekTestAccount(tc.baseURL)
			account.Platform = tc.platform

			_, err := svc.forwardResponsesViaRawChatCompletions(context.Background(), c, account, body)

			require.NoError(t, err)
			require.Equal(t, tc.expected, v027AssistantReasoning(upstream.lastBody))
		})
	}
}

func TestV027DeepSeekReasoning_FallbackUsesPlaceholderForEncryptedOnlyAndStreaming(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"deepseek-reasoner","stream":true,"input":[{"type":"reasoning","encrypted_content":"ciphertext"},{"type":"message","role":"assistant","content":"prior answer"},{"type":"message","role":"user","content":"continue"}]}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	upstream := &v027DeepSeekUpstream{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader("data: {\"id\":\"chatcmpl\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"index\":0,\"delta\":{\"role\":\"assistant\"},\"finish_reason\":null}]}\n\ndata: {\"id\":\"chatcmpl\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n")),
	}}
	svc := &OpenAIGatewayService{cfg: v027DeepSeekTestConfig(), httpUpstream: upstream}
	account := v027DeepSeekTestAccount("https://api.deepseek.com/v1")

	result, err := svc.forwardResponsesViaRawChatCompletions(context.Background(), c, account, body)

	require.NoError(t, err)
	require.True(t, result.Stream)
	require.Equal(t, " ", v027AssistantReasoning(upstream.lastBody))
	require.Contains(t, rec.Body.String(), "data: [DONE]")
}
