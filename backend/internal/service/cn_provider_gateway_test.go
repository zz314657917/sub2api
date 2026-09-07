package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func newCNNativeGatewayTestService(upstream HTTPUpstream) *OpenAIGatewayService {
	cfg := cnAllowlist("127.0.0.1")
	cfg.Security.URLAllowlist.AllowPrivateHosts = true
	cfg.Gateway.StreamDataIntervalTimeout = 5
	cfg.Gateway.MaxLineSize = defaultMaxLineSize
	return &OpenAIGatewayService{cfg: cfg, httpUpstream: upstream}
}

func nativeAnthropicTestAccount(platform string) *Account {
	account := cnAccount(1, platform, AccountModePayG, "https://127.0.0.1")
	account.Credentials["api_protocol"] = APIProtocolAnthropic
	return account
}

func TestCNProviderGatewayPlatformIsolation(t *testing.T) {
	for _, platform := range []string{PlatformKimi, PlatformZhipu, PlatformDeepseek} {
		require.Equal(t, platform, normalizeOpenAICompatiblePlatform(platform))
		require.Equal(t, platform, NormalizeOpenAICompatiblePlatform(platform))
		require.True(t, isOpenAICompatibleAccount(&Account{Platform: platform}))
	}
	require.Equal(t, PlatformOpenAI, normalizeOpenAICompatiblePlatform("unknown"))
}

func TestCNProviderAnthropicNativeMessagesPassthrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := newCNNativeGatewayTestService(nil)
	account := nativeAnthropicTestAccount(PlatformZhipu)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	c.Request.Header.Set("anthropic-beta", "prompt-caching-2024-07-31")
	c.Request.Header.Set("user-agent", "test-client")
	body := []byte(`{"model":"glm-4.7","messages":[{"role":"user","content":"hi"}]}`)
	req, forwardedBody, err := svc.buildNativeAnthropicUpstreamRequest(context.Background(), c, account, body, "sk-test", "https://127.0.0.1/v1/messages")
	require.NoError(t, err)
	require.Equal(t, "https://127.0.0.1/v1/messages", req.URL.String())
	require.Equal(t, "sk-test", getHeaderRaw(req.Header, "x-api-key"))
	require.Empty(t, getHeaderRaw(req.Header, "authorization"))
	require.Equal(t, string(body), string(forwardedBody))
	require.Equal(t, "test-client", req.Header.Get("User-Agent"))
}

func TestNormalizeGLMOpenAIReasoningEffortGLM53(t *testing.T) {
	input := []byte(`{"model":"glm-5.3","reasoning_effort":"low","messages":[]}`)
	got, applied := NormalizeGLMOpenAIReasoningEffort(input, "glm-5.3")
	require.False(t, applied, "GLM-5.3 low is already in the native OpenAI scale")
	require.Equal(t, string(input), string(got))
}

func TestNativeAnthropicPassthroughNormalizesGLM53Thinking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name       string
		stream     bool
		preference string
		wantEffort string
	}{
		{name: "low buffered", preference: `"thinking":{"type":"low"},`, wantEffort: "low"},
		{name: "enabled streaming", stream: true, preference: `"thinking":{"type":"enabled"},`, wantEffort: "high"},
		{name: "adaptive streaming", stream: true, preference: `"thinking":{"type":"adaptive"},`, wantEffort: "high"},
		{name: "xhigh buffered", preference: `"output_config":{"effort":"xhigh"},`, wantEffort: "max"},
		{name: "output effort wins", preference: `"thinking":{"type":"adaptive"},"output_config":{"effort":"low"},`, wantEffort: "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := "false"
			responseBody := `{"id":"msg_1","type":"message","role":"assistant","content":[],"usage":{"input_tokens":1,"output_tokens":1}}`
			contentType := "application/json"
			if tt.stream {
				stream = "true"
				responseBody = miniAnthropicSSEStream()
				contentType = "text/event-stream"
			}
			body := []byte(fmt.Sprintf(`{"model":"glm-5.3","max_tokens":32,"stream":%s,%s"messages":[{"role":"user","content":"hi"}]}`, stream, tt.preference))
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{contentType}, "x-request-id": []string{"rid-glm53"}},
				Body:       io.NopCloser(strings.NewReader(responseBody)),
			}}
			svc := newCNNativeGatewayTestService(upstream)
			account := nativeAnthropicTestAccount(PlatformZhipu)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")

			_, err := svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
			require.NoError(t, err)
			require.Equal(t, "enabled", gjson.GetBytes(upstream.lastBody, "thinking.type").String())
			require.Equal(t, tt.wantEffort, gjson.GetBytes(upstream.lastBody, "output_config.effort").String())
		})
	}
}

func TestNativeAnthropicPassthroughLeavesOtherThinkingUntouched(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		body string
	}{
		{name: "glm 5.3 unspecified", body: `{"model":"glm-5.3","max_tokens":32,"stream":false,"messages":[]}`},
		{name: "glm 5.2 disabled", body: `{"model":"glm-5.2","max_tokens":32,"stream":false,"thinking":{"type":"disabled"},"messages":[]}`},
		{name: "non glm adaptive", body: `{"model":"deepseek-v4-pro","max_tokens":32,"stream":false,"thinking":{"type":"adaptive"},"messages":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := []byte(tt.body)
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"id":"msg_1","type":"message","role":"assistant","content":[],"usage":{"input_tokens":1,"output_tokens":1}}`)),
			}}
			svc := newCNNativeGatewayTestService(upstream)
			account := nativeAnthropicTestAccount(PlatformZhipu)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", bytes.NewReader(body))
			c.Request.Header.Set("Content-Type", "application/json")

			_, err := svc.ForwardAsAnthropic(context.Background(), c, account, body, "", "")
			require.NoError(t, err)
			require.JSONEq(t, tt.body, string(upstream.lastBody))
		})
	}
}

func TestCNProviderChatCompletionsAnthropicConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := &cnCountingUpstream{response: cnResponse(http.StatusOK, miniAnthropicSSEStream())}
	svc := newCNNativeGatewayTestService(upstream)
	account := nativeAnthropicTestAccount(PlatformZhipu)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	body := []byte(`{"model":"glm-4.7","messages":[{"role":"user","content":"hi"}],"stream":false}`)
	result, err := svc.forwardChatCompletionsViaNativeAnthropic(context.Background(), c, account, body, "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 1, upstream.calls)
	require.Contains(t, rec.Body.String(), "Hello")
	require.Equal(t, 10, result.Usage.InputTokens)
	require.Equal(t, 5, result.Usage.OutputTokens)
}

func TestCNProviderResponsesAnthropicConversion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := &cnCountingUpstream{response: cnResponse(http.StatusOK, miniAnthropicSSEStream())}
	svc := newCNNativeGatewayTestService(upstream)
	account := nativeAnthropicTestAccount(PlatformKimi)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	body := []byte(`{"model":"kimi-k2","input":"hi","stream":false}`)
	result, err := svc.forwardResponsesViaNativeAnthropic(context.Background(), c, account, body, "")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 1, upstream.calls)
	require.Contains(t, rec.Body.String(), "Hello")
	require.Equal(t, 10, result.Usage.InputTokens)
	require.Equal(t, 5, result.Usage.OutputTokens)
}

func TestDeepSeekResponsesURLAndBodyNormalization(t *testing.T) {
	svc := newCNNativeGatewayTestService(nil)
	account := cnAccount(1, PlatformDeepseek, AccountModePayG, "https://127.0.0.1")
	account.Credentials["api_protocol"] = APIProtocolResponses
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	body := []byte(`{"model":"deepseek-reasoner","store":true,"previous_response_id":"resp_1","input":"hi"}`)
	req, err := svc.buildUpstreamRequest(context.Background(), c, account, body, "", false, "", false)
	require.NoError(t, err)
	require.Equal(t, "https://127.0.0.1/responses", req.URL.String())
	encoded, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(encoded, &payload))
	require.Equal(t, false, payload["store"])
	require.NotContains(t, payload, "previous_response_id")
}

func TestCNProviderProtocolCredentialAndWebSocketPaths(t *testing.T) {
	svc := newCNNativeGatewayTestService(nil)
	account := cnAccount(1, PlatformDeepseek, AccountModePayG, "https://127.0.0.1")
	account.Credentials["api_protocol"] = APIProtocolResponses

	token, kind, err := svc.GetAccessToken(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, "secret", token)
	require.Equal(t, "apikey", kind)

	wsURL, err := svc.buildOpenAIResponsesWSURL(account)
	require.NoError(t, err)
	require.Equal(t, "wss://127.0.0.1/responses", wsURL)
}

func TestCNProviderCountTokensRouting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := &cnCountingUpstream{}
	svc := &OpenAIGatewayService{httpUpstream: upstream}
	account := cnAccount(1, PlatformKimi, AccountModePayG, "https://127.0.0.1")
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages/count_tokens", nil)
	err := svc.ForwardCountTokensAsAnthropic(context.Background(), c, account, []byte(`{"model":"kimi-k2","messages":[{"role":"user","content":"hello"}]}`), "")
	require.NoError(t, err)
	require.Zero(t, upstream.calls)
	var response map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Greater(t, response["input_tokens"].(float64), float64(0))
}

func TestCNProviderReactive429AndBalanceRecovery(t *testing.T) {
	account := cnAccount(1, PlatformDeepseek, AccountModePayG, "https://api.deepseek.com")
	repo := &cnProviderRepoStub{accounts: map[int64]*Account{1: account}}
	rl := &RateLimitService{
		accountRepo: repo,
		cfg:         &config.Config{Gateway: config.GatewayConfig{CNProviders: config.GatewayCNProvidersConfig{BalanceCheckIntervalMinutes: 1}}},
	}
	rl.HandleUpstreamError(context.Background(), account, http.StatusTooManyRequests, nil, []byte(`{"error":{"message":"insufficient balance"}}`))
	require.Equal(t, 1, repo.setCalls)
	require.NotEmpty(t, repo.updates)
	require.Equal(t, true, repo.updates[0][cnExtraKey(PlatformDeepseek, cnBalanceExtraSuffixLow)])

	until := time.Now().Add(time.Hour)
	account.TempUnschedulableUntil = &until
	account.TempUnschedulableReason = cnBalanceLowReason("insufficient balance")
	healthy := &cnCountingUpstream{response: cnResponse(http.StatusOK, `{"is_available":true,"balance_infos":[{"currency":"USD","total_balance":"2"}]}`)}
	checker := NewCNProviderBalanceCheckService(repo, NewCNProviderBalanceService(repo, nil, healthy, nil), nil, &config.Config{}, 0)
	checker.checkOne(context.Background(), account, 1)
	require.Equal(t, 1, repo.clearCalls)
}

func TestCNProviderProtocolProbeMode(t *testing.T) {
	t.Run("non-deepseek skips network and marks unsupported", func(t *testing.T) {
		account := cnAccount(1, PlatformKimi, AccountModePayG, "https://api.kimi.com")
		repo := &cnProviderRepoStub{accounts: map[int64]*Account{1: account}}
		upstream := &cnCountingUpstream{}
		svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream}
		svc.ProbeOpenAIAPIKeyResponsesSupport(context.Background(), 1)
		require.Zero(t, upstream.calls)
		require.Equal(t, false, repo.updates[0][openai_compat.ExtraKeyResponsesSupported])
	})
	t.Run("deepseek responses forces mode without network", func(t *testing.T) {
		account := cnAccount(2, PlatformDeepseek, AccountModePayG, "https://api.deepseek.com")
		account.Credentials["api_protocol"] = APIProtocolResponses
		repo := &cnProviderRepoStub{accounts: map[int64]*Account{2: account}}
		upstream := &cnCountingUpstream{}
		svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream}
		svc.ProbeOpenAIAPIKeyResponsesSupport(context.Background(), 2)
		require.Zero(t, upstream.calls)
		require.Equal(t, string(openai_compat.ResponsesSupportModeForceResponses), repo.updates[0][openai_compat.ExtraKeyResponsesMode])
		require.Equal(t, true, repo.updates[0][openai_compat.ExtraKeyResponsesSupported])
	})
}
