package apicompat

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDiscoveredToolsChatFallback(t *testing.T) {
	var req ResponsesRequest
	require.NoError(t, json.Unmarshal([]byte(`{"tools":[{"type":"tool_search"},{"type":"custom","name":"exec"}],"input":[{"type":"tool_search_call","call_id":"s","arguments":{}},{"type":"tool_search_output","call_id":"s","status":"completed","tools":[{"type":"custom","name":"exec"},{"type":"namespace","name":"workspace","tools":[{"type":"function","name":"read","parameters":{"type":"object"}}]}]}]}`), &req))
	tools, err := EffectiveResponsesTools(&req)
	require.NoError(t, err)
	require.Len(t, tools, 3)
	require.True(t, CustomToolNames(tools)["exec"])
	require.Equal(t, "workspace", NamespaceToolNames(tools)["workspace__read"].Namespace)
	chat, err := ResponsesToChatCompletionsRequest(&req)
	require.NoError(t, err)
	require.Len(t, chat.Tools, 3)
	require.Contains(t, string(chat.Messages[1].Content), "workspace")
	require.Len(t, req.Tools, 2)
}

func TestDiscoveredToolsRejectConflictingSchema(t *testing.T) {
	var req ResponsesRequest
	require.NoError(t, json.Unmarshal([]byte(`{"tools":[{"type":"tool_search"},{"type":"function","name":"read","parameters":{"type":"object"}}],"input":[{"type":"tool_search_output","status":"completed","tools":[{"type":"function","name":"read","parameters":{"type":"string"}}]}]}`), &req))
	_, err := EffectiveResponsesTools(&req)
	require.ErrorContains(t, err, "conflicts")
}
