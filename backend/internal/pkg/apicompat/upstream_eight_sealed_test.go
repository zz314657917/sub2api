package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpstreamEightSealedArguments(t *testing.T) {
	for _, raw := range []string{"", "null", "[]", `{"ciphertext":"密文","version":1}`, `["sealed"]`} {
		t.Run("value="+raw, func(t *testing.T) {
			suffix := ""
			if raw != "" {
				suffix = `,"encrypted_function_args":` + raw
			}
			var item ResponsesOutput
			require.NoError(t, json.Unmarshal([]byte(`{"type":"function_call","call_id":"call_1","name":"read","namespace":"agent","arguments":""`+suffix+`}`), &item))
			acc := NewBufferedResponseAccumulator()
			acc.ProcessEvent(&ResponsesStreamEvent{Type: "response.output_item.added", Item: &item})
			acc.ProcessEvent(&ResponsesStreamEvent{Type: "response.function_call_arguments.delta", Delta: `{"path":"x"}`})
			output := acc.BuildOutput()
			require.Len(t, output, 1)
			require.Equal(t, "agent", output[0].Namespace)
			require.Equal(t, `{"path":"x"}`, output[0].Arguments)
			for _, value := range []any{item, output[0]} {
				data, err := json.Marshal(value)
				require.NoError(t, err)
				checkUpstreamEightSealed(t, data, raw)
			}
			for _, kind := range []string{"response.output_item.added", "response.output_item.done"} {
				data, err := json.Marshal(ResponsesStreamEvent{Type: kind, Item: &output[0]})
				require.NoError(t, err)
				var wire map[string]json.RawMessage
				require.NoError(t, json.Unmarshal(data, &wire))
				checkUpstreamEightSealed(t, wire["item"], raw)
			}
		})
	}
}

func checkUpstreamEightSealed(t *testing.T, data []byte, want string) {
	t.Helper()
	var obj map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &obj))
	require.JSONEq(t, `"agent"`, string(obj["namespace"]))
	value, exists := obj["encrypted_function_args"]
	require.Equal(t, want != "", exists, "absence must remain distinct from explicit null")
	if exists {
		require.JSONEq(t, want, string(value))
	}
}
