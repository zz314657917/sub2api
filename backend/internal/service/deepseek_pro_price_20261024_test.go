package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestDeepseekProSelectedPrice20261024(t *testing.T) {
	// User-selected CNY card: read .15, uncached input 4.50, output 13.50.
	// Site credits map 1:1 to CNY; charge these exact amounts without FX.
	for _, source := range []string{"fallback", "stale-dynamic"} {
		t.Run(source, func(t *testing.T) {
			bs := NewBillingService(&config.Config{}, nil)
			if source == "stale-dynamic" {
				bs.pricingService = &PricingService{pricingData: map[string]*LiteLLMModelPricing{
					"deepseek-v4-pro":      {InputCostPerToken: .15e-6, OutputCostPerToken: .60e-6, CacheReadInputTokenCost: .003e-6},
					"deepseek-v4-pro-0813": {InputCostPerToken: 99, OutputCostPerToken: 99},
				}}
			}
			for _, model := range []string{"deepseek-v4-pro", "deepseek-v4-pro-0813", "deepseek-v4-pro-20260813", "DEEPSEEK-V4-PRO"} {
				for _, pricing := range []func(string) *ModelPricing{
					bs.getFallbackPricing,
					func(model string) *ModelPricing {
						p, err := bs.GetModelPricing(model)
						require.NoError(t, err)
						return p
					},
				} {
					card := pricing(model)
					require.NotNil(t, card)
					require.InDelta(t, 4.50e-6, card.InputPricePerToken, 1e-18)
					require.InDelta(t, 13.50e-6, card.OutputPricePerToken, 1e-18)
					require.InDelta(t, .15e-6, card.CacheReadPricePerToken, 1e-18)
				}
				for _, tc := range []struct {
					at     string
					factor float64
				}{
					{"2026-09-14T04:00:00Z", 1}, // Old transition must not switch Pro to Flash.
					{"2026-10-24T01:00:00Z", 1}, // Saturday 09:00 Beijing.
					{"2026-10-26T01:00:00Z", 2}, // Monday 09:00 Beijing.
				} {
					at, err := time.Parse(time.RFC3339, tc.at)
					require.NoError(t, err)
					for _, resolver := range []*ModelPricingResolver{nil, NewModelPricingResolver(nil, bs)} {
						bd, err := bs.CalculateTokenCostForRequest(TokenCostRequest{
							Ctx: context.Background(), PricingAt: at, Model: model, Resolver: resolver,
							Tokens:         UsageTokens{InputTokens: 1_000_000, OutputTokens: 1_000_000, CacheReadTokens: 1_000_000},
							RateMultiplier: 3,
						})
						require.NoError(t, err)
						require.InDelta(t, 4.50*tc.factor, bd.InputCost, 1e-12)
						require.InDelta(t, 13.50*tc.factor, bd.OutputCost, 1e-12)
						require.InDelta(t, .15*tc.factor, bd.CacheReadCost, 1e-12)
						require.InDelta(t, 18.15*tc.factor*3, bd.ActualCost, 1e-12)
					}
				}
			}
		})
	}
}

func TestDeepseekProOneToOneSettlement20261024(t *testing.T) {
	for _, openAI := range []bool{false, true} {
		for _, rate := range []float64{1, .25} {
			for _, tc := range []struct {
				at     string
				factor float64
			}{
				{"2026-10-24T01:00:00Z", 1},
				{"2026-10-26T01:00:00Z", 2},
			} {
				t.Run(fmt.Sprintf("openai=%v/rate=%g/%s", openAI, rate, tc.at), func(t *testing.T) {
					cfg := &config.Config{}
					cfg.Default.RateMultiplier = 1
					bs := NewBillingService(cfg, nil)
					logs, users, debits := &newModelBilling20261022Logs{}, &newModelBilling20261022Users{}, &newModelBilling20261022Debits{}
					group := &Group{ID: 77, RateMultiplier: rate, SubscriptionType: SubscriptionTypeStandard}
					cs := newModelBilling20261022Channel(group.ID, "", nil)
					resolver := NewModelPricingResolver(cs, bs)
					apiKey := &APIKey{ID: 55, GroupID: &group.ID, Group: group}
					user := &User{ID: 66}
					account := &Account{ID: 88, Type: AccountTypeAPIKey, Platform: PlatformOpenAI}
					at, err := time.Parse(time.RFC3339, tc.at)
					require.NoError(t, err)
					const model = "deepseek-v4-pro-0813"
					if openAI {
						svc := &OpenAIGatewayService{cfg: cfg, usageLogRepo: logs, usageBillingRepo: debits, userRepo: users,
							billingService: bs, billingCacheService: &BillingCacheService{}, deferredService: &DeferredService{},
							channelService: cs, resolver: resolver}
						require.NoError(t, svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
							Result: &OpenAIForwardResult{RequestID: "pro-one-to-one-openai", Model: model, BillingModel: model, UpstreamModel: model,
								Usage: OpenAIUsage{InputTokens: 2_000_000, CacheReadInputTokens: 1_000_000, OutputTokens: 1_000_000}, Duration: time.Second},
							APIKey: apiKey, User: user, Account: account, RequestStartedAt: at,
						}))
					} else {
						svc := &GatewayService{cfg: cfg, usageLogRepo: logs, usageBillingRepo: debits, userRepo: users,
							billingService: bs, billingCacheService: &BillingCacheService{}, deferredService: &DeferredService{},
							channelService: cs, resolver: resolver}
						require.NoError(t, svc.RecordUsage(context.Background(), &RecordUsageInput{
							Result: &ForwardResult{RequestID: "pro-one-to-one-gateway", Model: model, UpstreamModel: model,
								Usage: ClaudeUsage{InputTokens: 1_000_000, CacheReadInputTokens: 1_000_000, OutputTokens: 1_000_000}, Duration: time.Second},
							ChannelUsageFields: ChannelUsageFields{OriginalModel: model, BillingModelSource: BillingModelSourceRequested},
							APIKey:             apiKey, User: user, Account: account, RequestStartedAt: at,
						}))
					}
					require.NotNil(t, logs.log)
					want := 18.15 * tc.factor
					require.InDelta(t, 4.50*tc.factor, logs.log.InputCost, 1e-12)
					require.InDelta(t, 13.50*tc.factor, logs.log.OutputCost, 1e-12)
					require.InDelta(t, .15*tc.factor, logs.log.CacheReadCost, 1e-12)
					require.InDelta(t, want, logs.log.TotalCost, 1e-12)
					require.InDelta(t, want*rate, logs.log.ActualCost, 1e-12)
					require.NotNil(t, logs.log.AccountStatsCost)
					require.InDelta(t, want, *logs.log.AccountStatsCost, 1e-12)
					require.Equal(t, 1, debits.applied)
					require.InDelta(t, QuantizeUsageBillingAmount(want*rate), debits.amount, 1e-12)
					require.Zero(t, users.calls, "atomic settlement must not also deduct through the legacy repository")
				})
			}
		}
	}
}
