package antigravity

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeminiV0217UsageClamp(t *testing.T) {
	for _, tc := range []struct {
		name, metadata string
		input, cached  int
	}{
		{"missing_prompt", `"cachedContentTokenCount":30`, 0, 30},
		{"cache_exceeds_prompt", `"promptTokenCount":10,"cachedContentTokenCount":30`, 0, 30},
		{"ordinary_cache", `"promptTokenCount":100,"cachedContentTokenCount":30`, 70, 30},
		{"no_cache", `"promptTokenCount":100`, 100, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"candidates":[{"content":{"parts":[{"text":"hi"}],"role":"model"},"finishReason":"STOP"}],"usageMetadata":{%s,"candidatesTokenCount":5,"thoughtsTokenCount":2}}`, tc.metadata)
			t.Run("nonstream", func(t *testing.T) {
				_, usage, err := TransformGeminiToClaude([]byte(body), "gemini-2.5-flash")
				require.NoError(t, err)
				require.Equal(t, tc.input, usage.InputTokens)
				require.Equal(t, tc.cached, usage.CacheReadInputTokens)
				require.Equal(t, 7, usage.OutputTokens)
				require.Zero(t, usage.CacheCreationInputTokens)
			})
			t.Run("stream", func(t *testing.T) {
				p := NewStreamingProcessor("gemini-2.5-flash")
				out := p.ProcessLine(`data: {"response":` + body + `}`)
				require.Contains(t, string(out), `"message_start"`)
				require.Contains(t, string(out), fmt.Sprintf(`"input_tokens":%d`, tc.input))
				require.NotContains(t, string(out), `"input_tokens":-`)
				_, usage := p.Finish()
				require.Equal(t, tc.input, usage.InputTokens)
				require.Equal(t, tc.cached, usage.CacheReadInputTokens)
				require.Equal(t, 7, usage.OutputTokens)
				require.Zero(t, usage.CacheCreationInputTokens)
			})
		})
	}
}

func TestGeminiV0217UsageClampLaterStreamUpdate(t *testing.T) {
	p := NewStreamingProcessor("gemini-2.5-flash")
	p.ProcessLine(`data: {"response":{"usageMetadata":{"promptTokenCount":100,"cachedContentTokenCount":30}}}`)
	out := p.ProcessLine(`data: {"response":{"usageMetadata":{"cachedContentTokenCount":40},"candidates":[{"finishReason":"STOP"}]}}`)
	require.NotContains(t, string(out), `"input_tokens":-`)
	_, usage := p.Finish()
	require.Zero(t, usage.InputTokens)
	require.Equal(t, 40, usage.CacheReadInputTokens)
}
