package service

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

const passbackThinkingBody = `{"model":"deepseek-v4-pro","thinking":{"type":"enabled","budget_tokens":1024},"messages":[{"role":"assistant","content":[{"type":"thinking","thinking":"keep me"},{"type":"text","text":"answer"}]}]}`

func TestResolveThinkingProtocol(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  ThinkingProtocol
	}{
		{"claude", "claude-sonnet-4-5", ThinkingProtocolAnthropicStrict},
		{"opus", "opus-4-5", ThinkingProtocolAnthropicStrict},
		{"deepseek", "DeepSeek-V4-Pro", ThinkingProtocolPassbackRequired},
		{"kimi", "kimi-coding", ThinkingProtocolPassbackRequired},
		{"moonshot", "moonshot-v1-32k", ThinkingProtocolPassbackRequired},
		{"glm", "glm-5.1", ThinkingProtocolPassbackRequired},
		{"minimax", "MiniMax-M2.7-highspeed", ThinkingProtocolPassbackRequired},
		{"qwen thinking", "qwen3-235b-a22b-thinking-2507", ThinkingProtocolPassbackRequired},
		{"qwen non thinking", "qwen3-32b", ThinkingProtocolUnknown},
		{"gpt", "gpt-5.1", ThinkingProtocolUnknown},
		{"empty", "", ThinkingProtocolUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, ResolveThinkingProtocol(tt.model))
		})
	}
}

func TestThinkingFiltersSkipPassbackRequired(t *testing.T) {
	in := []byte(passbackThinkingBody)
	require.True(t, bytes.Equal(in, FilterThinkingBlocks(in, "deepseek-v4-pro")))
	require.True(t, bytes.Equal(in, FilterThinkingBlocksForRetry(in, "kimi-coding")))
	require.True(t, bytes.Equal(in, FilterSignatureSensitiveBlocksForRetry(in, "glm-5.1")))
}

func TestThinkingFiltersStillApplyForAnthropicStrict(t *testing.T) {
	in := []byte(passbackThinkingBody)
	out := FilterThinkingBlocks(in, "claude-sonnet-4-5")
	require.False(t, bytes.Equal(in, out))
	require.NotContains(t, string(out), `"type":"thinking"`)
}

func TestNormalizeChineseLLMThinking(t *testing.T) {
	out, applied := NormalizeChineseLLMThinking([]byte(`{"thinking":{"type":"enabled","budget_tokens":8192},"messages":[]}`), "MiniMax-M2.7")
	require.True(t, applied)
	require.Contains(t, string(out), `"type":"adaptive"`)

	unchanged := []byte(`{"thinking":{"type":"enabled"},"messages":[]}`)
	out, applied = NormalizeChineseLLMThinking(unchanged, "kimi-coding")
	require.False(t, applied)
	require.Equal(t, unchanged, out)
}

func TestApplyThinkingEnabledFallback(t *testing.T) {
	got := ApplyThinkingEnabledFallback(nil, []byte(`{"thinking":{"type":"enabled"}}`), "kimi-coding")
	require.NotNil(t, got)
	require.Equal(t, "high", *got)

	got = ApplyThinkingEnabledFallback(nil, []byte(`{"thinking":{"type":"enabled"}}`), "deepseek-v4-pro")
	require.Nil(t, got)

	existing := "medium"
	got = ApplyThinkingEnabledFallback(&existing, []byte(`{"thinking":{"type":"enabled"}}`), "glm-5.1")
	require.NotNil(t, got)
	require.Equal(t, "medium", *got)
}
