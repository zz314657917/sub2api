package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestV027GeminiThinkingVariantRespectsMappings(t *testing.T) {
	account := &Account{Platform: PlatformAntigravity, Credentials: map[string]any{"model_mapping": map[string]any{
		"gemini-3.8-flash-low": "upstream-low", "gemini-3.8-flash-medium": "upstream-medium", "gemini-3.8-flash-high": "upstream-high",
	}}}
	for _, tt := range []struct{ body, want string }{
		{`{"generationConfig":{"thinkingConfig":{"thinkingBudget":1000}}}`, "upstream-low"},
		{`{"generationConfig":{"thinkingConfig":{"thinkingBudget":4000}}}`, "upstream-medium"},
		{`{"generationConfig":{"thinkingConfig":{"thinkingBudget":-1}}}`, "upstream-high"},
	} {
		got, ok := resolveGeminiThinkingVariant(account, "gemini-3.8-flash", []byte(tt.body))
		require.True(t, ok)
		require.Equal(t, tt.want, got)
	}
	account.Credentials["model_mapping"] = map[string]any{"gemini-3.8-flash": "explicit", "gemini-3.8-flash-high": "upstream-high"}
	_, ok := resolveGeminiThinkingVariant(account, "gemini-3.8-flash", []byte(`{}`))
	require.False(t, ok)
	account.Credentials["model_mapping"] = map[string]any{"gemini-3.8-flash": "gemini-3.8-flash", "gemini-3.8-flash-low": "upstream-low", "gemini-3.8-flash-high": "upstream-high"}
	got, ok := resolveGeminiThinkingVariant(account, "gemini-3.8-flash", []byte(`{"generationConfig":{"thinkingConfig":{"thinkingBudget":1000}}}`))
	require.False(t, ok, "explicit identity mapping must prevent thinking variant selection")
	require.Empty(t, got)
	_, ok = resolveGeminiThinkingVariant(account, "gemini-3.8-flash-high", []byte(`{}`))
	require.False(t, ok)
	_, ok = resolveGeminiThinkingVariant(account, "claude-sonnet-4-6", []byte(`{}`))
	require.False(t, ok)
}

func TestV027GeminiForwardUsesThinkingVariantInFakeUpstreamRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	upstream := &queuedHTTPUpstreamStub{responses: []*http.Response{{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader("data: {\"response\":{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"ok\"}]}}]}}\n\n")),
	}}}
	svc := &AntigravityGatewayService{
		tokenProvider:  NewAntigravityTokenProvider(nil, nil, nil),
		httpUpstream:   upstream,
		settingService: &SettingService{cfg: &config.Config{}},
	}
	account := &Account{
		ID: 1, Platform: PlatformAntigravity, Type: AccountTypeUpstream,
		Credentials: map[string]any{
			"api_key": "fake", "project_id": "project",
			"model_mapping": map[string]any{
				"gemini-3.8-flash-low":    "upstream-low",
				"gemini-3.8-flash-medium": "upstream-medium",
				"gemini-3.8-flash-high":   "upstream-high",
			},
		},
	}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3.8-flash:generateContent", nil)
	result, err := svc.ForwardGemini(context.Background(), c, account, "gemini-3.8-flash", "generateContent", false, []byte(`{"contents":[],"generationConfig":{"thinkingConfig":{"thinkingLevel":"low","thinkingBudget":24576}}}`), false)
	require.NoError(t, err)
	require.Equal(t, "upstream-low", result.UpstreamModel)
	require.Len(t, upstream.requestBodies, 1)
	var wrapped struct {
		Model string `json:"model"`
	}
	require.NoError(t, json.Unmarshal(upstream.requestBodies[0], &wrapped))
	require.Equal(t, "upstream-low", wrapped.Model)
}

func TestV027GeminiStreamingSkipsCommentsForGenAI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	run := func(t *testing.T, client string) string {
		t.Helper()
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request, _ = http.NewRequest(http.MethodPost, "/v1beta/models/gemini:streamGenerateContent", nil)
		c.Request.Header.Set("User-Agent", client)
		svc := &AntigravityGatewayService{settingService: &SettingService{cfg: &config.Config{Gateway: config.GatewayConfig{StreamKeepaliveInterval: 1}}}}
		reader, writer := io.Pipe()
		done := make(chan error, 1)
		go func() {
			_, err := svc.handleGeminiStreamingResponse(c, &http.Response{StatusCode: http.StatusOK, Body: reader}, time.Now())
			done <- err
		}()
		_, err := io.WriteString(writer, "data: {\"response\":{\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"ok\"}]}}]}}\n\n")
		require.NoError(t, err)
		time.Sleep(1200 * time.Millisecond)
		require.NoError(t, writer.Close())
		require.NoError(t, <-done)
		return rec.Body.String()
	}
	require.Contains(t, run(t, "curl/8.7.1"), ":\n\n")
	out := run(t, "google-genai-sdk/1.71 gl-go/go1.28")
	require.Contains(t, out, `"text":"ok"`)
	require.NotContains(t, out, ":\n\n")
	require.True(t, geminiClientRejectsSSEComments("google-genai-sdk/1.20 gl-python/3.12"))
	require.False(t, geminiClientRejectsSSEComments("google-genai-sdk/1.20 gl-node/22"))
}
