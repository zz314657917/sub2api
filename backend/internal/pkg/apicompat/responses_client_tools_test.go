package apicompat

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdaptResponsesClientTools_CustomOnly(t *testing.T) {
	req := map[string]any{
		"tools": []any{
			map[string]any{"type": "custom", "name": "exec", "format": map[string]any{"type": "text"}},
			map[string]any{"type": "function", "name": "ordinary", "parameters": map[string]any{"type": "object"}},
			map[string]any{"type": "namespace", "name": "mcp", "tools": []any{map[string]any{"type": "function", "name": "run"}}},
		},
		"tool_choice": map[string]any{"type": "custom", "name": "exec"},
		"input":       []any{map[string]any{"type": "custom_tool_call", "id": "ctc_1", "call_id": "call_1", "name": "exec", "input": "pwd"}, map[string]any{"type": "custom_tool_call_output", "id": "ctco_1", "call_id": "call_1", "output": map[string]any{"ok": true}}},
	}
	mapping, changed, err := AdaptResponsesClientTools(req)
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, map[string]bool{"exec": true}, mapping.CustomTools)
	tools := req["tools"].([]any)
	require.Equal(t, "function", tools[0].(map[string]any)["type"])
	require.Equal(t, "ordinary", tools[1].(map[string]any)["name"])
	require.Equal(t, "namespace", tools[2].(map[string]any)["type"])
	require.Equal(t, "function", req["tool_choice"].(map[string]any)["type"])
	call := req["input"].([]any)[0].(map[string]any)
	require.Equal(t, "function_call", call["type"])
	require.NotContains(t, call, "id")
	require.JSONEq(t, `{"input":"pwd"}`, call["arguments"].(string))
	out := req["input"].([]any)[1].(map[string]any)
	require.Equal(t, "function_call_output", out["type"])
	require.NotContains(t, out, "id")
	require.Equal(t, `{"ok":true}`, out["output"])
}

func TestAdaptResponsesClientTools_NamespaceAndFunctionNoop(t *testing.T) {
	req := map[string]any{"tools": []any{map[string]any{"type": "namespace", "name": "mcp"}, map[string]any{"type": "function", "name": "run"}}}
	before, _ := json.Marshal(req)
	mapping, changed, err := AdaptResponsesClientTools(req)
	require.NoError(t, err)
	require.False(t, changed)
	require.Empty(t, mapping.CustomTools)
	after, _ := json.Marshal(req)
	require.JSONEq(t, string(before), string(after))
}

func TestAdaptResponsesClientToolsWithInheritedMapping(t *testing.T) {
	req := map[string]any{"input": []any{map[string]any{"type": "custom_tool_call", "call_id": "c1", "name": "exec", "input": "ls"}}}
	mapping, changed, err := AdaptResponsesClientToolsWithInheritedMapping(req, ResponsesClientToolMapping{CustomTools: map[string]bool{"exec": true}})
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "function_call", req["input"].([]any)[0].(map[string]any)["type"])
	require.Equal(t, mapping.CustomTools, map[string]bool{"exec": true})

	explicit := map[string]any{"tools": []any{}, "input": []any{map[string]any{"type": "custom_tool_call", "name": "exec", "input": "ls"}}}
	_, changed, err = AdaptResponsesClientToolsWithInheritedMapping(explicit, mapping)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, "custom_tool_call", explicit["input"].([]any)[0].(map[string]any)["type"])
}

func TestRestoreResponsesClientToolPayload_CustomCall(t *testing.T) {
	payload := []byte(`{"id":"r1","output":[{"type":"function_call","id":"fc_1","call_id":"c1","name":"exec","arguments":"{\"input\":\"pwd\"}"},{"type":"function_call","name":"ordinary","arguments":"{}"}]}`)
	restored, changed, err := RestoreResponsesClientToolPayload(payload, ResponsesClientToolMapping{CustomTools: map[string]bool{"exec": true}})
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "custom_tool_call", gjsonGet(restored, "output.0.type"))
	require.Equal(t, "pwd", gjsonGet(restored, "output.0.input"))
	require.Equal(t, "function_call", gjsonGet(restored, "output.1.type"))
}

func TestResponsesClientToolStreamRestorer(t *testing.T) {
	restorer := NewResponsesClientToolStreamRestorer(ResponsesClientToolMapping{CustomTools: map[string]bool{"exec": true}})
	added, changed, err := restorer.RestoreEvent([]byte(`{"type":"response.output_item.added","sequence_number":0,"output_index":0,"item":{"type":"function_call","id":"fc_1","call_id":"c1","name":"exec","arguments":""}}`))
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "custom_tool_call", gjsonGet(added[0], "item.type"))
	_, _, err = restorer.RestoreEvent([]byte(`{"type":"response.function_call_arguments.delta","sequence_number":1,"item_id":"fc_1","delta":"{\"input\":\"pwd\"}"}`))
	require.NoError(t, err)
	done, changed, err := restorer.RestoreEvent([]byte(`{"type":"response.function_call_arguments.done","sequence_number":2,"item_id":"fc_1","call_id":"c1","name":"exec","arguments":"{\"input\":\"pwd\"}"}`))
	require.NoError(t, err)
	require.True(t, changed)
	require.Len(t, done, 2)
	require.Equal(t, "response.custom_tool_call_input.done", gjsonGet(done[1], "type"))
	require.Equal(t, "pwd", gjsonGet(done[1], "input"))
}

func gjsonGet(payload []byte, path string) string {
	var root any
	_ = json.Unmarshal(payload, &root)
	parts := strings.Split(path, ".")
	for _, part := range parts {
		switch value := root.(type) {
		case map[string]any:
			root = value[part]
		case []any:
			var index int
			_, _ = fmt.Sscanf(part, "%d", &index)
			if index < 0 || index >= len(value) {
				return ""
			}
			root = value[index]
		default:
			return ""
		}
	}
	if text, ok := root.(string); ok {
		return text
	}
	return ""
}
