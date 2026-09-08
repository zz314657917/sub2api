package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertCodexCatalogPreservesWorkflowMetadata(t *testing.T) {
	for _, override := range []string{`"high"`, `null`} {
		body := []byte(`{"data":[{"id":"gpt-6-astra","multi_agent_reasoning_effort":` + override + `,"multi_agent_version":"v2","supported_reasoning_levels":[{"effort":"high"},{"effort":"ultra"}],"use_responses_lite":true}]}`)
		converted := convertOpenAIModelListToCodexManifest(body)
		adjusted, err := adjustAPIKeyCodexModelsManifest(converted)
		require.NoError(t, err)
		var envelope struct {
			Models []map[string]json.RawMessage `json:"models"`
		}
		require.NoError(t, json.Unmarshal(adjusted, &envelope))
		require.Len(t, envelope.Models, 1)
		model := envelope.Models[0]
		require.JSONEq(t, `"gpt-6-astra"`, string(model["slug"]))
		require.Equal(t, override, string(model["multi_agent_reasoning_effort"]))
		require.JSONEq(t, `"v2"`, string(model["multi_agent_version"]))
		require.JSONEq(t, `[{"effort":"high"},{"effort":"ultra"}]`, string(model["supported_reasoning_levels"]))
		require.Equal(t, "false", string(model["use_responses_lite"]))
	}
}
