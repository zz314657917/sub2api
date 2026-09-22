package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV027DeepSeekLiftResponsesToolOutputMediaKeepsParallelBatchContiguous(t *testing.T) {
	var input any
	require.NoError(t, json.Unmarshal([]byte(`[
		{"type":"function_call","call_id":"call_A"},
		{"type":"function_call","call_id":"call_B"},
		{"type":"function_call_output","call_id":"call_A","output":[{"type":"input_image","image_url":"data:image/png;base64,QQ=="},{"type":"input_text","text":"A"}]},
		{"type":"message","role":"developer","content":"notice A"},
		{"type":"function_call_output","call_id":"call_B","output":[{"type":"input_image","image_url":{"url":"data:image/png;base64,Qg=="}},{"type":"input_text","text":"B"}]},
		{"type":"message","role":"system","content":"notice B"}
	]`), &input))

	lifted, changed := LiftResponsesToolOutputMedia(input)
	require.True(t, changed)
	items := lifted.([]any)
	require.Len(t, items, 7)
	for index, callID := range []string{"call_A", "call_B"} {
		item := items[index+2].(map[string]any)
		require.Equal(t, callID, item["call_id"])
		require.IsType(t, "", item["output"])
		require.NotContains(t, item["output"], "data:image")
	}
	media := items[4].(map[string]any)
	require.Equal(t, "user", media["role"])
	parts := media["content"].([]map[string]any)
	require.Equal(t, "data:image/png;base64,QQ==", parts[1]["image_url"])
	require.Equal(t, "data:image/png;base64,Qg==", parts[3]["image_url"])
	require.Equal(t, "developer", items[5].(map[string]any)["role"])
	require.Equal(t, "system", items[6].(map[string]any)["role"])

	again, changedAgain := LiftResponsesToolOutputMedia(lifted)
	require.False(t, changedAgain)
	require.Equal(t, lifted, again)
}

func TestV027DeepSeekLiftResponsesToolOutputMediaLeavesPlainOutputUntouched(t *testing.T) {
	input := []any{map[string]any{"type": "function_call_output", "call_id": "call_text", "output": "plain output"}}
	lifted, changed := LiftResponsesToolOutputMedia(input)
	require.False(t, changed)
	require.Equal(t, input, lifted)
}

func TestV027DeepSeekLiftResponsesToolOutputMediaPreservesUserBoundary(t *testing.T) {
	input := []any{
		map[string]any{"type": "function_call_output", "call_id": "plain", "output": "plain"},
		map[string]any{"type": "message", "role": "developer", "content": "notice"},
		map[string]any{"type": "message", "role": "user", "content": "next turn"},
		map[string]any{"type": "function_call_output", "call_id": "image", "output": []any{map[string]any{"type": "input_image", "image_url": "data:image/png;base64,AQ=="}}},
	}

	lifted, changed := LiftResponsesToolOutputMedia(input)
	require.True(t, changed)
	items := lifted.([]any)
	require.Len(t, items, 5)
	require.Equal(t, "plain", items[0].(map[string]any)["output"])
	require.Equal(t, "developer", items[1].(map[string]any)["role"])
	require.Equal(t, "user", items[2].(map[string]any)["role"])
	require.Equal(t, "function_call_output", items[3].(map[string]any)["type"])
	require.Equal(t, "user", items[4].(map[string]any)["role"])
}
