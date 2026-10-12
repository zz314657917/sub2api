package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type gpt61SubscriptionUserRepo struct {
	openAIRecordUsageUserRepoStub
	balance float64
}

func (r *gpt61SubscriptionUserRepo) DeductBalance(ctx context.Context, id int64, amount float64) error {
	if err := r.openAIRecordUsageUserRepoStub.DeductBalance(ctx, id, amount); err != nil {
		return err
	}
	r.balance -= amount
	return nil
}

type gpt61SubscriptionSubRepo struct {
	openAIRecordUsageSubRepoStub
	amount float64
}

func (r *gpt61SubscriptionSubRepo) IncrementUsage(ctx context.Context, id int64, amount float64) error {
	if err := r.openAIRecordUsageSubRepoStub.IncrementUsage(ctx, id, amount); err != nil {
		return err
	}
	r.amount += amount
	return nil
}

type gpt61SubscriptionAccountRepo struct {
	AccountRepository
	parent *Account
	err    error
	ids    []int64
}

func (r *gpt61SubscriptionAccountRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	r.ids = append(r.ids, id)
	return r.parent, r.err
}

type gpt61SubscriptionFixture struct {
	svc   *OpenAIGatewayService
	usage *openAIRecordUsageLogRepoStub
	user  *gpt61SubscriptionUserRepo
	sub   *gpt61SubscriptionSubRepo
	input *OpenAIRecordUsageInput
}

func newGPT61SubscriptionFixture(withResolver bool) *gpt61SubscriptionFixture {
	f := &gpt61SubscriptionFixture{
		usage: &openAIRecordUsageLogRepoStub{inserted: true},
		user:  &gpt61SubscriptionUserRepo{balance: 100},
		sub:   &gpt61SubscriptionSubRepo{},
	}
	f.svc = newOpenAIRecordUsageServiceForTest(f.usage, f.user, f.sub, nil)
	f.svc.channelService = gpt61SubscriptionChannel(&Channel{ID: 61})
	if withResolver {
		f.svc.resolver = NewModelPricingResolver(f.svc.channelService, f.svc.billingService)
	}
	f.input = &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{Model: "gpt-6.1-sol", RequestID: "gpt61-subscription",
			Usage: OpenAIUsage{InputTokens: 300000, OutputTokens: 1000}},
		Account: &Account{ID: 31, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
			Extra: map[string]any{"openai_long_context_billing_enabled": true}},
		User:   &User{ID: 21},
		APIKey: &APIKey{ID: 11, GroupID: i64p(61), Group: &Group{ID: 61, LongContextPricingEnabled: true, RateMultiplier: 1.3}},
	}
	return f
}

func (f *gpt61SubscriptionFixture) assertSettlement(t *testing.T, total, multiplier, stats float64, longContext bool) {
	t.Helper()
	log := f.usage.lastLog
	require.NotNil(t, log)
	require.InDelta(t, total, log.TotalCost, 1e-10)
	require.InDelta(t, total*multiplier, log.ActualCost, 1e-10)
	require.InDelta(t, multiplier, log.RateMultiplier, 1e-12)
	require.Equal(t, longContext, log.LongContextBillingApplied)
	require.Equal(t, f.input.Account.ID, log.AccountID)
	require.Equal(t, f.input.Account.BillingRateMultiplier(), *log.AccountRateMultiplier)
	require.NotNil(t, log.AccountStatsCost)
	require.InDelta(t, stats, *log.AccountStatsCost, 1e-10)
	if f.input.Subscription != nil {
		require.Equal(t, BillingTypeSubscription, log.BillingType)
		require.Equal(t, f.input.Subscription.ID, *log.SubscriptionID)
		require.Equal(t, 1, f.sub.incrementCalls)
		require.InDelta(t, total*multiplier, f.sub.amount, 1e-10)
		require.Zero(t, f.user.deductCalls)
		require.Equal(t, 100.0, f.user.balance)
	} else {
		require.Equal(t, BillingTypeBalance, log.BillingType)
		require.Equal(t, 1, f.user.deductCalls)
		require.InDelta(t, total*multiplier, f.user.lastAmount, 1e-10)
		require.InDelta(t, 100-total*multiplier, f.user.balance, 1e-10)
		require.Zero(t, f.sub.incrementCalls)
	}
}

func gpt61SubscriptionChannel(channel *Channel, prices ...*ChannelModelPricing) *ChannelService {
	channel.Status = StatusActive
	cache := newEmptyChannelCache()
	cache.channelByGroupID[61] = channel
	cache.groupPlatform[61] = PlatformOpenAI
	for _, pricing := range prices {
		for _, model := range pricing.Models {
			cache.pricingByGroupModel[channelModelKey{groupID: 61, platform: PlatformOpenAI, model: model}] = pricing
		}
	}
	cache.loadedAt = time.Now()
	cs := &ChannelService{}
	cs.cache.Store(cache)
	return cs
}

func TestOpenAISubscriptionGPT61RecordUsage(t *testing.T) {
	usage := &openAIRecordUsageLogRepoStub{inserted: true}
	user := &openAIRecordUsageUserRepoStub{}
	svc := newOpenAIRecordUsageServiceForTest(usage, user, &openAIRecordUsageSubRepoStub{}, nil)
	svc.channelService = gpt61SubscriptionChannel(&Channel{ID: 61})
	svc.resolver = NewModelPricingResolver(svc.channelService, svc.billingService)
	require.NoError(t, svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{Model: "gpt-6.1-sol", RequestID: "gpt61-subscription",
			Usage: OpenAIUsage{InputTokens: 300000, OutputTokens: 1000}},
		Account: &Account{ID: 31, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
			Extra: map[string]any{"openai_long_context_billing_enabled": true}},
		User:   &User{ID: 21},
		APIKey: &APIKey{ID: 11, GroupID: i64p(61), Group: &Group{ID: 61, LongContextPricingEnabled: true, RateMultiplier: 1.3}},
	}))
	require.NotNil(t, usage.lastLog)
	require.InDelta(t, .6, usage.lastLog.InputCost, 1e-12)
	require.InDelta(t, .01, usage.lastLog.OutputCost, 1e-12)
	require.InDelta(t, .61, usage.lastLog.TotalCost, 1e-12)
	require.False(t, usage.lastLog.LongContextBillingApplied)
	require.Equal(t, 1, user.deductCalls)
	require.InDelta(t, .61*1.3, user.lastAmount, 1e-12)
	require.NotNil(t, usage.lastLog.AccountStatsCost)
	require.InDelta(t, .61, *usage.lastLog.AccountStatsCost, 1e-12)
}

func TestOpenAISubscriptionGPT61SettlementMatrix(t *testing.T) {
	for _, credential := range []string{AccountTypeOAuth, AccountTypeSetupToken, AccountTypeAPIKey} {
		for _, resolver := range []bool{false, true} {
			for _, accountGate := range []bool{false, true} {
				for _, groupGate := range []bool{false, true} {
					for _, n := range []int{271999, 272000, 272001} {
						for _, tier := range []string{"", "priority"} {
							for _, cached := range []bool{false, true} {
								name := fmt.Sprintf("%s/resolver=%t/account=%t/group=%t/n=%d/tier=%s/cache=%t", credential, resolver, accountGate, groupGate, n, tier, cached)
								t.Run(name, func(t *testing.T) {
									f := newGPT61SubscriptionFixture(resolver)
									f.input.Account.Type = credential
									f.input.Account.Extra["openai_long_context_billing_enabled"] = accountGate
									f.input.APIKey.Group.LongContextPricingEnabled = groupGate
									f.input.Result.Usage.InputTokens = n
									f.input.Result.ServiceTier = &tier
									if cached {
										f.input.Result.Usage.CacheCreationInputTokens = 2000
										f.input.Result.Usage.CacheReadInputTokens = 1000
									}
									require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
									// Preserve the existing legacy no-resolver account-only gate.
									long := credential == AccountTypeAPIKey && n > 272000 && accountGate && (!resolver || groupGate)
									im, om, tm := 1.0, 1.0, 1.0
									if long {
										im, om = 2, 1.5
									}
									if tier == "priority" {
										tm = 2
									}
									write, read := 0, 0
									if cached {
										write, read = 2000, 1000
									}
									in := float64(n-write-read) * 2e-6 * im * tm
									out := .01 * om * tm
									wc := float64(write) * 2.5e-6 * im * tm
									rc := float64(read) * .1e-6 * im * tm
									statsIn, statsOut := 1.0, 1.0
									// Model-file account stats historically do not use account/group
									// billing gates or the request service tier; only exemption is new.
									if credential == AccountTypeAPIKey && n > 272000 {
										statsIn, statsOut = 2, 1.5
									}
									stats := (float64(n-write-read)*2e-6+float64(write)*2.5e-6+float64(read)*.1e-6)*statsIn + .01*statsOut
									f.assertSettlement(t, in+out+wc+rc, 1.3, stats, long)
									require.InDelta(t, in, f.usage.lastLog.InputCost, 1e-10)
									require.InDelta(t, out, f.usage.lastLog.OutputCost, 1e-10)
									require.InDelta(t, wc, f.usage.lastLog.CacheCreationCost, 1e-10)
									require.InDelta(t, rc, f.usage.lastLog.CacheReadCost, 1e-10)
									require.Equal(t, n-write-read, f.usage.lastLog.InputTokens)
									require.Equal(t, write, f.usage.lastLog.CacheCreationTokens)
									require.Equal(t, read, f.usage.lastLog.CacheReadTokens)
								})
							}
						}
					}
				}
			}
		}
	}
}

func TestOpenAISubscriptionGPT61CustomerSubscriptionIsIndependent(t *testing.T) {
	for _, credential := range []string{AccountTypeOAuth, AccountTypeAPIKey} {
		for _, customerSubscription := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/subscription=%t", credential, customerSubscription), func(t *testing.T) {
				f := newGPT61SubscriptionFixture(true)
				f.input.Account.Type = credential
				if customerSubscription {
					f.input.Subscription = &UserSubscription{ID: 71}
					f.input.APIKey.Group.SubscriptionType = SubscriptionTypeSubscription
				}
				require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
				total := .61
				if credential == AccountTypeAPIKey {
					total = 1.215
				}
				f.assertSettlement(t, total, 1.3, total, credential == AccountTypeAPIKey)
			})
		}
	}
}

func TestOpenAISubscriptionGPT61SentModelAndCredentialIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, requested, sent, billed, response, platform, credential string
		exempt                                                        bool
	}{
		{"sent target customer alias", "customer-alias", "gpt-6.1-sol", "gpt-6.1-sol", "gpt-6-sol", PlatformOpenAI, AccountTypeOAuth, true},
		{"sent alias", "customer-alias", " openai/GPT_6.1_SOL_high ", "gpt-6.1-sol", "", PlatformOpenAI, AccountTypeSetupToken, true},
		{"fallback requested target", "gpt-6.1-sol", "", "gpt-6.1-sol", "", PlatformOpenAI, AccountTypeOAuth, true},
		{"mapped away", "gpt-6.1-sol", "gpt-6-sol", "gpt-6.1-sol", "gpt-6.1-sol", PlatformOpenAI, AccountTypeOAuth, false},
		{"only billing and response target", "gpt-6-sol", "gpt-6-sol", "gpt-6.1-sol", "gpt-6.1-sol", PlatformOpenAI, AccountTypeOAuth, false},
		{"unknown similar suffix", "gpt-6.1-sol-unknown", "gpt-6.1-sol-unknown", "gpt-6.1-sol", "", PlatformOpenAI, AccountTypeOAuth, false},
		{"unknown similar prefix", "gpt-6.1-solstice", "gpt-6.1-solstice", "gpt-6.1-sol", "", PlatformOpenAI, AccountTypeOAuth, false},
		{"api key target", "gpt-6.1-sol", "gpt-6.1-sol", "gpt-6.1-sol", "", PlatformOpenAI, AccountTypeAPIKey, false},
		{"other platform oauth", "gpt-6.1-sol", "gpt-6.1-sol", "gpt-6.1-sol", "", PlatformGrok, AccountTypeOAuth, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newGPT61SubscriptionFixture(true)
			// Give unknown names a real card so lack of exemption cannot hide
			// behind pricing-unavailable zero-cost behavior.
			if tc.name == "unknown similar suffix" || tc.name == "unknown similar prefix" {
				f.svc.billingService.pricingService = &PricingService{pricingData: map[string]*LiteLLMModelPricing{
					tc.sent: {InputCostPerToken: 2e-6, OutputCostPerToken: 10e-6,
						LongContextInputTokenThreshold: 272000, LongContextInputCostMultiplier: 2,
						LongContextOutputCostMultiplier: 1.5, LongContextMetadataPresent: true},
				}}
			}
			f.input.Account.Platform, f.input.Account.Type = tc.platform, tc.credential
			f.input.Result.Model, f.input.Result.UpstreamModel = tc.requested, tc.sent
			f.input.Result.BillingModel, f.input.Result.UpstreamResponseModel = tc.billed, tc.response
			require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
			total := 1.215
			if tc.exempt {
				total = .61
			}
			f.assertSettlement(t, total, 1.3, total, !tc.exempt)
		})
	}
}

func TestOpenAISubscriptionGPT61ShadowCredentials(t *testing.T) {
	for _, tc := range []struct {
		name   string
		parent *Account
		err    error
		noRepo bool
		badID  bool
	}{
		{"eligible oauth", &Account{ID: 91, Platform: PlatformOpenAI, Type: AccountTypeOAuth}, nil, false, false},
		{"no repository", nil, nil, true, false},
		{"invalid parent id", nil, nil, false, true},
		{"missing parent", nil, nil, false, false},
		{"lookup failure", nil, errors.New("parent lookup failed"), false, false},
		{"api key parent", &Account{ID: 91, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, nil, false, false},
		{"setup parent not eligible locally", &Account{ID: 91, Platform: PlatformOpenAI, Type: AccountTypeSetupToken}, nil, false, false},
		{"other platform parent", &Account{ID: 91, Platform: PlatformGrok, Type: AccountTypeOAuth}, nil, false, false},
		{"nested shadow", &Account{ID: 91, Platform: PlatformOpenAI, Type: AccountTypeOAuth, ParentAccountID: i64p(92)}, nil, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newGPT61SubscriptionFixture(true)
			f.input.Account.ParentAccountID = i64p(91)
			if tc.badID {
				f.input.Account.ParentAccountID = i64p(0)
			}
			repo := &gpt61SubscriptionAccountRepo{parent: tc.parent, err: tc.err}
			if !tc.noRepo {
				f.svc.accountRepo = repo
			}
			err := f.svc.RecordUsage(context.Background(), f.input)
			if tc.name != "eligible oauth" {
				require.ErrorContains(t, err, "resolve OpenAI usage credential account")
				require.Zero(t, f.usage.calls)
				require.Zero(t, f.user.deductCalls)
				require.Zero(t, f.sub.incrementCalls)
				return
			}
			require.NoError(t, err)
			require.Equal(t, []int64{91}, repo.ids)
			f.assertSettlement(t, .61, 1.3, .61, false)
			require.Equal(t, int64(31), f.usage.lastLog.AccountID, "credential parent must not own billed usage")
		})
	}
}

func TestOpenAISubscriptionGPT61MappedBillingSources(t *testing.T) {
	for _, source := range []string{BillingModelSourceRequested, BillingModelSourceChannelMapped, BillingModelSourceUpstream} {
		for _, sent := range []string{"gpt-6.1-sol", "gpt-6-sol"} {
			t.Run(source+"/"+sent, func(t *testing.T) {
				f := newGPT61SubscriptionFixture(true)
				f.input.OriginalModel = "gpt-6.1-sol"
				f.input.ChannelMappedModel = "gpt-6-sol"
				f.input.BillingModelSource = source
				f.input.Result.Model, f.input.Result.UpstreamModel = "gpt-6.1-sol", sent
				f.input.Result.UpstreamResponseModel = "gpt-6.1-sol"
				require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
				total := 1.215
				if sent == "gpt-6.1-sol" {
					total = .61
				}
				f.assertSettlement(t, total, 1.3, total, sent != "gpt-6.1-sol")
				require.Equal(t, "gpt-6.1-sol", f.usage.lastLog.RequestedModel)
			})
		}
	}
}

func TestOpenAISubscriptionGPT61ShadowPreservesSettlementAndStatsOwner(t *testing.T) {
	f := newGPT61SubscriptionFixture(true)
	parentRate, selectedRate, customStats := 9.0, 2.0, .42
	f.input.Account.ParentAccountID = i64p(91)
	f.input.Account.Type = AccountTypeAPIKey
	f.input.Account.RateMultiplier = &selectedRate
	f.svc.accountRepo = &gpt61SubscriptionAccountRepo{
		parent: &Account{ID: 91, Platform: PlatformOpenAI, Type: AccountTypeOAuth, RateMultiplier: &parentRate},
	}
	f.svc.channelService = gpt61SubscriptionChannel(&Channel{ID: 61, AccountStatsPricingRules: []AccountStatsPricingRule{
		{AccountIDs: []int64{31}, Pricing: []ChannelModelPricing{{Models: []string{"gpt-6.1-sol"}, BillingMode: BillingModePerRequest, PerRequestPrice: &customStats}}},
	}})
	f.svc.resolver = NewModelPricingResolver(f.svc.channelService, f.svc.billingService)
	require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
	f.assertSettlement(t, .61, 1.3, .42, false)
}

func TestOpenAISubscriptionGPT61AdministratorPricing(t *testing.T) {
	flat, interval, zero := 3e-6, 7e-6, 0.0
	short := 100
	for _, tc := range []struct {
		name      string
		pricing   ChannelModelPricing
		wantPrice float64
	}{
		{"flat", ChannelModelPricing{InputPrice: &flat, OutputPrice: &flat, CacheWritePrice: &flat, CacheReadPrice: &flat}, flat},
		{"flat zero", ChannelModelPricing{InputPrice: &zero, OutputPrice: &zero, CacheWritePrice: &zero, CacheReadPrice: &zero}, zero},
		{"interval", ChannelModelPricing{Intervals: []PricingInterval{{MinTokens: 0, InputPrice: &interval, OutputPrice: &interval, CacheWritePrice: &interval, CacheReadPrice: &interval}}}, interval},
		{"interval zero", ChannelModelPricing{Intervals: []PricingInterval{{MinTokens: 0, InputPrice: &zero, OutputPrice: &zero, CacheWritePrice: &zero, CacheReadPrice: &zero}}}, zero},
		{"unmatched interval fallback", ChannelModelPricing{InputPrice: &flat, OutputPrice: &flat, CacheWritePrice: &flat, CacheReadPrice: &flat,
			Intervals: []PricingInterval{{MinTokens: 0, MaxTokens: &short, InputPrice: &interval}}}, flat},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newGPT61SubscriptionFixture(true)
			p := tc.pricing.Clone()
			p.Models, p.BillingMode = []string{"gpt-6.1-sol"}, BillingModeToken
			f.svc.channelService = gpt61SubscriptionChannel(&Channel{ID: 61}, &p)
			f.svc.resolver = NewModelPricingResolver(f.svc.channelService, f.svc.billingService)
			f.input.Result.Usage.CacheCreationInputTokens = 2000
			f.input.Result.Usage.CacheReadInputTokens = 1000
			require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
			in, out, write, read := 297000*tc.wantPrice, 1000*tc.wantPrice, 2000*tc.wantPrice, 1000*tc.wantPrice
			if tc.name == "unmatched interval fallback" {
				// Existing resolver ignores flat token overrides when valid
				// intervals exist; a miss uses the model-file base card.
				in, out, write, read = 297000*2e-6, .01, 2000*2.5e-6, 1000*.1e-6
			}
			total := in + out + write + read
			require.InDelta(t, in, f.usage.lastLog.InputCost, 1e-12)
			require.InDelta(t, out, f.usage.lastLog.OutputCost, 1e-12)
			require.InDelta(t, write, f.usage.lastLog.CacheCreationCost, 1e-12)
			require.InDelta(t, read, f.usage.lastLog.CacheReadCost, 1e-12)
			// Model-file stats remain independent of administrator customer prices.
			stats := 297000*2e-6 + .01 + 2000*2.5e-6 + 1000*.1e-6
			if total == 0 {
				require.False(t, f.usage.lastLog.LongContextBillingApplied)
				require.Zero(t, f.user.deductCalls)
				require.Equal(t, 100.0, f.user.balance)
				require.Zero(t, f.usage.lastLog.TotalCost)
				require.InDelta(t, stats, *f.usage.lastLog.AccountStatsCost, 1e-12)
			} else {
				f.assertSettlement(t, total, 1.3, stats, false)
			}
			require.Equal(t, tc.pricing.Intervals, p.Intervals, "policy must not mutate administrator intervals")
			card, err := f.svc.billingService.GetModelPricing("gpt-6.1-sol")
			require.NoError(t, err)
			require.Equal(t, 272000, card.LongContextInputThreshold)
			require.Equal(t, 2.0, card.LongContextInputMultiplier, "API long-context prices must remain intact")
		})
	}
}

func TestOpenAISubscriptionGPT61AccountStatsPrecedence(t *testing.T) {
	custom := 7.0
	for _, tc := range []struct {
		name          string
		applyCustomer bool
		customRule    bool
		stats         float64
	}{
		{"model file exemption", false, false, .61},
		{"apply customer total", true, false, .9},
		{"custom before model file", false, true, 7},
		{"custom before apply customer", true, true, 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newGPT61SubscriptionFixture(true)
			channel := &Channel{ID: 61, ApplyPricingToAccountStats: tc.applyCustomer}
			if tc.customRule {
				channel.AccountStatsPricingRules = []AccountStatsPricingRule{{AccountIDs: []int64{31},
					Pricing: []ChannelModelPricing{{Models: []string{"gpt-6.1-sol"}, BillingMode: BillingModePerRequest, PerRequestPrice: &custom}}}}
			}
			price := 3e-6
			f.svc.channelService = gpt61SubscriptionChannel(channel, &ChannelModelPricing{
				Models: []string{"gpt-6.1-sol"}, InputPrice: &price, OutputPrice: new(float64),
			})
			f.svc.resolver = NewModelPricingResolver(f.svc.channelService, f.svc.billingService)
			require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
			f.assertSettlement(t, .9, 1.3, tc.stats, false)
		})
	}
	// Older shared signatures keep their default model-file surcharge behavior.
	f := newGPT61SubscriptionFixture(false)
	tokens := UsageTokens{InputTokens: 300000, OutputTokens: 1000}
	require.InDelta(t, 1.215, *tryModelFilePricing(f.svc.billingService, "gpt-6.1-sol", tokens), 1e-12)
	require.InDelta(t, 1.215, *resolveAccountStatsCost(context.Background(), f.svc.channelService, f.svc.billingService, 31, 61, "gpt-6.1-sol", tokens, 1, .61), 1e-12)
}

func TestOpenAISubscriptionGPT61NonTokenPricing(t *testing.T) {
	setTestTimezone(t, "UTC")
	for _, mode := range []BillingMode{BillingModePerRequest, BillingModeImage, BillingModeVideo} {
		t.Run(string(mode), func(t *testing.T) {
			f := newGPT61SubscriptionFixture(true)
			price := .2
			p := &ChannelModelPricing{Models: []string{"gpt-6.1-sol"}, BillingMode: mode, PerRequestPrice: &price}
			f.svc.channelService = gpt61SubscriptionChannel(&Channel{ID: 61, ApplyPricingToAccountStats: true}, p)
			f.svc.resolver = NewModelPricingResolver(f.svc.channelService, f.svc.billingService)
			group := f.input.APIKey.Group
			group.PeakRateEnabled, group.PeakStart, group.PeakEnd, group.PeakRateMultiplier = true, "14:00", "18:00", .7
			f.input.RequestStartedAt = at(17, 59)
			total := .2
			if mode == BillingModeImage {
				f.input.Result.ImageCount = 2
				f.input.Result.ImageSize = "1K"
				total = .4
			}
			if mode == BillingModeVideo {
				f.input.RequestCountOverride = 6
				total = 1.2
			}
			require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
			f.assertSettlement(t, total, 1.3, total, false)
			require.Equal(t, string(mode), *f.usage.lastLog.BillingMode)
		})
	}
}

func TestOpenAISubscriptionGPT61MultipliersOverrideAndDedupe(t *testing.T) {
	setTestTimezone(t, "UTC")
	t.Run("time and user rate", func(t *testing.T) {
		f := newGPT61SubscriptionFixture(true)
		group := f.input.APIKey.Group
		group.PeakRateEnabled, group.PeakStart, group.PeakEnd, group.PeakRateMultiplier = true, "14:00", "18:00", .7
		f.input.RequestStartedAt = at(17, 59)
		require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
		f.assertSettlement(t, .61, 1.3*.7, .61, false)
	})
	t.Run("explicit authoritative override", func(t *testing.T) {
		f := newGPT61SubscriptionFixture(true)
		f.input.CostOverride = &CostBreakdown{TotalCost: 99, ActualCost: 99}
		f.input.Result.CostOverride = &CostBreakdown{InputCost: .5, OutputCost: .2, TotalCost: .7, ActualCost: .91, BillingMode: string(BillingModeToken), LongContextBillingApplied: true}
		require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
		f.assertSettlement(t, .7, 1.3, .61, true)
		require.InDelta(t, .5, f.usage.lastLog.InputCost, 1e-12)
		require.InDelta(t, .2, f.usage.lastLog.OutputCost, 1e-12)
	})
	t.Run("usage log deduplication", func(t *testing.T) {
		f := newGPT61SubscriptionFixture(true)
		events, owned := []string{}, true
		repo := &gatewayTrialSettlementLogRepo{events: &events, owned: &owned}
		f.svc.usageLogRepo = repo
		require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
		owned = false
		require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
		require.Equal(t, 1, f.user.deductCalls)
		require.InDelta(t, 100-.61*1.3, f.user.balance, 1e-12)
		require.Equal(t, []int64{101}, repo.applied)
		require.InDelta(t, .61*1.3, repo.payload.Primary.BalanceCost, 1e-8)
	})
	for _, subscription := range []bool{false, true} {
		t.Run(fmt.Sprintf("unified deduplication/subscription=%t", subscription), func(t *testing.T) {
			f := newGPT61SubscriptionFixture(true)
			repo := &openAIRecordUsageDedupeBillingRepoStub{}
			f.svc.usageBillingRepo = repo
			if subscription {
				f.input.Subscription = &UserSubscription{ID: 71}
				f.input.APIKey.Group.SubscriptionType = SubscriptionTypeSubscription
			}
			for range 2 {
				require.NoError(t, f.svc.RecordUsage(context.Background(), f.input))
			}
			require.Equal(t, 1, repo.appliedCalls)
			require.Zero(t, f.user.deductCalls, "unified repo owns monetary effects")
			require.Zero(t, f.sub.incrementCalls)
			require.Equal(t, int64(31), repo.commands[0].AccountID)
			if subscription {
				require.InDelta(t, .61*1.3, repo.appliedSubCost, 1e-8)
				require.Zero(t, repo.appliedBalanceCost)
			} else {
				require.InDelta(t, .61*1.3, repo.appliedBalanceCost, 1e-8)
				require.Zero(t, repo.appliedSubCost)
			}
			require.False(t, f.usage.lastLog.LongContextBillingApplied)
			require.InDelta(t, .61, *f.usage.lastLog.AccountStatsCost, 1e-12)
		})
	}
}

func TestGPT61DisableModelLongContextPricing(t *testing.T) {
	f := newGPT61SubscriptionFixture(false)
	enabled := true
	for _, resolver := range []bool{false, true} {
		for _, disable := range []bool{false, true} {
			t.Run(fmt.Sprintf("resolver=%t/disable=%t", resolver, disable), func(t *testing.T) {
				input := CostInput{Model: "gpt-6.1-sol", Group: f.input.APIKey.Group, RateMultiplier: 1.3,
					Tokens:                    UsageTokens{InputTokens: 300000, OutputTokens: 1000},
					LongContextBillingEnabled: &enabled, DisableModelLongContextPricing: disable}
				if resolver {
					input.Resolver = NewModelPricingResolver(nil, f.svc.billingService)
				}
				cost, err := f.svc.billingService.CalculateCostUnified(input)
				require.NoError(t, err)
				total := 1.215
				if disable {
					total = .61
				}
				require.InDelta(t, total, cost.TotalCost, 1e-12)
				require.InDelta(t, total*1.3, cost.ActualCost, 1e-12)
				require.Equal(t, !disable, cost.LongContextBillingApplied)
				require.True(t, enabled, "policy must not mutate the normal account gate")
				require.True(t, f.input.APIKey.Group.LongContextPricingEnabled)
			})
		}
	}
}
