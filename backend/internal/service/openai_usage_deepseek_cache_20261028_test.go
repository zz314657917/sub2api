package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUpstreamAFixesDeepSeekCacheJSONAndStreaming(t *testing.T) {
	for _, tc := range []struct {
		name   string
		fields string
		want   int
	}{
		{"deepseek", `"prompt_cache_hit_tokens":160,"prompt_cache_miss_tokens":40`, 160},
		{"nested_input", `"input_tokens_details":{"cached_tokens":7},"prompt_cache_hit_tokens":160`, 7},
		{"nested_prompt", `"prompt_tokens_details":{"cached_tokens":8},"prompt_cache_hit_tokens":160`, 8},
		{"nested_input_zero", `"input_tokens_details":{"cached_tokens":0},"prompt_cache_hit_tokens":160`, 0},
		{"nested_prompt_zero", `"prompt_tokens_details":{"cached_tokens":0},"prompt_cache_hit_tokens":160`, 0},
		{"nested_invalid", `"prompt_tokens_details":{"cached_tokens":"bad"},"prompt_cache_hit_tokens":160`, 0},
		{"legacy_input", `"cache_read_input_tokens":9,"prompt_cache_hit_tokens":160`, 9},
		{"legacy_read", `"cache_read_tokens":10,"prompt_cache_hit_tokens":160`, 10},
		{"legacy_cached", `"cached_tokens":11,"prompt_cache_hit_tokens":160`, 11},
		{"standard", `"prompt_tokens_details":{"cached_tokens":12}`, 12},
		{"absent", `"prompt_cache_miss_tokens":200`, 0},
		{"negative", `"prompt_cache_hit_tokens":-1`, 0},
		{"fraction", `"prompt_cache_hit_tokens":1.5`, 0},
		{"overflow", `"prompt_cache_hit_tokens":9223372036854775808`, 0},
		{"huge_exponent", `"prompt_cache_hit_tokens":1e99999999`, 0},
		{"string", `"prompt_cache_hit_tokens":"160"`, 0},
		{"null", `"prompt_cache_hit_tokens":null`, 0},
		{"bool", `"prompt_cache_hit_tokens":true`, 0},
		{"object", `"prompt_cache_hit_tokens":{}`, 0},
		{"exponent", `"prompt_cache_hit_tokens":16e1`, 160},
	} {
		for _, stream := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stream=%t", tc.name, stream), func(t *testing.T) {
				payload := `{"id":"chatcmpl_core","model":"deepseek-v4-flash","choices":[],"usage":{"prompt_tokens":200,"completion_tokens":3,` + tc.fields + `}}`
				svc := &OpenAIGatewayService{cfg: &config.Config{}}
				writer := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(writer)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
				account := &Account{ID: 1, Platform: PlatformDeepseek, Type: AccountTypeAPIKey}
				header := http.Header{"Content-Type": {"application/json"}}
				body := payload
				if stream {
					header.Set("Content-Type", "text/event-stream")
					body = "data: " + payload + "\n\ndata: [DONE]\n\n"
				}
				response := &http.Response{StatusCode: http.StatusOK, Header: header, Body: io.NopCloser(strings.NewReader(body))}
				var usage *OpenAIUsage
				if stream {
					result, err := svc.handleStreamingResponse(context.Background(), response, c, account, time.Now(), "deepseek-v4-flash", "deepseek-v4-flash")
					require.NoError(t, err)
					usage = result.usage
				} else {
					result, err := svc.handleNonStreamingResponse(context.Background(), response, c, account, "deepseek-v4-flash", "deepseek-v4-flash")
					require.NoError(t, err)
					usage = result.usage
				}
				require.Equal(t, 200, usage.InputTokens)
				require.Equal(t, 3, usage.OutputTokens)
				require.Equal(t, tc.want, usage.CacheReadInputTokens)
				require.Equal(t, http.StatusOK, writer.Code)
				require.Contains(t, writer.Body.String(), `"prompt_tokens":200`)
			})
		}
	}
}

func TestUpstreamAFixesDeepSeekCacheTerminalPrecedence(t *testing.T) {
	svc := &OpenAIGatewayService{}
	usage := &OpenAIUsage{}
	svc.parseSSEUsage(`{"usage":{"prompt_tokens":200,"prompt_cache_hit_tokens":160}}`, usage)
	require.Equal(t, 160, usage.CacheReadInputTokens)
	svc.parseSSEUsage(`{"type":"response.completed","response":{"usage":{"input_tokens":200,"input_tokens_details":{"cached_tokens":0},"prompt_cache_hit_tokens":160}}}`, usage)
	require.Zero(t, usage.CacheReadInputTokens)
	require.Equal(t, 200, usage.InputTokens)
	svc.parseSSEUsage("[DONE]", usage)
	require.Zero(t, usage.CacheReadInputTokens)
}

type upstreamAUsageMemoryRepo struct {
	UsageLogRepository
	logs map[string]*UsageLog
}

func (r *upstreamAUsageMemoryRepo) Create(_ context.Context, log *UsageLog) (bool, error) {
	if r.logs == nil {
		r.logs = make(map[string]*UsageLog)
	}
	if r.logs[log.RequestID] != nil {
		return false, nil
	}
	copy := *log
	r.logs[log.RequestID] = &copy
	return true, nil
}

func TestUpstreamAFixesDeepSeekCacheRecorderIdempotency(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprintf("stream=%t", stream), func(t *testing.T) {
			logs := &upstreamAUsageMemoryRepo{}
			bills := &openAIRecordUsageDedupeBillingRepoStub{}
			users := &openAIRecordUsageUserRepoStub{}
			svc := newOpenAIRecordUsageServiceWithBillingRepoForTest(logs, bills, users, &openAIRecordUsageSubRepoStub{}, nil)
			// Freeze the existing off-peak prices, without changing the billing implementation or catalog.
			started := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
			payload := `{"usage":{"prompt_tokens":2000000,"completion_tokens":1000000,"prompt_cache_hit_tokens":1500000,"prompt_cache_miss_tokens":500000}}`
			usage, ok := extractOpenAIUsageFromJSONBytes([]byte(payload))
			require.True(t, ok)
			if stream {
				usage = *svc.parseSSEUsageFromBody("data: " + payload + "\n\ndata: [DONE]\n\n")
			}
			input := &OpenAIRecordUsageInput{
				Result: &OpenAIForwardResult{RequestID: "core_deepseek_cache", Model: "deepseek-v4-flash", Usage: usage, Stream: stream},
				APIKey: &APIKey{ID: 100, Group: &Group{RateMultiplier: 1}},
				User:   &User{ID: 200}, Account: &Account{ID: 300, Platform: PlatformDeepseek, Type: AccountTypeAPIKey},
				RequestStartedAt: started,
			}
			require.NoError(t, svc.RecordUsage(context.Background(), input))
			require.NoError(t, svc.RecordUsage(context.Background(), input))
			require.Len(t, logs.logs, 1)
			log := logs.logs["core_deepseek_cache"]
			require.NotNil(t, log)
			require.Equal(t, 500000, log.InputTokens)
			require.Equal(t, 1500000, log.CacheReadTokens)
			require.Equal(t, 1000000, log.OutputTokens)
			// Existing Flash prices: input .15/output .60/cache read .003 per million.
			const wantCost = .075 + .60 + .0045
			require.InDelta(t, .0045, log.CacheReadCost, 1e-10)
			require.InDelta(t, wantCost, log.TotalCost, 1e-10)
			// No GroupID is selected, so the fixture's existing default multiplier is 1.1.
			require.InDelta(t, wantCost*1.1, log.ActualCost, 1e-10)
			require.Equal(t, 2, bills.calls)
			require.Equal(t, 1, bills.appliedCalls)
			require.InDelta(t, wantCost*1.1, bills.appliedBalanceCost, 1e-10)
			require.Equal(t, 1500000, bills.commands[0].CacheReadTokens)
			require.Equal(t, bills.commands[0].RequestFingerprint, bills.commands[1].RequestFingerprint)
			require.Zero(t, users.deductCalls, "the atomic billing repository owns the single charge")
		})
	}
}
