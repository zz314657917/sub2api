package service

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/stretchr/testify/require"
)

func TestCopyOpenAIUsageFromResponsesUsagePreservesCanonicalCacheWrite(t *testing.T) {
	var usage apicompat.ResponsesUsage
	require.NoError(t, json.Unmarshal([]byte(`{
		"input_tokens":20,
		"output_tokens":2,
		"input_tokens_details":{"cached_tokens":3,"cache_write_tokens":7}
	}`), &usage))

	got := copyOpenAIUsageFromResponsesUsage(&usage)

	require.Equal(t, 20, got.InputTokens)
	require.Equal(t, 2, got.OutputTokens)
	require.Equal(t, 3, got.CacheReadInputTokens)
	require.Equal(t, 7, got.CacheCreationInputTokens)
}

func TestCopyOpenAIUsageFromResponsesUsageTrustsCanonicalCacheCreationZero(t *testing.T) {
	usage := &apicompat.ResponsesUsage{
		InputTokens:              20,
		OutputTokens:             2,
		CacheCreationInputTokens: 0,
		InputTokensDetails: &apicompat.ResponsesInputTokensDetails{
			CachedTokens:     3,
			CacheWriteTokens: 19,
		},
	}

	got := copyOpenAIUsageFromResponsesUsage(usage)

	require.Equal(t, 3, got.CacheReadInputTokens)
	require.Zero(t, got.CacheCreationInputTokens)
}

func TestExtractCCStreamUsageCapturesCacheWriteTokens(t *testing.T) {
	got := extractCCStreamUsage(`{
		"usage":{
			"prompt_tokens":12,
			"completion_tokens":3,
			"prompt_tokens_details":{"cached_tokens":4,"cache_write_tokens":6}
		}
	}`)

	require.NotNil(t, got)
	require.Equal(t, 12, got.InputTokens)
	require.Equal(t, 3, got.OutputTokens)
	require.Equal(t, 4, got.CacheReadInputTokens)
	require.Equal(t, 6, got.CacheCreationInputTokens)
}
