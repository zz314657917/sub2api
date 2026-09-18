package service

import (
	"bytes"
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

type upstreamResponseModelPathCapture struct {
	body, downstream string
	result           *OpenAIForwardResult
	log              *UsageLog
	billing          *UsageBillingCommand
	calls            int
}

func runUpstreamResponseModelProductionPath(t *testing.T, stream bool, contentType, upstreamBody string) upstreamResponseModelPathCapture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	body := []byte(`{"model":"gpt-5.2","stream":` + map[bool]string{false: "false", true: "true"}[stream] + `,"input":[{"type":"text","text":"hello"}]}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {contentType}, "x-request-id": {"audit"}}, Body: io.NopCloser(strings.NewReader(upstreamBody))}}
	account := &Account{ID: 701, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "test-key", "base_url": "https://api.openai.com"}, Extra: map[string]any{"openai_passthrough": true}, Status: StatusActive, Schedulable: true, Concurrency: 1}
	forward := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
	result, err := forward.Forward(context.Background(), c, account, body)
	require.NoError(t, err)
	require.NotNil(t, result)
	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	billingRepo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
	usage := newOpenAIRecordUsageServiceWithBillingRepoForTest(usageRepo, billingRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
	require.NoError(t, usage.RecordUsage(context.Background(), &OpenAIRecordUsageInput{Result: result, APIKey: &APIKey{ID: 501, Quota: 100}, User: &User{ID: 601}, Account: account, RequestIDOverride: "audit-fixed"}))
	require.NotNil(t, usageRepo.lastLog)
	require.NotNil(t, billingRepo.lastCmd)
	return upstreamResponseModelPathCapture{body: string(upstream.lastBody), downstream: rec.Body.String(), result: result, log: usageRepo.lastLog, billing: billingRepo.lastCmd, calls: len(upstream.requests)}
}

func requireUpstreamResponseModelNonInterference(t *testing.T, same, different upstreamResponseModelPathCapture) {
	t.Helper()
	require.Equal(t, same.body, different.body, "observation must not change outbound payload")
	require.Equal(t, normalizeUpstreamModelDeclaration(t, same.downstream), normalizeUpstreamModelDeclaration(t, different.downstream), "only provider-declared response model and response IDs may differ downstream")
	require.Equal(t, same.result.Model, different.result.Model)
	require.Equal(t, same.result.UpstreamModel, different.result.UpstreamModel)
	require.Equal(t, same.result.BillingModel, different.result.BillingModel)
	require.Equal(t, same.result.ServiceTier, different.result.ServiceTier)
	require.Equal(t, same.result.Usage, different.result.Usage)
	require.Equal(t, same.log.Model, different.log.Model)
	require.Equal(t, same.log.UpstreamModel, different.log.UpstreamModel)
	require.Equal(t, same.log.TotalCost, different.log.TotalCost)
	require.Equal(t, same.log.ActualCost, different.log.ActualCost)
	require.Equal(t, same.billing, different.billing)
	require.Equal(t, same.calls, different.calls)
	require.Equal(t, "gpt-5.2", same.result.UpstreamResponseModel)
	require.Equal(t, "gpt-5.3", different.result.UpstreamResponseModel)
	require.False(t, *same.log.UpstreamModelMismatch)
	require.True(t, *different.log.UpstreamModelMismatch)
}

func normalizeUpstreamModelDeclaration(t *testing.T, body string) string {
	t.Helper()
	if strings.Contains(body, "data: ") {
		lines := strings.Split(body, "\n")
		for i, line := range lines {
			if !strings.HasPrefix(line, "data: ") || line == "data: [DONE]" {
				continue
			}
			lines[i] = "data: " + normalizeUpstreamModelJSON(t, strings.TrimPrefix(line, "data: "))
		}
		return strings.Join(lines, "\n")
	}
	return normalizeUpstreamModelJSON(t, body)
}

func normalizeUpstreamModelJSON(t *testing.T, body string) string {
	t.Helper()
	var payload any
	require.NoError(t, json.Unmarshal([]byte(body), &payload))
	stripUpstreamModelDeclaration(payload)
	normalized, err := json.Marshal(payload)
	require.NoError(t, err)
	return string(normalized)
}

func stripUpstreamModelDeclaration(value any) {
	switch typed := value.(type) {
	case map[string]any:
		delete(typed, "id")
		delete(typed, "model")
		for _, nested := range typed {
			stripUpstreamModelDeclaration(nested)
		}
	case []any:
		for _, nested := range typed {
			stripUpstreamModelDeclaration(nested)
		}
	}
}

func TestUpstreamResponseModelHTTPNonInterference(t *testing.T) {
	same := runUpstreamResponseModelProductionPath(t, false, "application/json", `{"id":"same","model":"gpt-5.2","output":[],"usage":{"input_tokens":11,"output_tokens":7}}`)
	different := runUpstreamResponseModelProductionPath(t, false, "application/json", `{"id":"different","model":"gpt-5.3","output":[],"usage":{"input_tokens":11,"output_tokens":7}}`)
	requireUpstreamResponseModelNonInterference(t, same, different)
}

func TestUpstreamResponseModelSSENonInterference(t *testing.T) {
	same := runUpstreamResponseModelProductionPath(t, true, "text/event-stream", "data: {\"type\":\"response.completed\",\"response\":{\"model\":\"gpt-5.2\",\"usage\":{\"input_tokens\":11,\"output_tokens\":7}}}\n\ndata: [DONE]\n\n")
	different := runUpstreamResponseModelProductionPath(t, true, "text/event-stream", "data: {\"type\":\"response.completed\",\"response\":{\"model\":\"gpt-5.3\",\"usage\":{\"input_tokens\":11,\"output_tokens\":7}}}\n\ndata: [DONE]\n\n")
	requireUpstreamResponseModelNonInterference(t, same, different)
}

func runUpstreamResponseModelWSProductionPath(t *testing.T, observedModel string) upstreamResponseModelPathCapture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/openai/v1/responses", nil)
	c.Request.Header.Set("User-Agent", "unit-test-agent/1.0")
	groupID := int64(701)
	c.Set("api_key", &APIKey{GroupID: &groupID})

	cfg := &config.Config{}
	cfg.Security.URLAllowlist.Enabled = false
	cfg.Security.URLAllowlist.AllowInsecureHTTP = true
	cfg.Gateway.OpenAIWS.Enabled = true
	cfg.Gateway.OpenAIWS.OAuthEnabled = true
	cfg.Gateway.OpenAIWS.APIKeyEnabled = true
	cfg.Gateway.OpenAIWS.ResponsesWebsocketsV2 = true
	cfg.Gateway.OpenAIWS.MaxConnsPerAccount = 1
	cfg.Gateway.OpenAIWS.MinIdlePerAccount = 0
	cfg.Gateway.OpenAIWS.MaxIdlePerAccount = 1
	cfg.Gateway.OpenAIWS.QueueLimitPerConn = 8
	cfg.Gateway.OpenAIWS.DialTimeoutSeconds = 3
	cfg.Gateway.OpenAIWS.ReadTimeoutSeconds = 5
	cfg.Gateway.OpenAIWS.WriteTimeoutSeconds = 3

	captureConn := &openAIWSCaptureConn{events: [][]byte{
		[]byte(`{"type":"response.created","response":{"id":"resp_audit","model":"` + observedModel + `"}}`),
		[]byte(`{"type":"response.completed","response":{"id":"resp_audit","model":"` + observedModel + `","usage":{"input_tokens":11,"output_tokens":7}}}`),
	}}
	pool := newOpenAIWSConnPool(cfg)
	t.Cleanup(pool.Close)
	pool.setClientDialerForTest(&openAIWSCaptureDialer{conn: captureConn})
	account := &Account{
		ID:          701,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: true,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key": "sk-test",
			"model_mapping": map[string]any{
				"custom-original-model": "gpt-5.2",
			},
		},
		Extra: map[string]any{"responses_websockets_v2_enabled": true},
	}
	svc := &OpenAIGatewayService{
		cfg:              cfg,
		httpUpstream:     &httpUpstreamRecorder{},
		cache:            &stubGatewayCache{},
		openaiWSResolver: NewOpenAIWSProtocolResolver(cfg),
		toolCorrector:    NewCodexToolCorrector(),
		openaiWSPool:     pool,
	}
	body := []byte(`{"model":"custom-original-model","stream":false,"input":[{"type":"input_text","text":"hello"}]}`)
	result, err := svc.Forward(context.Background(), c, account, body)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.OpenAIWSMode, "审计必须覆盖生产 WebSocket 转发路径")
	require.Len(t, captureConn.writes, 1)

	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	billingRepo := &openAIRecordUsageBillingRepoStub{result: &UsageBillingApplyResult{Applied: true}}
	usage := newOpenAIRecordUsageServiceWithBillingRepoForTest(usageRepo, billingRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
	require.NoError(t, usage.RecordUsage(context.Background(), &OpenAIRecordUsageInput{Result: result, APIKey: &APIKey{ID: 501, Quota: 100}, User: &User{ID: 601}, Account: account, RequestIDOverride: "audit-fixed"}))
	require.NotNil(t, usageRepo.lastLog)
	require.NotNil(t, billingRepo.lastCmd)
	return upstreamResponseModelPathCapture{body: requestToJSONString(captureConn.writes[0]), downstream: rec.Body.String(), result: result, log: usageRepo.lastLog, billing: billingRepo.lastCmd, calls: len(captureConn.writes)}
}

func TestUpstreamResponseModelWSNonInterference(t *testing.T) {
	same := runUpstreamResponseModelWSProductionPath(t, "gpt-5.2")
	different := runUpstreamResponseModelWSProductionPath(t, "gpt-5.3")
	requireUpstreamResponseModelNonInterference(t, same, different)
}
