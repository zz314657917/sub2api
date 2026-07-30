package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSanitizeGrokResponsesToolChoiceRequiresNonEmptyArray(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		wantToolChoice bool
	}{
		{name: "missing tools", body: `{"input":"hello","tool_choice":"auto"}`},
		{name: "null tools", body: `{"input":"hello","tools":null,"tool_choice":"auto"}`},
		{name: "empty tools", body: `{"input":"hello","tools":[],"tool_choice":"auto"}`},
		{
			name:           "non-empty tools",
			body:           `{"input":"hello","tools":[{"type":"function","name":"lookup"}],"tool_choice":"auto"}`,
			wantToolChoice: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patched, err := sanitizeGrokResponsesToolChoice([]byte(tt.body))
			require.NoError(t, err)
			require.True(t, json.Valid(patched))
			require.Equal(t, tt.wantToolChoice, gjson.GetBytes(patched, "tool_choice").Exists())
		})
	}
}

func TestOpenAIModelCapacityRetryLimit(t *testing.T) {
	capacityMessage := "Selected model is at capacity. Please try a different model."
	require.Equal(t, openAIModelCapacitySameAccountRetryLimit, openAIModelCapacityRetryLimit(capacityMessage, nil))
	require.Equal(t, openAIModelCapacitySameAccountRetryLimit, openAIModelCapacityRetryLimit("", []byte(`{"error":{"message":"Selected model is at capacity"}}`)))
	require.Zero(t, openAIModelCapacityRetryLimit("temporary upstream failure", []byte(`{"error":{"message":"temporary upstream failure"}}`)))
}
