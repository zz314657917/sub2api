package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
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
