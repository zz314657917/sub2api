package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponsesToAnthropic_CustomToolNormalizesToObjectSchema(t *testing.T) {
	req := &ResponsesRequest{
		Model: "gpt-5.2",
		Input: json.RawMessage(`[{"role":"user","content":"Hello"}]`),
		Tools: []ResponsesTool{{
			Type:        "custom",
			Name:        "custom_shell",
			Description: "Run shell command",
			Parameters:  json.RawMessage(`{"type":"string"}`),
		}},
	}

	resp, err := ResponsesToAnthropicRequest(req)
	require.NoError(t, err)
	require.Len(t, resp.Tools, 1)
	require.Empty(t, resp.Tools[0].Type)
	require.Equal(t, "custom_shell", resp.Tools[0].Name)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(resp.Tools[0].InputSchema, &schema))
	require.Equal(t, "object", schema["type"])
	require.Contains(t, schema, "properties")
}

func TestResponsesToAnthropic_UnknownToolNormalizesInvalidSchema(t *testing.T) {
	req := &ResponsesRequest{
		Model: "gpt-5.2",
		Input: json.RawMessage(`[{"role":"user","content":"Hello"}]`),
		Tools: []ResponsesTool{{
			Type:       "local_shell",
			Name:       "run",
			Parameters: json.RawMessage(`not-json`),
		}},
	}

	resp, err := ResponsesToAnthropicRequest(req)
	require.NoError(t, err)
	require.Len(t, resp.Tools, 1)
	require.Equal(t, "local_shell", resp.Tools[0].Type)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(resp.Tools[0].InputSchema, &schema))
	require.Equal(t, "object", schema["type"])
	require.Contains(t, schema, "properties")
}
