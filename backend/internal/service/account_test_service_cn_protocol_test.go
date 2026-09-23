package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type cnProtocolTestUpstream struct {
	requests  []*http.Request
	responses []*http.Response
}

func (u *cnProtocolTestUpstream) Do(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return nil, fmt.Errorf("unexpected non-TLS request")
}

func (u *cnProtocolTestUpstream) DoWithTLS(req *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	u.requests = append(u.requests, req)
	if len(u.responses) == 0 {
		return nil, fmt.Errorf("unexpected upstream request")
	}
	resp := u.responses[0]
	u.responses = u.responses[1:]
	return resp, nil
}

func cnProtocolTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/account/test", nil).WithContext(context.Background())
	return c, recorder
}

func cnProtocolTestService(upstream *cnProtocolTestUpstream) *AccountTestService {
	return &AccountTestService{
		httpUpstream: upstream,
		cfg: &config.Config{Security: config.SecurityConfig{
			URLAllowlist: config.URLAllowlistConfig{Enabled: false},
		}},
	}
}

func cnProtocolTestResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestAccountTestService_CNChatProbeUsesProviderURLKeyAndMappedModel(t *testing.T) {
	upstream := &cnProtocolTestUpstream{responses: []*http.Response{
		cnProtocolTestResponse("data: {\"choices\":[{\"delta\":{\"content\":\"ok\"},\"finish_reason\":\"stop\"}]}\n\ndata: [DONE]\n\n"),
	}}
	service := cnProtocolTestService(upstream)
	account := cnProviderAccount(PlatformKimi, "", APIProtocolChatCompletions, "https://relay.example/v1")
	account.Credentials["model_mapping"] = map[string]any{"gpt-test": "kimi-test"}

	c, recorder := cnProtocolTestContext()
	err := service.testCNProviderChatCompletionsConnection(c, account, "gpt-test", "hello")

	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://relay.example/v1/chat/completions", upstream.requests[0].URL.String())
	require.Equal(t, "Bearer sk-test", upstream.requests[0].Header.Get("Authorization"))
	require.Contains(t, recorder.Body.String(), "ok")
}

func TestAccountTestService_CNAnthropicProbeUsesNativeDefaultWithoutBetaQuery(t *testing.T) {
	upstream := &cnProtocolTestUpstream{responses: []*http.Response{
		cnProtocolTestResponse("data: {\"type\":\"message_start\"}\n\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"text\":\"ok\"}}\n\ndata: {\"type\":\"message_stop\"}\n\n"),
	}}
	service := cnProtocolTestService(upstream)
	account := cnProviderAccount(PlatformZhipu, "", APIProtocolAnthropic, "")

	c, recorder := cnProtocolTestContext()
	err := service.testCNProviderAnthropicConnection(c, account, "claude-test")

	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://open.bigmodel.cn/api/anthropic/v1/messages", upstream.requests[0].URL.String())
	require.Empty(t, upstream.requests[0].URL.Query())
	require.Equal(t, []string{"sk-test"}, upstream.requests[0].Header["x-api-key"])
	require.Empty(t, upstream.requests[0].Header.Get("Authorization"))
	require.Contains(t, recorder.Body.String(), "ok")
}

func TestAccountTestService_CNPelicanProtocolsPreserveExactPrompt(t *testing.T) {
	const prompt = "  生成鹈鹕 HTML\n保留换行  "
	for _, tc := range []struct {
		name     string
		platform string
		protocol string
		model    string
		body     string
		assert   func(*testing.T, map[string]any)
	}{
		{name: "native anthropic", platform: PlatformZhipu, protocol: APIProtocolAnthropic, model: "claude-test", body: "data: {\"type\":\"message_start\"}\n\ndata: {\"type\":\"message_stop\"}\n\n", assert: func(t *testing.T, body map[string]any) {
			messages := body["messages"].([]any)
			content := messages[0].(map[string]any)["content"].([]any)
			require.Equal(t, prompt, content[0].(map[string]any)["text"])
		}},
		{name: "deepseek responses", platform: PlatformDeepseek, protocol: APIProtocolResponses, model: "deepseek-test", body: "data: {\"type\":\"response.completed\"}\n\n", assert: func(t *testing.T, body map[string]any) {
			input := body["input"].([]any)
			content := input[0].(map[string]any)["content"].([]any)
			require.Equal(t, prompt, content[0].(map[string]any)["text"])
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			upstream := &cnProtocolTestUpstream{responses: []*http.Response{cnProtocolTestResponse(tc.body)}}
			service := cnProtocolTestService(upstream)
			baseURL := "https://relay.example/v1"
			if tc.protocol == APIProtocolAnthropic {
				baseURL = ""
			}
			account := cnProviderAccount(tc.platform, "", tc.protocol, baseURL)
			c, _ := cnProtocolTestContext()
			c.Set("pelican_test", true)
			c.Set("pelican_prompt", prompt)
			var err error
			if tc.protocol == APIProtocolAnthropic {
				err = service.testCNProviderAnthropicConnection(c, account, tc.model)
			} else {
				err = service.testCNProviderResponsesConnection(c, account, tc.model)
			}
			require.NoError(t, err)
			require.Len(t, upstream.requests, 1)
			raw, readErr := io.ReadAll(upstream.requests[0].Body)
			require.NoError(t, readErr)
			var body map[string]any
			require.NoError(t, json.Unmarshal(raw, &body))
			tc.assert(t, body)
		})
	}
}

func TestAccountTestService_CNAnthropicProbeRejectsOpenAIShapedURLBeforeOutbound(t *testing.T) {
	upstream := &cnProtocolTestUpstream{}
	service := cnProtocolTestService(upstream)
	account := cnProviderAccount(PlatformZhipu, "", APIProtocolAnthropic, "https://relay.example/v1")

	c, recorder := cnProtocolTestContext()
	err := service.testCNProviderAnthropicConnection(c, account, "claude-test")

	require.Error(t, err)
	require.Empty(t, upstream.requests)
	require.Contains(t, recorder.Body.String(), "looks like an OpenAI-compatible endpoint")
}

func TestAccountTestService_DeepSeekResponsesProbeIgnoresStaleCapabilityMetadata(t *testing.T) {
	upstream := &cnProtocolTestUpstream{responses: []*http.Response{
		cnProtocolTestResponse("data: {\"type\":\"response.completed\"}\n\n"),
	}}
	service := cnProtocolTestService(upstream)
	account := cnProviderAccount(PlatformDeepseek, "", APIProtocolResponses, "https://relay.example/v1")
	account.Extra = map[string]any{"openai_responses_supported": false}

	c, _ := cnProtocolTestContext()
	err := service.testCNProviderResponsesConnection(c, account, "deepseek-test")

	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://relay.example/v1/responses", upstream.requests[0].URL.String())
	require.Equal(t, "Bearer sk-test", upstream.requests[0].Header.Get("Authorization"))
	require.NotContains(t, upstream.requests[0].URL.String(), "chat/completions")
}
