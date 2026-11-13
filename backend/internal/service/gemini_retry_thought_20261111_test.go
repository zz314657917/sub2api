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

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type geminiRetry20261111Upstream struct {
	calls       int
	statuses    []int
	bodies      []string
	transport   bool
	transportAt map[int]bool
	recorder    *httptest.ResponseRecorder
}

func (s *geminiRetry20261111Upstream) Do(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	i := s.calls
	s.calls++
	if s.transport || s.transportAt[i] {
		return nil, errors.New("test transport failure")
	}
	if i >= len(s.statuses) {
		return nil, errors.New("unexpected extra request")
	}
	return &http.Response{StatusCode: s.statuses[i], Header: http.Header{}, Body: io.NopCloser(strings.NewReader(s.bodies[i]))}, nil
}

func (s *geminiRetry20261111Upstream) DoWithTLS(r *http.Request, p string, id int64, n int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return s.Do(r, p, id, n)
}

func geminiRetry20261111Forward(t *testing.T, route string, retries int, upstream *geminiRetry20261111Upstream) error {
	t.Helper()
	gin.SetMode(gin.TestMode)
	upstream.recorder = httptest.NewRecorder()
	c, _ := gin.CreateTestContext(upstream.recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	cfg := &config.Config{}
	cfg.Gateway.MaxSameAccountRetriesGemini = retries
	s := &GeminiMessagesCompatService{cfg: cfg, httpUpstream: upstream, tokenProvider: &GeminiTokenProvider{}}
	account := &Account{ID: 101, Platform: PlatformGemini, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "test_key"}, Concurrency: 1}
	var err error
	switch route {
	case "native":
		_, err = s.ForwardNative(context.Background(), c, account, "gemini-3-flash", "generateContent", false, []byte(`{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`))
	case "messages":
		_, err = s.Forward(context.Background(), c, account, []byte(`{"model":"gemini-3-flash","max_tokens":100,"messages":[{"role":"user","content":"hi"}]}`))
	default:
		_, err = s.ForwardAsChatCompletions(context.Background(), c, account, []byte(`{"model":"gemini-3-flash","messages":[{"role":"user","content":"hi"}]}`))
	}
	return err
}

func TestGeminiRetry20261111(t *testing.T) {
	require.Equal(t, 5, geminiSameAccountMaxAttempts(nil))
	configured := &config.Config{}
	configured.Gateway.MaxSameAccountRetriesGemini = 4
	require.Equal(t, 5, geminiSameAccountMaxAttempts(configured))
	for _, route := range []string{"native", "messages", "chat"} {
		for _, retries := range []int{0, 1} {
			for _, transport := range []bool{false, true} {
				t.Run(route+"/"+string(rune('0'+retries))+"/"+map[bool]string{false: "status", true: "transport"}[transport], func(t *testing.T) {
					upstream := &geminiRetry20261111Upstream{statuses: []int{503, 503}, bodies: []string{`{"error":{"message":"unavailable"}}`, `{"error":{"message":"unavailable"}}`}, transport: transport}
					err := geminiRetry20261111Forward(t, route, retries, upstream)
					require.Error(t, err)
					require.Equal(t, retries+1, upstream.calls)
				})
			}
		}
	}
}

func TestGeminiRetry20261111SignatureStages(t *testing.T) {
	upstream := &geminiRetry20261111Upstream{
		statuses: []int{400, 400, 200},
		bodies: []string{
			`{"error":{"message":"Invalid thought_signature in functionCall"}}`,
			`{"error":{"message":"Invalid thought_signature in functionCall"}}`,
			`{"candidates":[{"content":{"parts":[{"text":"ok"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`,
		},
	}
	require.NoError(t, geminiRetry20261111Forward(t, "messages", 0, upstream))
	require.Equal(t, 3, upstream.calls)
	upstream.statuses[2] = 503
	upstream.bodies[2] = `{"error":{"message":"unavailable"}}`
	upstream.calls = 0
	require.Error(t, geminiRetry20261111Forward(t, "messages", 0, upstream))
	require.Equal(t, 3, upstream.calls)
}

func TestGeminiThought20261111(t *testing.T) {
	response := map[string]any{
		"candidates": []any{map[string]any{
			"content": map[string]any{"parts": []any{
				map[string]any{"thought": true, "text": "private"},
				map[string]any{"thought": false, "text": "visible"},
				map[string]any{"thought": "true", "text": "string"},
				map[string]any{"text": "plain"},
				map[string]any{"thought": true, "functionCall": map[string]any{"name": "lookup", "args": map[string]any{"q": "x"}}},
				map[string]any{"thought": true, "inlineData": map[string]any{"mimeType": "image/png", "data": "aW1hZ2U="}},
			}},
			"finishReason": "STOP",
		}},
		"usageMetadata": map[string]any{"promptTokenCount": 10, "candidatesTokenCount": 5, "thoughtsTokenCount": 2, "cachedContentTokenCount": 3},
	}
	raw, err := json.Marshal(response)
	require.NoError(t, err)
	messages, usage := convertGeminiToClaudeMessage(response, "gemini-3-flash", raw, true)
	require.Equal(t, 7, usage.OutputTokens)
	require.Equal(t, 7, usage.InputTokens)
	require.Equal(t, 3, usage.CacheReadInputTokens)
	blocks := messages["content"].([]any)
	require.Len(t, blocks, 5)
	encoded, err := json.Marshal(messages)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "private")
	for _, kept := range []string{"visible", "string", "plain", "lookup", "data:image/png;base64,aW1hZ2U="} {
		require.Contains(t, string(encoded), kept)
	}
	chat, chatUsage, err := geminiResponseToChatCompletions(response, "gemini-3-flash", raw, nil)
	require.NoError(t, err)
	require.Equal(t, 7, chatUsage.OutputTokens)
	encoded, err = json.Marshal(chat)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "private")
	for _, kept := range []string{"visible", "string", "plain", "lookup", "data:image/png;base64,aW1hZ2U="} {
		require.Contains(t, string(encoded), kept)
	}
}

func TestGeminiRetry20261111MixedSignatureBudget(t *testing.T) {
	const signature = `{"error":{"message":"Invalid thought_signature in functionCall"}}`
	const unavailable = `{"error":{"message":"unavailable"}}`
	const success = `{"candidates":[{"content":{"parts":[{"text":"ok"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":1}}`
	for _, tc := range []struct {
		name        string
		statuses    []int
		bodies      []string
		transportAt map[int]bool
		wantError   bool
	}{
		{"status_then_downgrades", []int{503, 400, 400, 200}, []string{unavailable, signature, signature, success}, nil, false},
		{"downgrade_transport_downgrade_terminal", []int{400, 0, 400, 503}, []string{signature, "", signature, unavailable}, map[int]bool{1: true}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			upstream := &geminiRetry20261111Upstream{statuses: tc.statuses, bodies: tc.bodies, transportAt: tc.transportAt}
			err := geminiRetry20261111Forward(t, "messages", 1, upstream)
			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, 4, upstream.calls)
		})
	}
}

func TestGeminiThought20261111ForwardUnsolicited(t *testing.T) {
	const body = `{"candidates":[{"content":{"parts":[{"thought":true,"text":"private unsolicited reasoning"},{"text":"visible answer"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":10,"candidatesTokenCount":5,"thoughtsTokenCount":2}}`
	for _, route := range []string{"messages", "chat"} {
		t.Run(route, func(t *testing.T) {
			upstream := &geminiRetry20261111Upstream{statuses: []int{200}, bodies: []string{body}}
			require.NoError(t, geminiRetry20261111Forward(t, route, 0, upstream))
			require.Equal(t, 1, upstream.calls)
			require.Equal(t, http.StatusOK, upstream.recorder.Code)
			wire := upstream.recorder.Body.String()
			require.Contains(t, wire, "visible answer")
			require.NotContains(t, wire, "private unsolicited reasoning")
			var response map[string]any
			require.NoError(t, json.Unmarshal([]byte(wire), &response))
			usage, ok := response["usage"].(map[string]any)
			require.True(t, ok)
			if route == "messages" {
				require.Equal(t, float64(7), usage["output_tokens"])
			} else {
				require.Equal(t, float64(7), usage["completion_tokens"])
			}
		})
	}
}
