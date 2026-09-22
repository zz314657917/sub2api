package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV027DeepSeekReasoning_OptInKeepsTurnBoundaries(t *testing.T) {
	req := &ResponsesRequest{Model: "deepseek-reasoner", Input: json.RawMessage(`[
		{"type":"reasoning","summary":[{"type":"summary_text","text":"first"}]},
		{"type":"function_call","call_id":"call_a","name":"a","arguments":"{}"},
		{"type":"function_call","call_id":"call_b","name":"b","arguments":"{}"},
		{"type":"function_call_output","call_id":"call_a","output":[{"type":"input_image","image_url":"data:image/png;base64,QQ=="}]},
		{"type":"function_call_output","call_id":"call_b","output":"ok"},
		{"type":"reasoning","summary":[{"type":"summary_text","text":"discard"}]},
		{"type":"function_call","call_id":"bad","name":"bad","arguments":"not-json"},
		{"type":"function_call_output","call_id":"bad","output":"ignored"},
		{"type":"reasoning","content":[{"type":"reasoning_text","text":"second"}]},
		{"type":"message","role":"assistant","content":"done"}
	]`)}

	out, err := ResponsesToChatCompletionsRequestWithDeepSeekReasoning(req)

	require.NoError(t, err)
	require.Len(t, out.Messages, 5)
	require.Equal(t, "assistant", out.Messages[0].Role)
	require.Equal(t, "first", out.Messages[0].ReasoningContent)
	require.Len(t, out.Messages[0].ToolCalls, 2)
	require.Equal(t, "tool", out.Messages[1].Role)
	require.Equal(t, "tool", out.Messages[2].Role)
	require.Equal(t, "user", out.Messages[3].Role, "media follows the completed parallel tool batch")
	require.Equal(t, "assistant", out.Messages[4].Role)
	require.Equal(t, "second", out.Messages[4].ReasoningContent, "discarded invalid-call turn must not leak reasoning into the next turn")
}

func TestV027DeepSeekReasoning_DefaultBridgeRemainsUnchanged(t *testing.T) {
	req := &ResponsesRequest{Model: "deepseek-reasoner", Input: json.RawMessage(`[
		{"type":"reasoning","summary":[{"type":"summary_text","text":"hidden"}]},
		{"type":"message","role":"assistant","reasoning_content":"explicit","content":"done"}
	]`)}

	defaultOut, err := ResponsesToChatCompletionsRequest(req)
	require.NoError(t, err)
	optInOut, err := ResponsesToChatCompletionsRequestWithDeepSeekReasoning(req)
	require.NoError(t, err)

	require.Len(t, defaultOut.Messages, 2, "the default bridge retains its pre-existing reasoning item conversion")
	require.Equal(t, "user", defaultOut.Messages[0].Role)
	require.Empty(t, defaultOut.Messages[0].ReasoningContent)
	require.Len(t, optInOut.Messages, 1)
	require.Equal(t, "explicit", optInOut.Messages[0].ReasoningContent)
}

func TestV027DeepSeekReasoning_OptInNeverForwardsEncryptedContent(t *testing.T) {
	req := &ResponsesRequest{Model: "deepseek-reasoner", Input: json.RawMessage(`[
		{"type":"reasoning","encrypted_content":"ciphertext"},
		{"type":"message","role":"assistant","content":"done"}
	]`)}

	out, err := ResponsesToChatCompletionsRequestWithDeepSeekReasoning(req)

	require.NoError(t, err)
	require.Len(t, out.Messages, 1)
	require.Empty(t, out.Messages[0].ReasoningContent)
}

func TestV027DeepSeekReasoning_OptInResetsDiscardedAndUserBoundaries(t *testing.T) {
	req := &ResponsesRequest{Model: "deepseek-reasoner", Input: json.RawMessage(`[
		{"type":"reasoning","summary":[{"type":"summary_text","text":"one"}]},
		{"type":"reasoning","summary":[{"type":"summary_text","text":" two"}]},
		{"type":"function_call","call_id":"missing","name":"missing","arguments":"{}"},
		{"type":"reasoning","summary":[{"type":"summary_text","text":"fresh"}]},
		{"type":"message","role":"assistant","content":"after missing reply"},
		{"type":"reasoning","summary":[{"type":"summary_text","text":"drop at user"}]},
		{"type":"message","role":"user","content":"new turn"},
		{"type":"message","role":"assistant","content":"after user"}
	]`)}

	out, err := ResponsesToChatCompletionsRequestWithDeepSeekReasoning(req)

	require.NoError(t, err)
	require.Len(t, out.Messages, 4)
	require.Equal(t, "assistant", out.Messages[0].Role)
	require.Equal(t, "one two", out.Messages[0].ReasoningContent, "reasoning-only assistant survives a missing tool reply")
	require.Empty(t, out.Messages[0].ToolCalls)
	require.Equal(t, "assistant", out.Messages[1].Role)
	require.Equal(t, "fresh", out.Messages[1].ReasoningContent, "missing-reply reasoning must not leak into the following turn")
	require.Equal(t, "user", out.Messages[2].Role)
	require.Equal(t, "assistant", out.Messages[3].Role)
	require.Empty(t, out.Messages[3].ReasoningContent, "a user boundary clears pending reasoning")

	consecutive := &ResponsesRequest{Model: "deepseek-reasoner", Input: json.RawMessage(`[
		{"type":"reasoning","summary":[{"type":"summary_text","text":"one"}]},
		{"type":"reasoning","summary":[{"type":"reasoning_text","text":" two"}]},
		{"type":"message","role":"assistant","content":"done"}
	]`)}
	consecutiveOut, err := ResponsesToChatCompletionsRequestWithDeepSeekReasoning(consecutive)
	require.NoError(t, err)
	require.Equal(t, "one two", consecutiveOut.Messages[0].ReasoningContent)
}

func TestV027DeepSeekReasoning_OptInStringInputResetsUserBoundary(t *testing.T) {
	req := &ResponsesRequest{Model: "deepseek-reasoner", Input: json.RawMessage(`[
		{"type":"reasoning","summary":[{"type":"summary_text","text":"must not leak"}]},
		"new user turn",
		{"type":"message","role":"assistant","content":"after string user"}
	]`)}

	out, err := ResponsesToChatCompletionsRequestWithDeepSeekReasoning(req)

	require.NoError(t, err)
	require.Len(t, out.Messages, 2)
	require.Equal(t, "user", out.Messages[0].Role)
	require.Equal(t, "assistant", out.Messages[1].Role)
	require.Empty(t, out.Messages[1].ReasoningContent, "a string input item is a user boundary")
}
