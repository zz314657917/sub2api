package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type newModelBilling20261022Logs struct {
	UsageLogRepository
	log *UsageLog
}

func (r *newModelBilling20261022Logs) Create(_ context.Context, log *UsageLog) (bool, error) {
	r.log = log
	return true, nil
}

type newModelBilling20261022Users struct {
	UserRepository
	calls  int
	amount float64
}

func (r *newModelBilling20261022Users) DeductBalance(_ context.Context, _ int64, cost float64) error {
	r.calls++
	r.amount += cost
	return nil
}

type newModelBilling20261022Debits struct {
	UsageBillingRepository
	calls, applied int
	amount         float64
	seen           map[string]bool
}

func (r *newModelBilling20261022Debits) Apply(_ context.Context, cmd *UsageBillingCommand) (*UsageBillingApplyResult, error) {
	r.calls++
	if r.seen == nil {
		r.seen = map[string]bool{}
	}
	if r.seen[cmd.RequestID] {
		return &UsageBillingApplyResult{Applied: false}, nil
	}
	r.seen[cmd.RequestID] = true
	r.applied++
	r.amount += cmd.BalanceCost + cmd.SubscriptionCost
	return &UsageBillingApplyResult{Applied: true}, nil
}

func newModelBilling20261022Channel(groupID int64, model string, price *ChannelModelPricing) *ChannelService {
	cache := newEmptyChannelCache()
	cache.channelByGroupID[groupID] = &Channel{ID: 90, Status: StatusActive}
	cache.groupPlatform[groupID] = ""
	cache.loadedAt = time.Now()
	if price != nil {
		cache.pricingByGroupModel[channelModelKey{groupID: groupID, model: model}] = price
	}
	cs := &ChannelService{}
	cs.cache.Store(cache)
	return cs
}

func TestNewModelBilling20261022RecordersDebitAndStats(t *testing.T) {
	for _, openAI := range []bool{false, true} {
		for _, zeroStarted := range []bool{false, true} {
			for _, withResolver := range []bool{false, true} {
				for _, admin := range []string{"default", "channel", "channel-free", "group", "group-free"} {
					if !withResolver && admin != "default" {
						continue
					}
					t.Run(fmt.Sprintf("openai=%v/zero=%v/resolver=%v/%s", openAI, zeroStarted, withResolver, admin), func(t *testing.T) {
						cfg := &config.Config{}
						cfg.Default.RateMultiplier = 1
						bs := NewBillingService(cfg, nil)
						at := newModelBilling20261022Time("2026-10-22T00:59:59.999999999Z")
						calls := 0
						bs.pricingNow = func() time.Time {
							calls++
							return at.Add(time.Duration(calls-1) * time.Second)
						}
						started := at
						if zeroStarted {
							started = time.Time{}
						}
						logs, users, debits := &newModelBilling20261022Logs{}, &newModelBilling20261022Users{}, &newModelBilling20261022Debits{}
						group := &Group{ID: 77, RateMultiplier: 2, LongContextPricingEnabled: true,
							PeakRateEnabled: true, PeakStart: "00:00", PeakEnd: "23:59", PeakRateMultiplier: 3,
							SubscriptionType: SubscriptionTypeStandard}
						// A free admin card must not be replaced by official rates.
						price := 8e-6
						if admin == "channel-free" || admin == "group-free" {
							price = 0
						}
						cp := &ChannelModelPricing{Models: []string{"deepseek-flash"}, BillingMode: BillingModeToken,
							InputPrice: &price, OutputPrice: &price, CacheReadPrice: &price}
						var channelPrice *ChannelModelPricing
						if admin == "channel" || admin == "channel-free" {
							channelPrice = cp
						}
						if admin == "group" || admin == "group-free" {
							group.ModelPricing = []ChannelModelPricing{*cp}
						}
						cs := newModelBilling20261022Channel(group.ID, "deepseek-flash", channelPrice)
						var resolver *ModelPricingResolver
						if withResolver {
							resolver = NewModelPricingResolver(cs, bs)
						}
						apiKey := &APIKey{ID: 55, GroupID: &group.ID, Group: group}
						user := &User{ID: 66}
						account := &Account{ID: 88, Type: AccountTypeAPIKey, Platform: PlatformOpenAI}
						var run func() error
						if openAI {
							svc := &OpenAIGatewayService{cfg: cfg, usageLogRepo: logs, usageBillingRepo: debits, userRepo: users,
								billingService: bs, billingCacheService: &BillingCacheService{}, deferredService: &DeferredService{},
								channelService: cs, resolver: resolver}
							run = func() error {
								return svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
									Result: &OpenAIForwardResult{RequestID: "new-model-openai", Model: "customer-alias", BillingModel: "deepseek-flash",
										UpstreamModel: "deepseek-v4-pro", Usage: OpenAIUsage{InputTokens: 2000, CacheReadInputTokens: 1000, OutputTokens: 1000}, Duration: time.Second},
									APIKey: apiKey, User: user, Account: account, RequestStartedAt: started})
							}
						} else {
							svc := &GatewayService{cfg: cfg, usageLogRepo: logs, usageBillingRepo: debits, userRepo: users,
								billingService: bs, billingCacheService: &BillingCacheService{}, deferredService: &DeferredService{},
								channelService: cs, resolver: resolver}
							run = func() error {
								return svc.RecordUsage(context.Background(), &RecordUsageInput{
									Result: &ForwardResult{RequestID: "new-model-gateway", Model: "deepseek-flash", UpstreamModel: "deepseek-v4-pro",
										Usage: ClaudeUsage{InputTokens: 1000, CacheReadInputTokens: 1000, OutputTokens: 1000}, Duration: time.Second},
									ChannelUsageFields: ChannelUsageFields{OriginalModel: "deepseek-flash", BillingModelSource: BillingModelSourceRequested},
									APIKey:             apiKey, User: user, Account: account, RequestStartedAt: started})
							}
						}
						require.NoError(t, run())
						require.NotNil(t, logs.log)
						want := .000753
						if admin != "default" {
							want = 3000 * price
						}
						require.InDelta(t, want, logs.log.TotalCost, 1e-12)
						require.InDelta(t, want*6, logs.log.ActualCost, 1e-12, "group peak and official peak each apply once")
						require.NotNil(t, logs.log.AccountStatsCost)
						require.InDelta(t, .01815, *logs.log.AccountStatsCost, 1e-12, "catalog stats use the final Pro model's own card, no admin leak")
						if want > 0 {
							require.Equal(t, 1, debits.calls)
							require.Equal(t, 1, debits.applied)
							require.InDelta(t, QuantizeUsageBillingAmount(want*6), debits.amount, 1e-12)
						} else {
							require.Zero(t, debits.amount)
						}
						require.Zero(t, users.calls, "atomic debit must not also deduct via legacy repository")
						if zeroStarted {
							require.Equal(t, 1, calls, "only one pricing clock capture for customer and stats")
						} else {
							require.Zero(t, calls)
						}
						// Repeat the same settlement identity: one actual balance debit.
						if want > 0 && !zeroStarted {
							require.NoError(t, run())
							require.Equal(t, 2, debits.calls)
							require.Equal(t, 1, debits.applied)
							require.InDelta(t, QuantizeUsageBillingAmount(want*6), debits.amount, 1e-12)
						}
					})
				}
			}
		}
	}
}

func TestNewModelBilling20261022AccountStatsPrecedence(t *testing.T) {
	bs := NewBillingService(&config.Config{}, nil)
	ctx := bs.captureBillingPricingTime(context.Background(), newModelBilling20261022Time("2026-09-14T03:59:59Z"))
	tokens := UsageTokens{InputTokens: 1000, CacheReadTokens: 1000, OutputTokens: 1000}
	cs := newModelBilling20261022Channel(77, "", nil)
	channel := cs.cache.Load().(*channelCache).channelByGroupID[77]
	got := resolveAccountStatsCost(ctx, cs, bs, 88, 77, "deepseek-v4-pro", tokens, 1, 99)
	require.NotNil(t, got)
	require.InDelta(t, (4.50+13.50+.15)*.001*2, *got, 1e-15)
	channel.ApplyPricingToAccountStats = true
	got = resolveAccountStatsCost(ctx, cs, bs, 88, 77, "deepseek-v4-pro", tokens, 1, 99)
	require.Equal(t, 99.0, *got)
	price := 7e-6
	channel.AccountStatsPricingRules = []AccountStatsPricingRule{{AccountIDs: []int64{88}, Pricing: []ChannelModelPricing{{
		Models: []string{"deepseek-v4-pro"}, InputPrice: &price, OutputPrice: &price, CacheReadPrice: &price,
	}}}}
	got = resolveAccountStatsCost(ctx, cs, bs, 88, 77, "deepseek-v4-pro", tokens, 1, 99)
	require.InDelta(t, .021, *got, 1e-15)
	require.Nil(t, resolveAccountStatsCost(ctx, nil, bs, 88, 77, "deepseek-v4-pro", tokens, 1, 99))
}

func TestNewModelBilling20261022GPT61RecorderProtection(t *testing.T) {
	for _, resolver := range []bool{false, true} {
		for _, credentialType := range []string{AccountTypeOAuth, AccountTypeSetupToken, AccountTypeAPIKey} {
			t.Run(fmt.Sprintf("%s/resolver=%v", credentialType, resolver), func(t *testing.T) {
				f := newGPT61SubscriptionFixture(resolver)
				f.input.Account.Type = credentialType
				f.input.Result.UpstreamModel = "gpt-6.1-sol"
				calls := 0
				f.svc.billingService.pricingNow = func() time.Time {
					calls++
					return newModelBilling20261022Time("2026-10-22T00:59:59Z").Add(time.Duration(calls-1) * time.Minute)
				}
				require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
				total := .61
				if credentialType == AccountTypeAPIKey {
					total = 1.215
				}
				f.assertSettlement(t, total, 1.3, total, credentialType == AccountTypeAPIKey)
				require.Equal(t, 1, calls)
			})
		}
	}
}

type newModelBilling20261022OpusDebits struct {
	newModelBilling20261022Debits
	charged int
}

func (r *newModelBilling20261022OpusDebits) Apply(ctx context.Context, cmd *UsageBillingCommand) (*UsageBillingApplyResult, error) {
	result, err := r.newModelBilling20261022Debits.Apply(ctx, cmd)
	if err == nil && result.Applied && cmd.BalanceCost+cmd.SubscriptionCost > 0 {
		r.charged++
	}
	return result, err
}

func TestNewModelBilling20261022OpusPriorityZeroRecordUsage(t *testing.T) {
	for _, withResolver := range []bool{false, true} {
		for _, atomic := range []bool{false, true} {
			for _, tc := range []struct {
				card   string
				writes int
				want   float64
			}{
				{"zero", 0, 0},
				{"zero", 8000, 0},
				{"nonzero", 8000, 1.048},
				{"partial-input", 8000, .968},
				{"partial-write", 8000, 1.048},
			} {
				t.Run(fmt.Sprintf("resolver=%v/atomic=%v/%s/write=%d", withResolver, atomic, tc.card, tc.writes), func(t *testing.T) {
					cfg := &config.Config{}
					cfg.Default.RateMultiplier = 1
					bs := newModelBilling20261022OpusPriorityCard(t, tc.card)
					logs, users := &newModelBilling20261022Logs{}, &newModelBilling20261022Users{}
					debits := &newModelBilling20261022OpusDebits{}
					group := &Group{ID: 77, RateMultiplier: 1, LongContextPricingEnabled: true}
					cs := newModelBilling20261022Channel(group.ID, "", nil)
					svc := &OpenAIGatewayService{cfg: cfg, usageLogRepo: logs, userRepo: users, billingService: bs,
						billingCacheService: &BillingCacheService{}, deferredService: &DeferredService{}, channelService: cs}
					if withResolver {
						svc.resolver = NewModelPricingResolver(cs, bs)
					}
					if atomic {
						svc.usageBillingRepo = debits
					}
					tier := "priority"
					input := &OpenAIRecordUsageInput{
						Result: &OpenAIForwardResult{RequestID: "opus-priority-zero-r3", Model: "claude-opus-5-5-high",
							UpstreamModel: "claude-opus-5-5-high", ServiceTier: &tier,
							Usage: OpenAIUsage{InputTokens: 90000 + tc.writes, OutputTokens: 10000,
								CacheReadInputTokens: 20000, CacheCreationInputTokens: tc.writes}},
						APIKey: &APIKey{ID: 55, GroupID: &group.ID, Group: group},
						User:   &User{ID: 66}, Account: &Account{ID: 88, Type: AccountTypeAPIKey, Platform: PlatformOpenAI},
						RequestStartedAt: newModelBilling20261022Time("2026-10-22T00:59:59Z"),
					}
					require.NoError(t, svc.RecordUsage(context.Background(), input))
					require.NotNil(t, logs.log)
					require.InDelta(t, tc.want, logs.log.TotalCost, 1e-12)
					require.InDelta(t, tc.want, logs.log.ActualCost, 1e-12)
					require.NotNil(t, logs.log.AccountStatsCost)
					require.InDelta(t, .484+float64(tc.writes)*5e-6, *logs.log.AccountStatsCost, 1e-12,
						"catalog account stats retain standard prices independently of customer priority zero")
					wantCharged := 0
					if tc.want > 0 {
						wantCharged = 1
					}
					if atomic {
						require.Equal(t, 1, debits.calls, "atomic settlement records even a free request")
						require.Equal(t, 1, debits.applied)
						require.Equal(t, wantCharged, debits.charged)
						require.InDelta(t, tc.want, debits.amount, 1e-12)
						require.Zero(t, users.calls)
						require.NoError(t, svc.RecordUsage(context.Background(), input))
						require.Equal(t, 2, debits.calls)
						require.Equal(t, 1, debits.applied)
						require.Equal(t, wantCharged, debits.charged)
						require.InDelta(t, tc.want, debits.amount, 1e-12)
					} else {
						require.Equal(t, wantCharged, users.calls)
						require.InDelta(t, tc.want, users.amount, 1e-12)
						require.Zero(t, debits.calls)
					}
				})
			}
		}
	}
}
