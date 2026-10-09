package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestBillingLongContext20261009Gemini(t *testing.T) {
	for _, model := range []string{"gemini-3.1-pro-preview", "gemini-3.1-pro", "models/gemini-3.1-pro"} {
		for _, enabled := range []bool{true, false} {
			for _, resolve := range []bool{true, false} {
				for _, haveResolver := range []bool{true, false} {
					b := NewBillingService(&config.Config{}, bundle20261009(t))
					r := NewModelPricingResolver(nil, b)
					group := &Group{ID: 1, LongContextPricingEnabled: enabled}
					req := TokenCostRequest{Ctx: context.Background(), Model: model, Group: group,
						Tokens: UsageTokens{InputTokens: 300000, OutputTokens: 1000}, RateMultiplier: 1,
						LegacyLongContext: b.LegacyLongContextRule(PlatformGemini)}
					if haveResolver {
						req.Resolver = r
					}
					if resolve {
						req.Resolved = r.Resolve(req.Ctx, PricingInput{Model: model, Group: group})
					}
					c, err := b.CalculateTokenCostForRequest(req)
					require.NoError(t, err)
					in, out := .6, .012
					if enabled {
						in, out = 1.2, .018
					}
					require.InDelta(t, in, c.InputCost, 1e-12)
					require.InDelta(t, out, c.OutputCost, 1e-12)
					require.Equal(t, enabled, c.LongContextBillingApplied)
					unified, err := b.CalculateCostUnified(b.tokenCostInput(req))
					require.NoError(t, err)
					require.InDelta(t, c.TotalCost, unified.TotalCost, 1e-12)
				}
			}
		}
		b := NewBillingService(&config.Config{}, bundle20261009(t))
		c, err := b.CalculateCostWithLongContext(model, UsageTokens{InputTokens: 300000, OutputTokens: 1000}, 1, 200000, 2)
		require.NoError(t, err)
		require.InDelta(t, 1.2, c.InputCost, 1e-12)
		require.InDelta(t, .018, c.OutputCost, 1e-12)
	}
	for _, fields := range []string{`,"long_context_input_token_threshold":0`,
		`,"long_context_input_cost_multiplier":2`, `,"long_context_input_token_threshold":200000`,
		`,"input_cost_per_token_above_200k_tokens":2e-6`} {
		ps := pricing20261009(t, `{"gemini-3.1-pro-preview":{"input_cost_per_token":2e-6,"output_cost_per_token":12e-6`+fields+`}}`)
		b := NewBillingService(&config.Config{}, ps)
		tokens := UsageTokens{InputTokens: 300000, OutputTokens: 1000}
		direct, err := b.CalculateCostWithLongContext("gemini-3.1-pro-preview", tokens, 1, 200000, 2)
		require.NoError(t, err)
		require.InDelta(t, .6, direct.InputCost, 1e-12)
		require.InDelta(t, .012, direct.OutputCost, 1e-12)
		require.False(t, direct.LongContextBillingApplied)
		native, err := b.CalculateTokenCostForRequest(TokenCostRequest{Model: "gemini-3.1-pro-preview", Tokens: tokens,
			Group: &Group{LongContextPricingEnabled: true}, RateMultiplier: 1, LegacyLongContext: b.LegacyLongContextRule(PlatformGemini)})
		require.NoError(t, err)
		require.InDelta(t, direct.TotalCost, native.TotalCost, 1e-12)
	}
	// Only metadata-free cards retain marginal legacy compatibility.
	b := NewBillingService(&config.Config{}, pricing20261009(t, `{"gemini-3.1-pro":{"input_cost_per_token":2e-6,"output_cost_per_token":12e-6}}`))
	c, err := b.CalculateTokenCostForRequest(TokenCostRequest{Model: "gemini-3.1-pro", Tokens: UsageTokens{InputTokens: 300000, OutputTokens: 1000},
		RateMultiplier: 1, LegacyLongContext: b.LegacyLongContextRule(PlatformGemini)})
	require.NoError(t, err)
	require.InDelta(t, .6, c.InputCost, 1e-12)
	require.InDelta(t, .012, c.OutputCost, 1e-12)
	require.InDelta(t, .812, c.ActualCost, 1e-12)
}

func TestBillingLongContext20261009TiersGatesCache(t *testing.T) {
	for _, ps := range []*PricingService{nil, bundle20261009(t)} {
		b := NewBillingService(&config.Config{}, ps)
		r := NewModelPricingResolver(nil, b)
		for _, model := range []string{"gpt-6", "gpt-6-astra", "openai/gpt-6-astra-high", "gpt-6.1-sol", "gpt-6-sol", "gpt-6-luna"} {
			base, err := b.GetModelPricing(model)
			require.NoError(t, err)
			for _, n := range []int{271999, 272000, 272001} {
				for _, tier := range []string{"", "priority", "flex", "ultrafast"} {
					for _, groupGate := range []bool{true, false} {
						for _, accountGate := range []bool{true, false} {
							tokens := UsageTokens{InputTokens: n - 3000, CacheReadTokens: 1000, CacheCreationTokens: 2000, OutputTokens: 1000}
							c, err := b.CalculateCostUnified(CostInput{Ctx: context.Background(), Model: model,
								Group: &Group{LongContextPricingEnabled: groupGate}, Resolver: r, Tokens: tokens,
								ServiceTier: tier, RateMultiplier: 1, LongContextBillingEnabled: &accountGate})
							require.NoError(t, err)
							mult := serviceTierCostMultiplier(tier)
							if tier == "ultrafast" && isOpenAIGPT6AstraPricingModel(model) {
								mult = 6
							}
							im, om := 1.0, 1.0
							if n > 272000 && groupGate && accountGate {
								im, om = 2, 1.5
							}
							require.InDelta(t, float64(tokens.InputTokens)*base.InputPricePerToken*mult*im, c.InputCost, 1e-10, model+"/"+tier)
							require.InDelta(t, 1000*base.OutputPricePerToken*mult*om, c.OutputCost, 1e-10)
							require.InDelta(t, 1000*base.CacheReadPricePerToken*mult*im, c.CacheReadCost, 1e-10)
							require.InDelta(t, 2000*base.CacheCreationPricePerToken*mult*im, c.CacheCreationCost, 1e-10)
							require.Equal(t, im > 1, c.LongContextBillingApplied)
						}
					}
				}
			}
		}
	}
	// 5m/1h counts partition cache creation; they must not count twice.
	b := NewBillingService(&config.Config{}, nil)
	p := &ModelPricing{InputPricePerToken: 2e-6, OutputPricePerToken: 10e-6, CacheReadPricePerToken: .1e-6,
		CacheCreation5mPrice: 2.5e-6, CacheCreation1hPrice: 4e-6, SupportsCacheBreakdown: true,
		LongContextInputThreshold: 272000, LongContextInputMultiplier: 2, LongContextOutputMultiplier: 0, LongContextMetadataPresent: true}
	tokens := UsageTokens{InputTokens: 269000, CacheReadTokens: 1000, CacheCreationTokens: 2000, CacheCreation5mTokens: 1000, CacheCreation1hTokens: 1000, OutputTokens: 1000}
	c := b.computeTokenBreakdown(p, tokens, 1, "", true)
	require.False(t, c.LongContextBillingApplied)
	tokens.InputTokens++
	c = b.computeTokenBreakdown(p, tokens, 1, "", true)
	require.InDelta(t, .013, c.CacheCreationCost, 1e-12)
	require.InDelta(t, .01, c.OutputCost, 1e-12)
	require.True(t, c.LongContextBillingApplied)
	c = b.computeTokenBreakdown(p, tokens, 0, "", true)
	require.False(t, c.LongContextBillingApplied)
	zero := *p
	zero.InputPricePerToken = 0
	zero.OutputPricePerToken = 0
	zero.CacheReadPricePerToken = 0
	zero.CacheCreation5mPrice = 0
	zero.CacheCreation1hPrice = 0
	require.False(t, b.computeTokenBreakdown(&zero, tokens, 1, "", true).LongContextBillingApplied)
}

func TestBillingLongContext20261009OverridesIntervals(t *testing.T) {
	b := NewBillingService(&config.Config{}, bundle20261009(t))
	r := NewModelPricingResolver(nil, b)
	for _, model := range []string{"gemini-3.1-pro", "gpt-6-astra"} {
		for _, price := range []float64{0, 3e-6} {
			for _, intervals := range []bool{false, true} {
				p, err := b.GetModelPricing(model)
				require.NoError(t, err)
				resolved := &ResolvedPricing{Mode: BillingModeToken, Source: PricingSourceChannel, BasePricing: p, longContextPricingEnabled: true}
				max := 100
				if intervals {
					resolved.Intervals = []PricingInterval{{MinTokens: 0, MaxTokens: &max, InputPrice: &price, OutputPrice: &price, CacheWritePrice: &price, CacheReadPrice: &price}}
				} else {
					group := &Group{LongContextPricingEnabled: true, ModelPricing: []ChannelModelPricing{{Models: []string{model}, InputPrice: &price, OutputPrice: &price, CacheWritePrice: &price, CacheReadPrice: &price}}}
					resolved = r.Resolve(context.Background(), PricingInput{Model: model, Group: group})
					require.True(t, resolved.BasePricing.LongContextMetadataPresent)
				}
				for _, n := range []int{50, 300000} {
					for _, tier := range []string{"", "ultrafast"} {
						input := CostInput{Model: model, Resolved: resolved, Resolver: r, Tokens: UsageTokens{InputTokens: n, OutputTokens: 1000}, RateMultiplier: 1, ServiceTier: tier}
						c, err := b.CalculateCostUnified(input)
						require.NoError(t, err)
						if tier == "" {
							native, err := b.CalculateTokenCostForRequest(TokenCostRequest{
								Model: model, Resolved: resolved, Resolver: r, Tokens: input.Tokens,
								RateMultiplier: 1, LegacyLongContext: b.LegacyLongContextRule(PlatformGemini),
							})
							require.NoError(t, err)
							require.InDelta(t, c.ActualCost, native.ActualCost, 1e-12)
						}
						if intervals {
							base := r.GetIntervalPricing(resolved, n)
							want := b.computeTokenBreakdown(b.applyModelSpecificPricingPolicy(model, base), input.Tokens, 1, tier, false)
							require.InDelta(t, want.TotalCost, c.TotalCost, 1e-12)
							require.False(t, c.LongContextBillingApplied)
						} else {
							m := 1.0
							if tier == "ultrafast" {
								m = 2
								if isOpenAIGPT6AstraPricingModel(model) {
									m = 6
								}
							}
							lm := 1.0
							if n > p.LongContextInputThreshold {
								lm = 2
							}
							require.InDelta(t, float64(n)*price*m*lm, c.InputCost, 1e-12)
						}
					}
				}
			}
		}
	}
	// Explicit zero cache prices survive parsing, policy and override cloning.
	ps := pricing20261009(t, `{"gpt-6.1-sol":{"input_cost_per_token":2e-6,"output_cost_per_token":10e-6,"cache_creation_input_token_cost":0,"cache_read_input_token_cost":0,"long_context_input_token_threshold":200000,"long_context_input_cost_multiplier":2}}`)
	b = NewBillingService(&config.Config{}, ps)
	c, err := b.CalculateCost("gpt-6.1-sol", UsageTokens{InputTokens: 300000, CacheCreationTokens: 1000, CacheReadTokens: 1000}, 1)
	require.NoError(t, err)
	require.Zero(t, c.CacheCreationCost)
	require.Zero(t, c.CacheReadCost)
}

func TestBillingLongContext20261009RecordUsageFlag(t *testing.T) {
	for _, enabled := range []bool{true, false} {
		for _, zeroRate := range []bool{true, false} {
			cfg := &config.Config{}
			usage := &peakPerRequestUsageRepo{}
			user := &peakPerRequestUserRepo{}
			b := NewBillingService(cfg, bundle20261009(t))
			svc := NewGatewayService(nil, nil, usage, nil, user, nil, nil, nil, cfg, nil, nil, b, nil, &BillingCacheService{}, nil, nil, &DeferredService{},
				nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			svc.resolver = NewModelPricingResolver(nil, b)
			rate := 1.0
			if zeroRate {
				rate = 0
			}
			group := &Group{ID: 71, Platform: PlatformGemini, LongContextPricingEnabled: enabled, RateMultiplier: rate, SubscriptionType: SubscriptionTypeStandard}
			err := svc.RecordUsageWithLongContext(context.Background(), &RecordUsageLongContextInput{
				Result: &ForwardResult{RequestID: "pricing-longcontext", Model: "gemini-3.1-pro-preview", Usage: ClaudeUsage{InputTokens: 300000, OutputTokens: 1000}, Duration: time.Second},
				APIKey: &APIKey{ID: 81, Group: group, GroupID: &group.ID}, User: &User{ID: 91}, Account: &Account{ID: 101, Platform: PlatformGemini},
				LongContextThreshold: 200000, LongContextMultiplier: 2})
			require.NoError(t, err)
			require.NotNil(t, usage.lastLog)
			require.Equal(t, enabled && !zeroRate, usage.lastLog.LongContextBillingApplied)
			in, out := .6, .012
			if enabled {
				in, out = 1.2, .018
			}
			require.InDelta(t, in, usage.lastLog.InputCost, 1e-12)
			require.InDelta(t, out, usage.lastLog.OutputCost, 1e-12)
			require.InDelta(t, rate*usage.lastLog.TotalCost, usage.lastLog.ActualCost, 1e-12)
		}
	}
}

func TestBillingLongContext20261009GeminiCacheWriteCosts(t *testing.T) {
	b := NewBillingService(&config.Config{}, bundle20261009(t))
	r := NewModelPricingResolver(nil, b)
	for _, model := range []string{"gemini-3.1-pro", "gemini-3.1-pro-preview"} {
		for _, tc := range []struct {
			name      string
			input     int
			tier      string
			groupOn   bool
			writeCost float64
			applied   bool
		}{
			{"standard-below", 198999, "", true, .002, false},
			{"standard-at", 199000, "", true, .002, false},
			{"standard-above", 199001, "", true, .004, true},
			{"standard-long", 200000, "", true, .004, true},
			{"standard-gate-off", 300000, "", false, .002, false},
			{"priority-at", 199000, "priority", true, .0036, false},
			{"priority-above", 200000, "priority", true, .0072, true},
			{"priority-gate-off", 300000, "priority", false, .0036, false},
		} {
			t.Run(model+"/"+tc.name, func(t *testing.T) {
				group := &Group{LongContextPricingEnabled: tc.groupOn}
				tokens := UsageTokens{InputTokens: tc.input, CacheCreationTokens: 1000}
				cost, err := b.CalculateCostUnified(CostInput{
					Ctx: context.Background(), Model: model, Group: group,
					Tokens: tokens, RateMultiplier: 1, ServiceTier: tc.tier, Resolver: r,
				})
				require.NoError(t, err)
				require.InDelta(t, tc.writeCost, cost.CacheCreationCost, 1e-12)
				require.Equal(t, tc.applied, cost.LongContextBillingApplied)
				if tc.tier == "" {
					for _, withResolver := range []bool{true, false} {
						req := TokenCostRequest{
							Ctx: context.Background(), Model: model, Group: group,
							Tokens: tokens, RateMultiplier: 1,
							LegacyLongContext: b.LegacyLongContextRule(PlatformGemini),
						}
						if withResolver {
							req.Resolver = r
						}
						native, err := b.CalculateTokenCostForRequest(req)
						require.NoError(t, err)
						require.InDelta(t, tc.writeCost, native.CacheCreationCost, 1e-12)
						require.InDelta(t, cost.ActualCost, native.ActualCost, 1e-12)
						require.Equal(t, tc.applied, native.LongContextBillingApplied)
					}
				}
			})
		}
		t.Run(model+"/explicit-zero-override", func(t *testing.T) {
			zero := 0.0
			group := &Group{
				LongContextPricingEnabled: true,
				ModelPricing: []ChannelModelPricing{{
					Models: []string{model}, CacheWritePrice: &zero,
				}},
			}
			tokens := UsageTokens{InputTokens: 200000, CacheCreationTokens: 1000}
			resolved := r.Resolve(context.Background(), PricingInput{Model: model, Group: group})
			require.True(t, resolved.BasePricing.CacheCreationPriceExplicit)
			for _, tier := range []string{"", "priority"} {
				cost, err := b.CalculateCostUnified(CostInput{
					Ctx: context.Background(), Model: model, Group: group,
					Tokens: tokens, RateMultiplier: 1, ServiceTier: tier, Resolver: r, Resolved: resolved,
				})
				require.NoError(t, err)
				require.Zero(t, cost.CacheCreationCost)
			}
			native, err := b.CalculateTokenCostForRequest(TokenCostRequest{
				Ctx: context.Background(), Model: model, Group: group, Tokens: tokens,
				RateMultiplier: 1, Resolver: r, LegacyLongContext: b.LegacyLongContextRule(PlatformGemini),
			})
			require.NoError(t, err)
			require.Zero(t, native.CacheCreationCost)
		})
	}
}
