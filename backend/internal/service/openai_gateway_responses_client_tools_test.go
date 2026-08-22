package service

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"net/http/httptest"
)

func TestAdaptOpenAIResponsesClientToolsRejectsTrailingJSON(t *testing.T) {
	body := []byte(`{"model":"gpt-5.6","tools":[{"type":"custom","name":"exec"}]} {"extra":true}`)
	adapted, mapping, err := adaptOpenAIResponsesClientTools(body)
	require.ErrorContains(t, err, "trailing data")
	require.Equal(t, body, adapted)
	require.Empty(t, mapping.CustomTools)
}

func TestAdaptOpenAIResponsesClientToolsLeavesNamespaceOnlyUnchanged(t *testing.T) {
	body := []byte(`{"model":"gpt-5.6","tools":[{"type":"namespace","name":"mcp","tools":[{"type":"function","name":"run"}]}]}`)
	adapted, mapping, err := adaptOpenAIResponsesClientTools(body)
	require.NoError(t, err)
	require.Equal(t, body, adapted)
	require.Empty(t, mapping.CustomTools)
}

func TestOpenAIPassthroughAPIKeyRestoresClientToolsStreaming(t *testing.T) {
	sse := strings.Join([]string{
		`data: {"type":"response.output_item.added","sequence_number":0,"output_index":0,"item":{"type":"function_call","id":"fc_1","call_id":"call_1","name":"exec","arguments":""}}`,
		`data: {"type":"response.function_call_arguments.done","sequence_number":1,"output_index":0,"item_id":"fc_1","call_id":"call_1","name":"exec","arguments":"{\"input\":\"pwd\"}"}`,
		`data: {"type":"response.output_item.done","sequence_number":2,"output_index":0,"item":{"type":"function_call","id":"fc_1","call_id":"call_1","name":"exec","arguments":"{\"input\":\"pwd\"}"}}`,
		`data: {"type":"response.completed","sequence_number":3,"response":{"output":[{"type":"function_call","id":"fc_1","call_id":"call_1","name":"exec","arguments":"{\"input\":\"pwd\"}"}]}}`,
		"",
	}, "\n")
	body := newResponsesClientToolStreamBody(io.NopCloser(strings.NewReader(sse)), apicompat.ResponsesClientToolMapping{CustomTools: map[string]bool{"exec": true}}, 1<<20)
	got, err := io.ReadAll(body)
	require.NoError(t, err)
	require.NoError(t, body.Close())
	require.Contains(t, string(got), `"type":"custom_tool_call"`)
	require.Contains(t, string(got), `"type":"response.custom_tool_call_input.done"`)
	require.Contains(t, string(got), `"input":"pwd"`)
	require.NotContains(t, string(got), `"input":{`)
}

func TestOpenAIPassthroughAPIKeyRestoresClientToolsNonStreaming(t *testing.T) {
	payload := []byte(`{"id":"r1","output":[{"type":"function_call","call_id":"c1","name":"exec","arguments":"{\"input\":\"pwd\"}"}]}`)
	restored, changed, err := apicompat.RestoreResponsesClientToolPayload(payload, apicompat.ResponsesClientToolMapping{CustomTools: map[string]bool{"exec": true}})
	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "custom_tool_call", gjson.GetBytes(restored, "output.0.type").String())
	require.Equal(t, "pwd", gjson.GetBytes(restored, "output.0.input").String())
}

func TestClearOpenAIResponsesClientToolMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(openAIResponsesClientToolMappingContextKey, apicompat.ResponsesClientToolMapping{CustomTools: map[string]bool{"exec": true}})
	clearOpenAIResponsesClientToolMapping(c)
	_, ok := openAIResponsesClientToolMapping(c)
	require.False(t, ok)
}

func TestAdaptOpenAIResponsesClientToolsNoMutationOnOrdinaryFunction(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","name":"run"}],"input":"hi"}`)
	adapted, mapping, err := adaptOpenAIResponsesClientTools(bytes.Clone(body))
	require.NoError(t, err)
	require.Equal(t, body, adapted)
	require.Empty(t, mapping.CustomTools)
}
