//go:build unit

package antigravity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestV0215ProtocolThinkingSignature(t *testing.T) {
	for _, tt := range []struct {
		name string
		item ClaudeContentItem
		want string
	}{
		{"thinking_plan", ClaudeContentItem{Type: "thinking", Thinking: "plan"}, `{"type":"thinking","thinking":"plan","signature":""}`},
		{"thinking_signature", ClaudeContentItem{Type: "thinking", Signature: "sig"}, `{"type":"thinking","thinking":"","signature":"sig"}`},
		{"thinking_both_empty", ClaudeContentItem{Type: "thinking"}, `{"type":"thinking","thinking":"","signature":""}`},
		{"text_nonempty", ClaudeContentItem{Type: "text", Text: "hi"}, `{"type":"text","text":"hi"}`},
		{"text_empty_baseline", ClaudeContentItem{Type: "text"}, `{"type":"text"}`},
		{"tool_use_no_thinking_fields", ClaudeContentItem{Type: "tool_use", ID: "call", Name: "lookup", Input: map[string]any{}}, `{"type":"tool_use","id":"call","name":"lookup","input":{}}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			wire, err := json.Marshal(tt.item)
			require.NoError(t, err)
			require.JSONEq(t, tt.want, string(wire))
		})
	}
}

func TestV0215ProtocolStreamThinkingSignature(t *testing.T) {
	for _, emit := range []func(*StreamingProcessor) []byte{
		func(p *StreamingProcessor) []byte { return p.processThinking("plan", "") },
		func(p *StreamingProcessor) []byte { return p.emitEmptyThinkingWithSignature("sig") },
	} {
		wire := string(emit(NewStreamingProcessor("claude")))
		require.Contains(t, wire, `"content_block":{"signature":"","thinking":"","type":"thinking"}`)
		require.Contains(t, wire, "event: content_block_start")
	}
}
