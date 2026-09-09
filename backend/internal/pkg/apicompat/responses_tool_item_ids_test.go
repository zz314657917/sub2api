package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAnthropicResponsesToolItemIDs(t *testing.T) {
	block := AnthropicContentBlock{Type: "tool_use", ID: "toolu_123", Name: "exec", Input: json.RawMessage(`{}`)}
	response := AnthropicToResponsesResponse(&AnthropicResponse{Content: []AnthropicContentBlock{block}})
	require.Len(t, response.Output, 1)
	require.Regexp(t, `^fc_[a-f0-9]{24}$`, response.Output[0].ID)
	state := NewAnthropicEventToResponsesState()
	events := AnthropicEventToResponsesEvents(&AnthropicStreamEvent{Type: "message_start"}, state)
	events = append(events, AnthropicEventToResponsesEvents(&AnthropicStreamEvent{Type: "content_block_start", ContentBlock: &block}, state)...)
	events = append(events, AnthropicEventToResponsesEvents(&AnthropicStreamEvent{Type: "content_block_stop"}, state)...)
	events = append(events, FinalizeAnthropicResponsesStream(state)...)
	assertToolItemEventIDs(t, events, "fc_", response.Output[0].CallID)
}

func TestChatResponsesToolItemIDs(t *testing.T) {
	for _, tc := range []struct{ name, prefix string }{
		{"exec", "fc_"}, {"custom", "ctc_"}, {toolSearchProxyName, "tsc_"}, {"ns_exec", "fc_"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			custom := map[string]bool{"custom": true}
			namespaces := map[string]NamespacedToolName{"ns_exec": {Name: "exec", Namespace: "ns"}}
			call := ChatToolCall{ID: "call_123", Type: "function"}
			call.Function.Name = tc.name
			call.Function.Arguments = `{}`
			outputs := chatMessageToResponsesOutput(ChatMessage{ToolCalls: []ChatToolCall{call}}, custom, true, namespaces)
			require.Len(t, outputs, 1)
			require.Regexp(t, "^"+tc.prefix+`[a-f0-9]{24}$`, outputs[0].ID)
			require.Equal(t, call.ID, outputs[0].CallID)
			state := NewChatCompletionsToResponsesStreamState("test")
			state.CustomTools, state.ToolSearchDeclared, state.NamespaceTools = custom, true, namespaces
			var chunk ChatCompletionsChunk
			require.NoError(t, json.Unmarshal([]byte(`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_123","type":"function","function":{"arguments":""}}]}}]}`), &chunk))
			events := ChatCompletionsChunkToResponsesEvents(&chunk, state)
			// A later chunk supplies the name; no item may be announced in the wrong namespace.
			chunk.Choices[0].Delta.ToolCalls[0] = call
			events = append(events, ChatCompletionsChunkToResponsesEvents(&chunk, state)...)
			events = append(events, FinalizeChatCompletionsResponsesStream(state)...)
			assertToolItemEventIDs(t, events, tc.prefix, call.ID)
		})
	}
}

func assertToolItemEventIDs(t *testing.T, events []ResponsesStreamEvent, prefix, callID string) {
	t.Helper()
	id := ""
	completed := false
	for _, event := range events {
		if event.Item != nil {
			if id == "" {
				id = event.Item.ID
			}
			require.Equal(t, id, event.Item.ID)
			require.Equal(t, callID, event.Item.CallID)
		}
		if event.ItemID != "" {
			require.Equal(t, id, event.ItemID)
		}
		if event.Type == "response.completed" {
			completed = true
			require.Len(t, event.Response.Output, 1)
			require.Equal(t, id, event.Response.Output[0].ID)
			require.Equal(t, callID, event.Response.Output[0].CallID)
		}
	}
	require.Regexp(t, "^"+prefix+`[a-f0-9]{24}$`, id)
	require.True(t, completed)
}
