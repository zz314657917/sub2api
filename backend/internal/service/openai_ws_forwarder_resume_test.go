package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestOpenAIWSCurrentTurnRetryPayloadRebuildsClientPayload(t *testing.T) {
	payload := []byte(`{"type":"response.create","model":"mapped-model","previous_response_id":"resp_old","input":[{"role":"user","content":"continue"}]}`)
	fullInput := []json.RawMessage{
		json.RawMessage(`{"role":"user","content":"first"}`),
		json.RawMessage(`{"type":"function_call","id":"fc_1","call_id":"call_1","name":"inspect","arguments":"{}"}`),
		json.RawMessage(`{"type":"function_call_output","call_id":"call_1","output":"done"}`),
		json.RawMessage(`{"role":"user","content":"continue"}`),
	}

	retryPayload, retrySafe, err := buildOpenAIWSCurrentTurnRetryPayload(payload, fullInput, true, "client-model")

	require.NoError(t, err)
	require.True(t, retrySafe)
	require.False(t, gjson.GetBytes(retryPayload, "previous_response_id").Exists())
	require.Equal(t, "client-model", gjson.GetBytes(retryPayload, "model").String())
	require.True(t, gjson.GetBytes(retryPayload, "input").IsArray())
	require.Len(t, gjson.GetBytes(retryPayload, "input").Array(), len(fullInput))
	retryPayload[0] = 'x'
	require.Equal(t, byte('{'), payload[0], "retry payload must be cloned")
}

func TestOpenAIWSCurrentTurnRetryPayloadRejectsOrphanToolOutput(t *testing.T) {
	payload := []byte(`{"type":"response.create","model":"mapped-model","previous_response_id":"resp_old"}`)
	fullInput := []json.RawMessage{
		json.RawMessage(`{"type":"function_call_output","call_id":"missing_call","output":"done"}`),
	}

	retryPayload, retrySafe, err := buildOpenAIWSCurrentTurnRetryPayload(payload, fullInput, true, "client-model")

	require.NoError(t, err)
	require.False(t, retrySafe)
	require.Nil(t, retryPayload)
}

func TestOpenAIWSToolCallReplayCollectorKeepsAllOutputItems(t *testing.T) {
	collector := &openAIWSToolCallReplayCollector{}
	collector.AddEvent("response.completed", []byte(`{"type":"response.completed","response":{"output":[{"id":"msg_1","type":"message","role":"assistant","content":[{"type":"output_text","text":"first"}]},{"id":"fc_1","type":"function_call","call_id":"call_1","name":"inspect","arguments":"{}"}]}}`))

	allItems := collector.AllItems()
	require.Len(t, allItems, 2)
	require.Equal(t, "message", gjson.GetBytes(allItems[0], "type").String())
	require.Equal(t, "function_call", gjson.GetBytes(allItems[1], "type").String())
	filtered := collector.Items()
	require.Len(t, filtered, 1)
	require.Equal(t, "function_call", gjson.GetBytes(filtered[0], "type").String())
}
