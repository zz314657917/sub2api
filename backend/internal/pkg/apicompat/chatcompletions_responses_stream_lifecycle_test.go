package apicompat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStream_InvalidToolArgumentsAreRejectedBeforeFinalize(t *testing.T) {
	idx := 0
	state := NewChatCompletionsToResponsesStreamState("deepseek-v4-flash")
	chunk := &ChatCompletionsChunk{
		Choices: []ChatChunkChoice{{
			Index: 0,
			Delta: ChatDelta{
				ToolCalls: []ChatToolCall{{
					Index: &idx,
					ID:    "call_bad",
					Type:  "function",
					Function: ChatFunctionCall{
						Name:      "exec_command",
						Arguments: `{"cmd": "ssh root@HOST`,
					},
				}},
			},
		}},
	}
	ChatCompletionsChunkToResponsesEvents(chunk, state)

	err := state.ValidateToolCallArguments()
	require.ErrorContains(t, err, "invalid JSON")
}

func TestStream_ValidToolCallAtOutputLimitKeepsIncompleteResponse(t *testing.T) {
	idx := 0
	state := NewChatCompletionsToResponsesStreamState("deepseek-v4-flash")
	chunk := &ChatCompletionsChunk{
		Choices: []ChatChunkChoice{{
			Index: 0,
			Delta: ChatDelta{
				ToolCalls: []ChatToolCall{{
					Index: &idx,
					ID:    "call_at_limit",
					Type:  "function",
					Function: ChatFunctionCall{
						Name:      "exec_command",
						Arguments: `{}`,
					},
				}},
			},
		}},
	}
	ChatCompletionsChunkToResponsesEvents(chunk, state)
	state.FinishReason = "length"

	require.NoError(t, state.ValidateToolCallArguments())
	var sawArgsDone, sawIncomplete bool
	for _, event := range FinalizeChatCompletionsResponsesStream(state) {
		switch event.Type {
		case "response.function_call_arguments.done":
			sawArgsDone = true
			require.Equal(t, `{}`, event.Arguments)
		case "response.completed":
			require.NotNil(t, event.Response)
			sawIncomplete = event.Response.Status == "incomplete"
		}
	}
	require.True(t, sawArgsDone)
	require.True(t, sawIncomplete)
}

func lifecycleStringPtr(value string) *string { return &value }

func runResponsesLifecycle(state *ChatCompletionsToResponsesStreamState, chunks ...*ChatCompletionsChunk) []ResponsesStreamEvent {
	var events []ResponsesStreamEvent
	for _, chunk := range chunks {
		events = append(events, ChatCompletionsChunkToResponsesEvents(chunk, state)...)
	}
	return append(events, FinalizeChatCompletionsResponsesStream(state)...)
}

func requireStreamIndicesMatchTerminal(t *testing.T, events []ResponsesStreamEvent) *ResponsesResponse {
	t.Helper()
	var terminal *ResponsesResponse
	for index := range events {
		if events[index].Type == "response.completed" {
			terminal = events[index].Response
		}
	}
	require.NotNil(t, terminal, "response.completed missing")

	for _, event := range events {
		expectedType := ""
		switch event.Type {
		case "response.output_item.added", "response.output_item.done":
			require.NotNil(t, event.Item)
			expectedType = event.Item.Type
		case "response.reasoning_summary_part.added", "response.reasoning_summary_text.delta",
			"response.reasoning_summary_text.done", "response.reasoning_summary_part.done":
			expectedType = "reasoning"
		case "response.content_part.added", "response.output_text.delta",
			"response.output_text.done", "response.content_part.done":
			expectedType = "message"
		case "response.function_call_arguments.delta", "response.function_call_arguments.done":
			expectedType = "function_call"
		case "response.custom_tool_call_input.delta", "response.custom_tool_call_input.done":
			expectedType = "custom_tool_call"
		default:
			continue
		}
		require.GreaterOrEqual(t, event.OutputIndex, 0, event.Type)
		require.Less(t, event.OutputIndex, len(terminal.Output), event.Type)
		require.Equal(t, expectedType, terminal.Output[event.OutputIndex].Type, event.Type)
		if event.Item != nil && event.Item.ID != "" {
			require.Equal(t, event.Item.ID, terminal.Output[event.OutputIndex].ID, event.Type)
		}
	}
	return terminal
}

func toolCallChunk(index int, id, name, arguments string) *ChatCompletionsChunk {
	return &ChatCompletionsChunk{Choices: []ChatChunkChoice{{Delta: ChatDelta{ToolCalls: []ChatToolCall{{
		Index: &index,
		ID:    id,
		Type:  "function",
		Function: ChatFunctionCall{
			Name:      name,
			Arguments: arguments,
		},
	}}}}}}
}

func TestStreamLifecycle_ToolOnlyIndicesMatchTerminalForEveryToolKind(t *testing.T) {
	tests := []struct {
		name          string
		configure     func(*ChatCompletionsToResponsesStreamState)
		toolName      string
		arguments     string
		wantType      string
		wantName      string
		wantNamespace string
	}{
		{
			name:      "ordinary function",
			toolName:  "wait",
			arguments: `{"cell_id":3}`,
			wantType:  "function_call",
			wantName:  "wait",
		},
		{
			name: "custom tool",
			configure: func(state *ChatCompletionsToResponsesStreamState) {
				state.CustomTools = map[string]bool{"exec": true}
			},
			toolName:  "exec",
			arguments: `{"input":"dir"}`,
			wantType:  "custom_tool_call",
			wantName:  "exec",
		},
		{
			name: "tool search",
			configure: func(state *ChatCompletionsToResponsesStreamState) {
				state.ToolSearchDeclared = true
			},
			toolName:  toolSearchProxyName,
			arguments: `{"query":"gmail"}`,
			wantType:  "tool_search_call",
		},
		{
			name: "namespace tool",
			configure: func(state *ChatCompletionsToResponsesStreamState) {
				state.NamespaceTools = map[string]NamespacedToolName{
					"mcp__svc__echo": {Namespace: "mcp__svc", Name: "echo"},
				}
			},
			toolName:      "mcp__svc__echo",
			arguments:     `{"text":"hi"}`,
			wantType:      "function_call",
			wantName:      "echo",
			wantNamespace: "mcp__svc",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewChatCompletionsToResponsesStreamState("gpt-5")
			if test.configure != nil {
				test.configure(state)
			}
			events := runResponsesLifecycle(state, toolCallChunk(0, "call_1", test.toolName, test.arguments))
			terminal := requireStreamIndicesMatchTerminal(t, events)
			require.Len(t, terminal.Output, 1)
			require.Equal(t, test.wantType, terminal.Output[0].Type)
			require.Equal(t, test.wantName, terminal.Output[0].Name)
			require.Equal(t, test.wantNamespace, terminal.Output[0].Namespace)

			for _, event := range events {
				if event.Type == "response.output_item.added" || event.Type == "response.output_item.done" {
					require.Equal(t, 0, event.OutputIndex, event.Type)
				}
			}
		})
	}
}

func TestStreamLifecycle_ReasoningClosesBeforeToolAndIndicesMatchTerminal(t *testing.T) {
	state := NewChatCompletionsToResponsesStreamState("gpt-5")
	events := runResponsesLifecycle(
		state,
		&ChatCompletionsChunk{Choices: []ChatChunkChoice{{Delta: ChatDelta{ReasoningContent: lifecycleStringPtr("plan")}}}},
		toolCallChunk(0, "call_1", "wait", `{"cell_id":3}`),
	)
	terminal := requireStreamIndicesMatchTerminal(t, events)
	require.Len(t, terminal.Output, 2)
	require.Equal(t, "reasoning", terminal.Output[0].Type)
	require.Equal(t, "function_call", terminal.Output[1].Type)

	positions := map[string]int{}
	for index, event := range events {
		if event.Type == "response.output_item.added" && event.Item != nil {
			positions["added:"+event.Item.Type] = index
			require.Equal(t, map[string]int{"reasoning": 0, "function_call": 1}[event.Item.Type], event.OutputIndex)
		}
		if event.Type == "response.reasoning_summary_text.delta" {
			positions["reasoning:delta"] = index
		}
		if event.Type == "response.output_item.done" && event.Item != nil && event.Item.Type == "reasoning" {
			positions["done:reasoning"] = index
		}
	}
	for _, key := range []string{"added:reasoning", "reasoning:delta", "done:reasoning", "added:function_call"} {
		require.Contains(t, positions, key)
	}
	require.Less(t, positions["added:reasoning"], positions["reasoning:delta"])
	require.Less(t, positions["reasoning:delta"], positions["done:reasoning"])
	require.Less(t, positions["done:reasoning"], positions["added:function_call"])
}

func TestStreamLifecycle_MessageContentPartUsesStableDynamicIndex(t *testing.T) {
	state := NewChatCompletionsToResponsesStreamState("gpt-5")
	events := runResponsesLifecycle(state, &ChatCompletionsChunk{Choices: []ChatChunkChoice{{
		Delta: ChatDelta{Content: lifecycleStringPtr("hello")},
	}}})
	terminal := requireStreamIndicesMatchTerminal(t, events)
	require.Len(t, terminal.Output, 1)
	require.Equal(t, "message", terminal.Output[0].Type)
	require.Equal(t, "hello", terminal.Output[0].Content[0].Text)

	wantOrder := []string{
		"response.output_item.added",
		"response.content_part.added",
		"response.output_text.delta",
		"response.output_text.done",
		"response.content_part.done",
		"response.output_item.done",
	}
	var gotOrder []string
	for _, event := range events {
		for _, wanted := range wantOrder {
			if event.Type == wanted {
				gotOrder = append(gotOrder, event.Type)
				require.Equal(t, 0, event.OutputIndex, event.Type)
			}
		}
	}
	require.Equal(t, wantOrder, gotOrder)
}

func TestStreamLifecycle_LateCustomNameKeepsAllocatedIndex(t *testing.T) {
	state := NewChatCompletionsToResponsesStreamState("gpt-5")
	state.CustomTools = map[string]bool{"exec": true}
	events := runResponsesLifecycle(
		state,
		toolCallChunk(0, "call_1", "", `{"inp`),
		toolCallChunk(0, "", "exec", `ut":"dir"}`),
	)
	terminal := requireStreamIndicesMatchTerminal(t, events)
	require.Len(t, terminal.Output, 1)
	require.Equal(t, "custom_tool_call", terminal.Output[0].Type)
	require.Equal(t, "dir", terminal.Output[0].Input)
}

func TestStreamLifecycle_ParallelToolCallsUseOpenOrderIndices(t *testing.T) {
	first, second := 0, 1
	state := NewChatCompletionsToResponsesStreamState("gpt-5")
	chunk := &ChatCompletionsChunk{
		Choices: []ChatChunkChoice{{
			Delta: ChatDelta{ToolCalls: []ChatToolCall{
				{Index: &first, ID: "call_1", Type: "function", Function: ChatFunctionCall{Name: "first", Arguments: `{}`}},
				{Index: &second, ID: "call_2", Type: "function", Function: ChatFunctionCall{Name: "second", Arguments: `{}`}},
			}},
		}},
	}
	events := runResponsesLifecycle(state, chunk)
	terminal := requireStreamIndicesMatchTerminal(t, events)
	require.Len(t, terminal.Output, 2)
	require.Equal(t, "first", terminal.Output[0].Name)
	require.Equal(t, "second", terminal.Output[1].Name)
}
