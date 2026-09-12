package service

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestStatelessReasoningReplayBoundaries(t *testing.T) {
	body := []byte(`{"store":false,"metadata":{"n":9007199254740993},"input":[{"type":"reasoning","id":"rs_missing","encrypted_content":"opaque","call_id":"wrong"},{"type":"reasoning","id":"rs_empty","summary":[]},{"type":"item_reference","id":"rs_missing"},{"type":"item_reference","id":"msg_keep"},{"type":"function_call","id":"fc_keep","call_id":"call_pair","name":"run","arguments":"{}"},{"type":"function_call_output","call_id":"call_pair","output":"ok"},{"type":"compaction","encrypted_content":"compact"}]}`)
	out, changed, err := normalizeOpenAIAPIKeyStoreFalseReasoningReplay(body, false)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, int64(5), gjson.GetBytes(out, "input.#").Int())
	require.False(t, gjson.GetBytes(out, "input.0.id").Exists())
	require.False(t, gjson.GetBytes(out, "input.0.call_id").Exists())
	require.Equal(t, "opaque", gjson.GetBytes(out, "input.0.encrypted_content").String())
	require.Equal(t, "[]", gjson.GetBytes(out, "input.0.summary").Raw)
	require.Equal(t, "msg_keep", gjson.GetBytes(out, "input.1.id").String())
	require.Equal(t, "call_pair", gjson.GetBytes(out, "input.2.call_id").String())
	require.Equal(t, "call_pair", gjson.GetBytes(out, "input.3.call_id").String())
	require.Equal(t, "compact", gjson.GetBytes(out, "input.4.encrypted_content").String())
	require.Equal(t, "9007199254740993", gjson.GetBytes(out, "metadata.n").Raw)
	require.Contains(t, string(body), `"rs_missing"`)
	again, changed, err := normalizeOpenAIAPIKeyStoreFalseReasoningReplay(out, false)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, out, again)
	for _, field := range []string{`"store":true,`, `"store":null,`, `"store":"false",`, ""} {
		original := []byte(`{` + field + `"input":[{"type":"reasoning","id":"rs_keep"},{"type":"item_reference","id":"rs_keep"}]}`)
		out, changed, err := normalizeOpenAIAPIKeyStoreFalseReasoningReplay(original, false)
		require.NoError(t, err)
		require.False(t, changed)
		require.Equal(t, original, out)
	}
	_, changed, err = normalizeOpenAIAPIKeyStoreFalseReasoningReplay([]byte(`{"input":[{"type":"reasoning","id":"rs_old","encrypted_content":"opaque"}]}`), true)
	require.NoError(t, err)
	require.True(t, changed)
}

func TestStatelessReasoningReplayForward(t *testing.T) {
	for _, compact := range []bool{false, true} {
		for _, passthrough := range []bool{false, true} {
			for _, stream := range []bool{false, true} {
				t.Run(fmt.Sprintf("compact=%v/pass=%v/stream=%v", compact, passthrough, stream), func(t *testing.T) {
					u := &cfportHTTP{sse: stream}
					s := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: u, toolCorrector: NewCodexToolCorrector()}
					a := &Account{ID: 42, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Concurrency: 1, Credentials: map[string]any{"api_key": "test", "access_token": "test"}, Extra: map[string]any{"openai_passthrough": passthrough, "openai_oauth_responses_websockets_v2_mode": OpenAIWSIngressModeOff}}
					c, _ := gin.CreateTestContext(httptest.NewRecorder())
					path, storeField := "/v1/responses", `"store":false,`
					if compact {
						path, storeField = "/v1/responses/compact", ""
					}
					c.Request = httptest.NewRequest("POST", path, nil)
					body := []byte(fmt.Sprintf(`{"model":"gpt-5.5",%s"stream":%v,"instructions":"Test","input":[{"type":"reasoning","id":"rs_43nl2pzqmqgup6g4hadwwu7csqi6thmhe2lzgj3olal4xox5a3bq","encrypted_content":"opaque"},{"type":"item_reference","id":"rs_missing"},{"role":"user","content":"hello"}]}`, storeField, stream))
					_, err := s.Forward(context.Background(), c, a, body)
					require.NoError(t, err)
					require.Equal(t, 1, u.calls)
					require.NotContains(t, string(u.body), "rs_")
					require.Equal(t, "opaque", gjson.GetBytes(u.body, "input.0.encrypted_content").String())
					require.Equal(t, "hello", gjson.GetBytes(u.body, "input.1.content").String())
				})
			}
		}
	}
}

func TestStatelessReasoningReplayWebSocketCompatibility(t *testing.T) {
	for _, accountType := range []string{AccountTypeAPIKey, AccountTypeOAuth} {
		body := []byte(`{"type":"response.create","store":false,"input":[{"type":"reasoning","id":"rs_missing","encrypted_content":"opaque"}]}`)
		out, changed, err := normalizeOpenAIResponsesWebSocketCompatibilityBody(body, &Account{Platform: PlatformOpenAI, Type: accountType})
		require.NoError(t, err)
		if accountType == AccountTypeAPIKey {
			require.True(t, changed)
			require.False(t, gjson.GetBytes(out, "input.0.id").Exists())
		} else {
			// OAuth keeps its existing transform; upstream scopes this helper to API keys.
			require.Equal(t, "rs_missing", gjson.GetBytes(out, "input.0.id").String())
		}
		require.Equal(t, "opaque", gjson.GetBytes(out, "input.0.encrypted_content").String())
	}
}

func TestStatelessReasoningReplayOAuthPassthrough(t *testing.T) {
	body := []byte(`{"store":false,"input":[{"type":"reasoning","id":"rs_oauth_missing","encrypted_content":"opaque"},{"type":"item_reference","id":"rs_oauth_missing"}]}`)
	out, changed, err := normalizeOpenAIPassthroughOAuthBody(body, false)
	require.NoError(t, err)
	require.True(t, changed)
	require.False(t, gjson.GetBytes(out, "input.0.id").Exists())
	require.False(t, gjson.GetBytes(out, "input.1").Exists())
	require.Equal(t, "opaque", gjson.GetBytes(out, "input.0.encrypted_content").String())
}
