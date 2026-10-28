package antigravity

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpstreamAFixesAntigravityMalformedTransform(t *testing.T) {
	for _, body := range []string{
		`{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL","content":{"parts":[{"thoughtSignature":"test-signature"}]}}]}`,
		`{"response":{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL"}],"usageMetadata":{"promptTokenCount":9}}}`,
		`{"candidates":[{"finishReason":"MALFORMED_FUNCTION_CALL","content":{"parts":[{"functionCall":{"name":"read","args":{}}}]}}]}`,
	} {
		response, usage, err := TransformGeminiToClaude([]byte(body), "test-model")
		require.ErrorContains(t, err, "MALFORMED_FUNCTION_CALL")
		require.ErrorIs(t, err, ErrMalformedFunctionCall)
		require.Nil(t, response)
		require.Nil(t, usage)
	}
}

func TestUpstreamAFixesAntigravityNormalTransform(t *testing.T) {
	for _, body := range []string{
		`{"candidates":[{"finishReason":"STOP","content":{"parts":[{"text":"hello"}]}}]}`,
		`{"response":{"candidates":[{"finishReason":"STOP","content":{"parts":[{"functionCall":{"name":"read","args":{}}}]}}]}}`,
	} {
		response, usage, err := TransformGeminiToClaude([]byte(body), "test-model")
		require.NoError(t, err)
		require.NotNil(t, response)
		require.NotNil(t, usage)
	}
}
