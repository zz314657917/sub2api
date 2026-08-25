package service

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestNormalizeOpenAIResponsesLiteTools_MovesNamespacesAndKeepsSupportedTools(t *testing.T) {
	reqBody := map[string]any{
		"reasoning":           map[string]any{"effort": "high", "context": "current_turn"},
		"parallel_tool_calls": true,
		"input":               "hello",
		"tools": []any{
			map[string]any{"type": "function", "name": "shell"},
			map[string]any{"type": "custom", "name": "exec"},
			map[string]any{"type": "tool_search"},
			map[string]any{"type": "namespace", "name": "collaboration", "tools": []any{map[string]any{"type": "function", "name": "spawn_agent"}}},
		},
	}

	changed, err := normalizeOpenAIResponsesLiteTools(reqBody)

	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "all_turns", reqBody["reasoning"].(map[string]any)["context"])
	require.Equal(t, false, reqBody["parallel_tool_calls"])
	tools := reqBody["tools"].([]any)
	require.Len(t, tools, 3)
	require.Equal(t, "function", tools[0].(map[string]any)["type"])
	input := reqBody["input"].([]any)
	require.Len(t, input, 2)
	require.Equal(t, "hello", input[0].(map[string]any)["content"])
	additional := input[1].(map[string]any)
	require.Equal(t, "additional_tools", additional["type"])
	require.Equal(t, "collaboration", additional["tools"].([]any)[0].(map[string]any)["name"])
}

func TestNormalizeOpenAIResponsesLiteTools_ValidatesAndPinsNoToolRequests(t *testing.T) {
	noTools := map[string]any{"reasoning": map[string]any{"context": "all_turns"}, "parallel_tool_calls": true}
	changed, err := normalizeOpenAIResponsesLiteTools(noTools)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, false, noTools["parallel_tool_calls"])

	withTools := map[string]any{"tools": []any{map[string]any{"type": "function", "name": "shell"}}}
	changed, err = normalizeOpenAIResponsesLiteTools(withTools)
	require.NoError(t, err)
	require.True(t, changed)
	require.Contains(t, withTools, "parallel_tool_calls")
	require.Equal(t, false, withTools["parallel_tool_calls"])

	for name, reqBody := range map[string]map[string]any{
		"parallel":  {"parallel_tool_calls": "false"},
		"reasoning": {"reasoning": []any{}},
		"tools":     {"tools": map[string]any{}},
	} {
		t.Run(name, func(t *testing.T) {
			changed, err := normalizeOpenAIResponsesLiteTools(reqBody)
			require.Error(t, err)
			require.False(t, changed)
		})
	}
}

func TestNormalizeOpenAIResponsesLiteTools_RejectsConflictingAdditionalTool(t *testing.T) {
	reqBody := map[string]any{
		"tools": []any{map[string]any{"type": "namespace", "name": "collaboration", "tools": []any{map[string]any{"type": "function", "name": "spawn_agent"}}}},
		"input": []any{map[string]any{"type": "additional_tools", "tools": []any{map[string]any{"type": "namespace", "name": "collaboration", "tools": []any{map[string]any{"type": "function", "name": "send_message"}}}}}},
	}

	changed, err := normalizeOpenAIResponsesLiteTools(reqBody)

	require.ErrorContains(t, err, `conflicts with migrated tool type "namespace" name "collaboration"`)
	require.False(t, changed)
	require.Len(t, reqBody["tools"], 1)
}

func TestNormalizeOpenAIResponsesLiteToolsPayload_PreservesResponseCreateShape(t *testing.T) {
	body := []byte(`{"type":"response.create","model":"gpt-5.6-terra","client_metadata":{"ws_request_header_x_openai_internal_codex_responses_lite":"true"},"input":[{"type":"message","role":"user","content":"hello"}],"tools":[{"type":"namespace","name":"collaboration"}],"tool_choice":{"type":"namespace","name":"collaboration"}}`)

	updated, changed, err := normalizeOpenAIResponsesLiteToolsPayload(body)

	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "response.create", gjson.GetBytes(updated, "type").String())
	require.False(t, gjson.GetBytes(updated, "tools").Exists())
	require.Equal(t, "collaboration", gjson.GetBytes(updated, `input.#(type=="additional_tools").tools.0.name`).String())
	require.Equal(t, "all_turns", gjson.GetBytes(updated, "reasoning.context").String())
	require.False(t, gjson.GetBytes(updated, "parallel_tool_calls").Bool())
}

func TestNormalizeOpenAIResponsesLiteToolsPayload_PreservesLargeNumbers(t *testing.T) {
	body := []byte(`{"type":"response.create","sequence":900719925474099312345,"input":"hello"}`)

	updated, changed, err := normalizeOpenAIResponsesLiteToolsPayload(body)

	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "900719925474099312345", gjson.GetBytes(updated, "sequence").Raw)
	require.Equal(t, false, gjson.GetBytes(updated, "parallel_tool_calls").Bool())
}

func TestNormalizeOpenAIResponsesLitePayloadForAccount_PinsAPIKeyParallelCalls(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	body := []byte(`{"type":"response.create","parallel_tool_calls":true,"input":"hello"}`)

	updated, changed, err := normalizeOpenAIResponsesLitePayloadForAccount(body, account)

	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, false, gjson.GetBytes(updated, "parallel_tool_calls").Bool())
	require.Equal(t, "hello", gjson.GetBytes(updated, "input").String())
}

func TestApplyCodexOAuthTransform_PreservesLiteNamespaceToolChoice(t *testing.T) {
	reqBody := map[string]any{
		"input":       []any{map[string]any{"type": "additional_tools", "tools": []any{map[string]any{"type": "namespace", "name": "collaboration"}}}},
		"tool_choice": map[string]any{"type": "namespace", "name": "collaboration"},
	}

	applyCodexOAuthTransform(reqBody, true, false)

	require.Equal(t, map[string]any{"type": "namespace", "name": "collaboration"}, reqBody["tool_choice"])
}

func TestOpenAIResponsesLiteForward_NormalizesOAuthAndRejectsMalformedRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	account := &Account{
		ID: 501, Name: "responses-lite", Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Concurrency: 1, Status: StatusActive, Schedulable: true, RateMultiplier: f64p(1),
		Credentials: map[string]any{"access_token": "oauth-token", "chatgpt_account_id": "chatgpt-account"},
	}
	for _, passthrough := range []bool{false, true} {
		t.Run("passthrough="+strconv.FormatBool(passthrough), func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewReader(nil))
			c.Request.Header.Set(responsesLiteHeader, "true")
			upstream := &httpUpstreamRecorder{resp: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
				Body:       io.NopCloser(strings.NewReader("data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_lite\",\"usage\":{\"input_tokens\":1,\"output_tokens\":1}}}\n\ndata: [DONE]\n\n")),
			}}
			svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: upstream}
			account.Extra = map[string]any{"openai_passthrough": passthrough}
			body := []byte(`{"model":"gpt-5.6-terra","stream":true,"instructions":"test","reasoning":{"effort":"high","context":"current_turn"},"parallel_tool_calls":true,"tools":[{"type":"function","name":"shell"},{"type":"namespace","name":"collaboration"}],"input":[{"type":"message","role":"user","content":"hello"}],"tool_choice":{"type":"namespace","name":"collaboration"}}`)

			result, err := svc.Forward(context.Background(), c, account, body)

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, "true", upstream.lastReq.Header.Get(responsesLiteHeader))
			require.Equal(t, "all_turns", gjson.GetBytes(upstream.lastBody, "reasoning.context").String())
			require.False(t, gjson.GetBytes(upstream.lastBody, `tools.#(type=="namespace")`).Exists())
			require.Equal(t, "collaboration", gjson.GetBytes(upstream.lastBody, `input.#(type=="additional_tools").tools.0.name`).String())
			require.False(t, gjson.GetBytes(upstream.lastBody, "parallel_tool_calls").Bool())
		})
	}

	badRec := httptest.NewRecorder()
	badCtx, _ := gin.CreateTestContext(badRec)
	badCtx.Request = httptest.NewRequest(http.MethodPost, "/v1/responses", nil)
	badCtx.Request.Header.Set(responsesLiteHeader, "true")
	badUpstream := &httpUpstreamRecorder{}
	svc := &OpenAIGatewayService{cfg: &config.Config{}, httpUpstream: badUpstream}
	result, err := svc.Forward(context.Background(), badCtx, account, []byte(`{"model":"gpt-5.6-terra","parallel_tool_calls":"false"}`))
	require.ErrorContains(t, err, "parallel_tool_calls to be a boolean")
	require.Nil(t, result)
	require.Equal(t, http.StatusBadRequest, badRec.Code)
	require.Equal(t, "parallel_tool_calls", gjson.Get(badRec.Body.String(), "error.param").String())
	require.Nil(t, badUpstream.lastReq)
}
