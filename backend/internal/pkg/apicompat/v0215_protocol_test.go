//go:build unit

package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV0215ProtocolInstructionRoles(t *testing.T) {
	for _, role := range []string{"system", "developer"} {
		for _, content := range []string{`"Answer in French."`, `[{"type":"text","text":"Answer in French."}]`} {
			t.Run(role+"/"+content, func(t *testing.T) {
				resp, err := ChatCompletionsToResponses(&ChatCompletionsRequest{
					Messages: []ChatMessage{{Role: role, Content: json.RawMessage(content)}},
				})
				require.NoError(t, err)
				var items []ResponsesInputItem
				require.NoError(t, json.Unmarshal(resp.Input, &items))
				require.Len(t, items, 1)
				require.Equal(t, "message", items[0].Type)
				require.Equal(t, role, items[0].Role)
				want := content
				if content[0] == '[' {
					want = `[{"type":"input_text","text":"Answer in French."}]`
				}
				require.JSONEq(t, want, string(items[0].Content))
			})
		}
	}
}

func TestV0215ProtocolToolChoice(t *testing.T) {
	cases := []struct{ raw, want string }{
		{`{"type":"function","function":{"name":" lookup "}}`, `{"type":"function","name":"lookup"}`},
		{`"auto"`, `"auto"`}, {`"none"`, `"none"`}, {`"required"`, `"required"`},
		{`{"type":"function","name":"lookup"}`, `{"type":"function","name":"lookup"}`},
		{`{"type":"allowed_tools","tools":[]}`, `{"type":"allowed_tools","tools":[]}`},
		{`{"type":"function","function":{"name":" "}}`, `{"type":"function","function":{"name":" "}}`},
		{`{"type":"custom","function":{"name":"lookup"}}`, `{"type":"custom","function":{"name":"lookup"}}`},
	}
	for _, tt := range cases {
		t.Run(tt.raw, func(t *testing.T) {
			resp, err := ChatCompletionsToResponses(&ChatCompletionsRequest{ToolChoice: json.RawMessage(tt.raw)})
			require.NoError(t, err)
			require.JSONEq(t, tt.want, string(resp.ToolChoice))
			if tt.raw == tt.want {
				require.Equal(t, tt.raw, string(resp.ToolChoice))
			}
		})
	}
	raw := json.RawMessage(`{"type":`)
	require.Equal(t, raw, convertChatToolChoiceToResponses(raw))
}

func TestV0215ProtocolLegacyPairing(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "function", Name: "orphan", Content: json.RawMessage(`"orphan"`)},
		{Role: "assistant", Content: json.RawMessage(`"explanation"`), FunctionCall: &ChatFunctionCall{Name: "lookup", Arguments: `{"n":1}`}},
		{Role: "assistant", FunctionCall: &ChatFunctionCall{Name: "lookup", Arguments: `{"n":2}`}},
		{Role: "assistant", FunctionCall: &ChatFunctionCall{Name: "other"}},
		{Role: "function", Name: "other", Content: json.RawMessage(`"third"`)},
		{Role: "function", Name: "lookup", Content: json.RawMessage(`"first"`)},
		{Role: "function", Name: "lookup", Content: json.RawMessage(`"second"`)},
		{Role: "assistant", ToolCalls: []ChatToolCall{{ID: "call_legacy_1", Type: "function", Function: ChatFunctionCall{Name: "modern", Arguments: `{}`}}},
			FunctionCall: &ChatFunctionCall{Name: "ignored"}},
		{Role: "tool", ToolCallID: "call_legacy_1", Content: json.RawMessage(`"modern"`)},
		{Role: "tool", ToolCallID: "call_legacy_2", Content: json.RawMessage(`"reserved"`)},
	}
	first, err := convertChatMessagesToResponsesInput(msgs)
	require.NoError(t, err)
	second, err := convertChatMessagesToResponsesInput(msgs)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.Nil(t, msgs[1].ToolCalls, "conversion must not mutate caller history")
	require.Len(t, first, 11)
	require.Equal(t, "orphan", first[0].CallID)
	require.Equal(t, "message", first[1].Type)
	require.Equal(t, "assistant", first[1].Role)
	require.Equal(t, "call_legacy_3", first[2].CallID)
	require.Equal(t, "call_legacy_4", first[3].CallID)
	require.Equal(t, "call_legacy_5", first[4].CallID)
	require.Equal(t, "{}", first[4].Arguments)
	require.Equal(t, first[4].CallID, first[5].CallID)
	require.Equal(t, first[2].CallID, first[6].CallID)
	require.Equal(t, first[3].CallID, first[7].CallID)
	require.Equal(t, "modern", first[8].Name)
	require.Equal(t, "call_legacy_1", first[8].CallID)
	require.Equal(t, first[8].CallID, first[9].CallID)
	require.Equal(t, "call_legacy_2", first[10].CallID)
}

func TestV0215ProtocolRefusalNonstream(t *testing.T) {
	var resp ResponsesResponse
	require.NoError(t, json.Unmarshal([]byte(`{"id":"resp_refusal","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"normal"},{"type":"refusal","refusal":"can't "},{"type":"refusal","refusal":"help"}]}]}`), &resp))
	chat := ResponsesToChatCompletions(&resp, "gpt-4o")
	require.Equal(t, "can't help", chat.Choices[0].Message.Refusal)
	require.JSONEq(t, `"normal"`, string(chat.Choices[0].Message.Content))
	wire, err := json.Marshal(chat)
	require.NoError(t, err)
	require.Contains(t, string(wire), `"refusal":"can't help"`)
}

func TestV0215ProtocolRefusalStreamAndBuffer(t *testing.T) {
	for _, withText := range []bool{false, true} {
		t.Run(map[bool]string{false: "refusal_only", true: "mixed"}[withText], func(t *testing.T) {
			acc := NewBufferedResponseAccumulator()
			state := NewResponsesEventToChatState()
			events := []ResponsesStreamEvent{
				{Type: "response.created", Response: &ResponsesResponse{ID: "resp_test"}},
				{Type: "response.refusal.delta", Delta: "can't "},
				{Type: "response.refusal.delta", Delta: "help"},
				{Type: "response.refusal.done", Delta: "can't help"},
				{Type: "response.refusal.delta"},
			}
			if withText {
				events = append(events, ResponsesStreamEvent{Type: "response.output_text.delta", Delta: "normal"})
			}
			events = append(events, ResponsesStreamEvent{Type: "response.completed", Response: &ResponsesResponse{ID: "resp_test", Status: "completed"}})
			var refusal, text string
			for _, evt := range events {
				acc.ProcessEvent(&evt)
				for _, chunk := range ResponsesEventToChatChunks(&evt, state) {
					wire, err := json.Marshal(chunk)
					require.NoError(t, err)
					var decoded ChatCompletionsChunk
					require.NoError(t, json.Unmarshal(wire, &decoded))
					for _, choice := range decoded.Choices {
						if choice.Delta.Refusal != nil {
							refusal += *choice.Delta.Refusal
						}
						if choice.Delta.Content != nil {
							text += *choice.Delta.Content
						}
					}
				}
			}
			require.Equal(t, "can't help", refusal, "done must not duplicate deltas")
			require.True(t, acc.HasContent())
			final := &ResponsesResponse{ID: "resp_test"}
			acc.SupplementResponseOutput(final)
			chat := ResponsesToChatCompletions(final, "gpt-4o")
			require.Equal(t, refusal, chat.Choices[0].Message.Refusal)
			if withText {
				require.Equal(t, "normal", text)
				require.JSONEq(t, `"normal"`, string(chat.Choices[0].Message.Content))
			} else {
				require.Empty(t, text)
				require.Empty(t, chat.Choices[0].Message.Content)
			}
			existing := &ResponsesResponse{Output: []ResponsesOutput{{Type: "message", Content: []ResponsesContentPart{{Type: "refusal", Refusal: "terminal"}}}}}
			acc.SupplementResponseOutput(existing)
			require.Equal(t, "terminal", ResponsesToChatCompletions(existing, "gpt-4o").Choices[0].Message.Refusal)
		})
	}
}

func TestV0215ProtocolThinkingSignature(t *testing.T) {
	for _, tt := range []struct {
		name  string
		block AnthropicContentBlock
		want  string
	}{
		{"thinking_plan", AnthropicContentBlock{Type: "thinking", Thinking: "plan"}, `{"type":"thinking","thinking":"plan","signature":""}`},
		{"thinking_signature", AnthropicContentBlock{Type: "thinking", Signature: "sig"}, `{"type":"thinking","thinking":"","signature":"sig"}`},
		{"thinking_both_empty", AnthropicContentBlock{Type: "thinking"}, `{"type":"thinking","thinking":"","signature":""}`},
		{"text_nonempty", AnthropicContentBlock{Type: "text", Text: "hi"}, `{"type":"text","text":"hi"}`},
		{"text_empty_required", AnthropicContentBlock{Type: "text"}, `{"type":"text","text":""}`},
		{"text_cache_control_ttl", AnthropicContentBlock{Type: "text", CacheControl: &AnthropicCacheControl{Type: "ephemeral", TTL: "1h"}}, `{"type":"text","text":"","cache_control":{"type":"ephemeral","ttl":"1h"}}`},
		{"thinking_cache_control", AnthropicContentBlock{Type: "thinking", CacheControl: &AnthropicCacheControl{Type: "ephemeral"}}, `{"type":"thinking","thinking":"","signature":"","cache_control":{"type":"ephemeral"}}`},
		{"tool_use_no_thinking_fields", AnthropicContentBlock{Type: "tool_use", ID: "call", Name: "lookup", Input: json.RawMessage(`{}`)}, `{"type":"tool_use","id":"call","name":"lookup","input":{}}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			wire, err := json.Marshal(tt.block)
			require.NoError(t, err)
			require.JSONEq(t, tt.want, string(wire))
		})
	}
	state := NewResponsesEventToAnthropicState()
	events := ResponsesEventToAnthropicEvents(&ResponsesStreamEvent{Type: "response.output_item.added", OutputIndex: 0, Item: &ResponsesOutput{Type: "reasoning"}}, state)
	require.NotEmpty(t, events)
	var sawStart bool
	for _, evt := range events {
		if evt.Type == "content_block_start" {
			wire, err := ResponsesAnthropicEventToSSE(evt)
			require.NoError(t, err)
			require.Contains(t, wire, `"signature":""`)
			sawStart = true
		}
	}
	require.True(t, sawStart)
}
