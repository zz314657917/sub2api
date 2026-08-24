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
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestStripEmptyChatToolCallIdentity_FollowingDelta(t *testing.T) {
	payload := []byte(`{"choices":[{"index":0,"delta":{"tool_calls":[{"index":0,"id":"","type":"function","function":{"name":"","arguments":"{\"query\":"}}]}}]}`)

	rewritten, changed := stripEmptyChatToolCallIdentity(payload)

	require.True(t, changed)
	require.False(t, gjson.GetBytes(rewritten, "choices.0.delta.tool_calls.0.id").Exists())
	require.False(t, gjson.GetBytes(rewritten, "choices.0.delta.tool_calls.0.function.name").Exists())
	require.Equal(t, `{"query":`, gjson.GetBytes(rewritten, "choices.0.delta.tool_calls.0.function.arguments").String())
	require.Equal(t, int64(0), gjson.GetBytes(rewritten, "choices.0.delta.tool_calls.0.index").Int())
	require.Equal(t, "function", gjson.GetBytes(rewritten, "choices.0.delta.tool_calls.0.type").String())
}

func TestStripEmptyChatToolCallIdentity_PreservesNonEmptyAndFailOpenPayloads(t *testing.T) {
	first := []byte(`{"choices":[{"delta":{"tool_calls":[{"id":"call_example","type":"function","function":{"name":"web_search","arguments":""}}]}}]}`)
	rewritten, changed := stripEmptyChatToolCallIdentity(first)
	require.False(t, changed)
	require.Equal(t, string(first), string(rewritten))
	require.Equal(t, "call_example", gjson.GetBytes(rewritten, "choices.0.delta.tool_calls.0.id").String())
	require.Equal(t, "", gjson.GetBytes(rewritten, "choices.0.delta.tool_calls.0.function.arguments").String())

	for _, payload := range [][]byte{[]byte(`{"choices":[{`), []byte(`{"choices":[{"delta":{"content":"hi"}}]}`)} {
		rewritten, changed = stripEmptyChatToolCallIdentity(payload)
		require.False(t, changed)
		require.Equal(t, string(payload), string(rewritten))
	}
}

func TestStripEmptyChatToolCallIdentityFromSSELine_PreservesSSEAndClientMergeIdentity(t *testing.T) {
	lines := []string{
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_example","type":"function","function":{"name":"web_search","arguments":""}}]}}]}`,
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"","type":"function","function":{"name":"","arguments":"{\"query\":"}}]}}]}`,
		`data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"","type":"function","function":{"name":"","arguments":"\"example\"}"}}]}}]}`,
	}
	var id, name, arguments string
	for _, line := range lines {
		sanitized := stripEmptyChatToolCallIdentityFromSSELine(line)
		require.True(t, strings.HasPrefix(sanitized, "data: "))
		payload, ok := extractOpenAISSEDataLine(sanitized)
		require.True(t, ok)
		toolCall := gjson.Get(payload, "choices.0.delta.tool_calls.0")
		if value := toolCall.Get("id"); value.Exists() {
			id = value.String()
		}
		if value := toolCall.Get("function.name"); value.Exists() {
			name = value.String()
		}
		arguments += toolCall.Get("function.arguments").String()
	}
	require.Equal(t, "call_example", id)
	require.Equal(t, "web_search", name)
	require.Equal(t, `{"query":"example"}`, arguments)
	require.Equal(t, "data: [DONE]", stripEmptyChatToolCallIdentityFromSSELine("data: [DONE]"))
}

func TestForwardAsRawChatCompletions_StripsEmptyToolCallIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := []byte(`{"model":"gpt-5.6","messages":[{"role":"user","content":"weather"}],"stream":true}`)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	upstreamBody := strings.Join([]string{
		`data: {"id":"chatcmpl_tool","choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_example","type":"function","function":{"name":"web_search","arguments":""}}]}}]}`,
		`data: {"id":"chatcmpl_tool","choices":[{"delta":{"tool_calls":[{"index":0,"id":"","type":"function","function":{"name":"","arguments":"{\"query\":"}}]}}]}`,
		`data: {"id":"chatcmpl_tool","choices":[{"delta":{"tool_calls":[{"index":0,"id":"","type":"function","function":{"name":"","arguments":"\"example\"}"}}]}}]}`,
		`data: [DONE]`,
		"",
	}, "\n\n")
	upstream := &httpUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(upstreamBody)),
	}}
	svc := &OpenAIGatewayService{
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false, AllowInsecureHTTP: true}}},
		httpUpstream: upstream,
	}
	account := &Account{ID: 101, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Concurrency: 1, Credentials: map[string]any{"api_key": "sk-test", "base_url": "http://upstream.example"}}

	result, err := svc.forwardAsRawChatCompletions(context.Background(), c, account, body, "")

	require.NoError(t, err)
	require.NotNil(t, result)
	downstream := rec.Body.String()
	require.Contains(t, downstream, `"id":"call_example"`)
	require.Contains(t, downstream, `"name":"web_search"`)
	require.Contains(t, downstream, "data: [DONE]")
	require.NotContains(t, downstream, `"id":""`)
	require.NotContains(t, downstream, `"name":""`)
}
