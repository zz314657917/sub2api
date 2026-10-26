package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type kimi31SettlementDebits struct {
	newModelBilling20261022Debits
	command *UsageBillingCommand
}

func (r *kimi31SettlementDebits) Apply(ctx context.Context, cmd *UsageBillingCommand) (*UsageBillingApplyResult, error) {
	copy := *cmd
	r.command = &copy
	return r.newModelBilling20261022Debits.Apply(ctx, cmd)
}

func TestKimi31RecordUsageSettlementAndOwnership(t *testing.T) {
	for _, openAI := range []bool{false, true} {
		for _, withResolver := range []bool{false, true} {
			for _, admin := range []string{"default", "channel", "channel-free", "group", "group-free"} {
				if !withResolver && admin != "default" {
					continue
				}
				for _, mapped := range []bool{false, true} {
					t.Run(fmt.Sprintf("openai=%v/resolver=%v/%s/mapped=%v", openAI, withResolver, admin, mapped), func(t *testing.T) {
						cfg := &config.Config{}
						cfg.Default.RateMultiplier = 1
						bs := NewBillingService(cfg, nil)
						logs, users, debits := &newModelBilling20261022Logs{}, &newModelBilling20261022Users{}, &kimi31SettlementDebits{}
						group := &Group{ID: 77, RateMultiplier: .25, LongContextPricingEnabled: true,
							PeakRateEnabled: true, PeakStart: "00:00", PeakEnd: "23:59", PeakRateMultiplier: 2,
							SubscriptionType: SubscriptionTypeStandard}
						price := 7e-6
						if admin == "channel-free" || admin == "group-free" {
							price = 0
						}
						cp := &ChannelModelPricing{Models: []string{"kimi-k3.1-pro"}, BillingMode: BillingModeToken,
							InputPrice: &price, OutputPrice: &price, CacheReadPrice: &price, CacheWritePrice: &price}
						var channelPrice *ChannelModelPricing
						if admin == "channel" || admin == "channel-free" {
							channelPrice = cp
						}
						if admin == "group" || admin == "group-free" {
							group.ModelPricing = []ChannelModelPricing{*cp}
						}
						cs := newModelBilling20261022Channel(group.ID, "kimi-k3.1-pro", channelPrice)
						var resolver *ModelPricingResolver
						if withResolver {
							resolver = NewModelPricingResolver(cs, bs)
						}
						apiKey, user := &APIKey{ID: 55, GroupID: &group.ID, Group: group}, &User{ID: 66}
						account := &Account{ID: 88, Type: AccountTypeAPIKey, Platform: PlatformOpenAI}
						upstream := "kimi-k3.1-pro"
						if mapped {
							upstream = "kimi-k3"
						}
						at := newModelBilling20261022Time("2026-10-26T12:00:00Z")
						var run func() error
						if openAI {
							svc := &OpenAIGatewayService{cfg: cfg, usageLogRepo: logs, usageBillingRepo: debits, userRepo: users,
								billingService: bs, billingCacheService: &BillingCacheService{}, deferredService: &DeferredService{},
								channelService: cs, resolver: resolver}
							run = func() error {
								return svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
									Result: &OpenAIForwardResult{RequestID: "kimi31-openai", Model: "customer-alias", BillingModel: "kimi-k3.1-pro",
										UpstreamModel: upstream, Duration: time.Second,
										Usage: OpenAIUsage{InputTokens: 3000, OutputTokens: 1000, CacheReadInputTokens: 1000, CacheCreationInputTokens: 1000}},
									APIKey: apiKey, User: user, Account: account, RequestStartedAt: at})
							}
						} else {
							svc := &GatewayService{cfg: cfg, usageLogRepo: logs, usageBillingRepo: debits, userRepo: users,
								billingService: bs, billingCacheService: &BillingCacheService{}, deferredService: &DeferredService{},
								channelService: cs, resolver: resolver}
							run = func() error {
								return svc.RecordUsage(context.Background(), &RecordUsageInput{
									Result: &ForwardResult{RequestID: "kimi31-gateway", Model: "kimi-k3.1-pro", UpstreamModel: upstream, Duration: time.Second,
										Usage: ClaudeUsage{InputTokens: 1000, OutputTokens: 1000, CacheReadInputTokens: 1000,
											CacheCreationInputTokens: 1000, CacheCreation5mTokens: 600, CacheCreation1hTokens: 400}},
									ChannelUsageFields: ChannelUsageFields{OriginalModel: "kimi-k3.1-pro", BillingModelSource: BillingModelSourceRequested},
									APIKey:             apiKey, User: user, Account: account, RequestStartedAt: at})
							}
						}
						require.NoError(t, run())
						require.NotNil(t, logs.log)
						input, output, read, write := .013, .068, .0013, .0182
						if openAI {
							write = .013
						} // OpenAIUsage supports aggregate cache creation only, no 1h split.
						if admin != "default" {
							input, output, read, write = 1000*price, 1000*price, 1000*price, 1000*price
						}
						total := input + output + read + write
						require.InDelta(t, input, logs.log.InputCost, 1e-12)
						require.InDelta(t, output, logs.log.OutputCost, 1e-12)
						require.InDelta(t, read, logs.log.CacheReadCost, 1e-12)
						require.InDelta(t, write, logs.log.CacheCreationCost, 1e-12)
						require.InDelta(t, total, logs.log.TotalCost, 1e-12)
						require.InDelta(t, total*.5, logs.log.ActualCost, 1e-12, "only explicit group .25x and 2x peak apply")
						require.False(t, logs.log.LongContextBillingApplied)
						require.NotNil(t, logs.log.AccountStatsCost)
						// Existing account-stats plumbing passes aggregate writes
						// to resolveAccountStatsCost for both recorders.
						stats := .0953
						if mapped {
							stats = .142
						}
						require.InDelta(t, stats, *logs.log.AccountStatsCost, 1e-12, "final account model owns default cost; admin prices do not leak")
						require.NotNil(t, debits.command)
						require.Equal(t, 1, debits.calls)
						require.Equal(t, 1, debits.applied)
						require.InDelta(t, QuantizeUsageBillingAmount(total*.5), debits.command.BalanceCost, 1e-12)
						require.InDelta(t, QuantizeUsageBillingAmount(total*.5), debits.amount, 1e-12)
						require.Equal(t, int64(88), debits.command.AccountID)
						require.Zero(t, users.calls)
						require.NoError(t, run())
						require.Equal(t, 2, debits.calls)
						require.Equal(t, 1, debits.applied)
						require.InDelta(t, QuantizeUsageBillingAmount(total*.5), debits.amount, 1e-12)
					})
				}
			}
		}
	}
}
