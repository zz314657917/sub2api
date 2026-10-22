package service

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGeminiV0217UsageClampAtRealEntries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, entry := range geminiSingleCheckEntrypoints {
		for _, stream := range []bool{false, true} {
			for _, tc := range []struct {
				name, metadata string
				input          int
			}{
				{"missing_prompt", `"cachedContentTokenCount":30`, 0},
				{"cache_exceeds_prompt", `"promptTokenCount":10,"cachedContentTokenCount":30`, 0},
				{"ordinary_cache", `"promptTokenCount":100,"cachedContentTokenCount":30`, 70},
			} {
				t.Run(fmt.Sprintf("%s/stream=%t/%s", entry, stream, tc.name), func(t *testing.T) {
					body := fmt.Sprintf(`{"candidates":[{"content":{"role":"model","parts":[{"text":"hi"}]},"finishReason":"STOP"}],"usageMetadata":{%s,"candidatesTokenCount":5,"thoughtsTokenCount":2}}`, tc.metadata)
					if stream {
						body = "data: " + body + "\n\n"
					}
					svc, upstream, repo := newGeminiSingleCheckService(geminiSingleCheckResponse{200, body})
					result, err, rec := geminiSingleCheckForward(t, entry, svc, geminiCacheTestAccount(), stream)
					require.NoError(t, err)
					require.NotNil(t, result)
					require.Equal(t, tc.input, result.Usage.InputTokens)
					require.Equal(t, 30, result.Usage.CacheReadInputTokens)
					require.Equal(t, 7, result.Usage.OutputTokens)
					require.Zero(t, result.Usage.CacheCreationInputTokens)
					require.Len(t, upstream.requests, 1)
					require.NotContains(t, rec.Body.String(), `"input_tokens":-`)
					require.Equal(t, &geminiSingleCheckRepo{}, repo)
					if entry == "native" {
						require.Equal(t, body, rec.Body.String(), "native wire usage must remain untouched")
					}
				})
			}
		}
	}
}
