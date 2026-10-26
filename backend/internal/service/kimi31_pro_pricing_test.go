package service

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

var kimi31Aliases = []string{"kimi-k3.1-pro", "kimi3.1pro", "kimi-k3-1-pro", "kimi-k3.1pro", "k3.1-pro"}
var kimi31Unknown = []string{"kimi-k3.1-pro-highspeed", "kimi-k3.1-pro-preview", "kimi-k3.1-pro-20261026", "kimi-k3.1-pro[1m]", "kimi-k3.10-pro", "kimi-k3.1-pro-extra", "other/kimi-k3.1-pro", "kimi/moonshot/kimi-k3.1-pro"}

func kimi31Sources(t *testing.T) map[string]*BillingService {
	t.Helper()
	body, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	ps := &PricingService{}
	data, err := ps.parsePricingData(body)
	require.NoError(t, err)
	stale := map[string]*LiteLLMModelPricing{}
	for _, alias := range kimi31Aliases {
		for _, prefix := range []string{"", "kimi/", "moonshot/"} {
			stale[prefix+alias] = &LiteLLMModelPricing{InputCostPerToken: 99, OutputCostPerToken: 99,
				CacheReadInputTokenCost: 99, CacheCreationInputTokenCost: 99, CacheCreationInputTokenCostAbove1hr: 199,
				LongContextInputTokenThreshold: 1, LongContextInputCostMultiplier: 9, LongContextOutputCostMultiplier: 9,
				LongContextMetadataPresent: true}
		}
	}
	// Exact unknown dynamic cards must not defeat fail-closed default isolation.
	for _, model := range kimi31Unknown {
		stale[model] = &LiteLLMModelPricing{InputCostPerToken: 99}
	}
	return map[string]*BillingService{
		"fallback": NewBillingService(&config.Config{}, nil),
		"bundle":   NewBillingService(&config.Config{}, &PricingService{pricingData: data}),
		"stale":    NewBillingService(&config.Config{}, &PricingService{pricingData: stale}),
	}
}

func TestKimi31DefaultSourcesAliasesAndIsolation(t *testing.T) {
	for source, bs := range kimi31Sources(t) {
		var before []byte
		if bs.pricingService != nil {
			before, _ = json.Marshal(bs.pricingService.pricingData)
		}
		for _, alias := range kimi31Aliases {
			for _, prefix := range []string{"", "kimi/", "moonshot/"} {
				model := strings.ToUpper(prefix + alias)
				t.Run(source+"/"+model, func(t *testing.T) {
					for i := 0; i < 2; i++ {
						p, err := bs.GetModelPricing(model)
						require.NoError(t, err)
						require.Equal(t, 13e-6, p.InputPricePerToken)
						require.Equal(t, 68e-6, p.OutputPricePerToken)
						require.Equal(t, 1.3e-6, p.CacheReadPricePerToken)
						require.Equal(t, 13e-6, p.CacheCreationPricePerToken)
						require.Equal(t, 13e-6, p.CacheCreation5mPrice)
						require.Equal(t, 26e-6, p.CacheCreation1hPrice)
						require.True(t, p.SupportsCacheBreakdown)
						require.Equal(t, p, bs.getFallbackPricing(model))
						if bs.pricingService != nil {
							lp := bs.pricingService.GetModelPricing(model)
							require.NotNil(t, lp)
							require.Equal(t, 13e-6, lp.InputCostPerToken)
							require.Equal(t, 68e-6, lp.OutputCostPerToken)
							require.Equal(t, 1.3e-6, lp.CacheReadInputTokenCost)
							require.Equal(t, 13e-6, lp.CacheCreationInputTokenCost)
							require.Equal(t, 26e-6, lp.CacheCreationInputTokenCostAbove1hr)
						}
						resolver := NewModelPricingResolver(nil, bs)
						resolved := resolver.Resolve(context.Background(), PricingInput{Model: model})
						require.Equal(t, p, resolved.BasePricing)
						for _, mult := range []float64{1, .25} {
							tokens := UsageTokens{InputTokens: 1000, OutputTokens: 1000, CacheReadTokens: 1000, CacheCreationTokens: 1000}
							legacy, err := bs.CalculateCost(model, tokens, mult)
							require.NoError(t, err)
							unified, err := bs.CalculateCostUnified(CostInput{Ctx: context.Background(), Model: model, Tokens: tokens, RateMultiplier: mult, Resolver: resolver})
							require.NoError(t, err)
							require.InDelta(t, .0953, legacy.TotalCost, 1e-12)
							require.InDelta(t, .0953*mult, legacy.ActualCost, 1e-12)
							require.Equal(t, string(BillingModeToken), unified.BillingMode)
							unified.BillingMode = legacy.BillingMode
							require.Equal(t, legacy, unified)
						}
					}
				})
			}
		}
		for _, model := range kimi31Unknown {
			t.Run(source+"/unknown/"+model, func(t *testing.T) {
				_, err := bs.GetModelPricing(model)
				require.ErrorIs(t, err, ErrModelPricingUnavailable)
				require.Nil(t, bs.getFallbackPricing(model))
				if bs.pricingService != nil {
					require.Nil(t, bs.pricingService.GetModelPricing(model))
				}
				_, err = bs.CalculateCostUnified(CostInput{Model: model, Tokens: UsageTokens{InputTokens: 1000}, RateMultiplier: 1, Resolver: NewModelPricingResolver(nil, bs)})
				require.Error(t, err)
			})
		}
		if bs.pricingService != nil {
			after, err := json.Marshal(bs.pricingService.pricingData)
			require.NoError(t, err)
			require.Equal(t, string(before), string(after), "shared source cards must remain untouched")
		}
	}
}

func TestKimi31CacheNormalizationAndNoSurcharge(t *testing.T) {
	for source, bs := range kimi31Sources(t) {
		for _, tc := range []struct {
			name                  string
			aggregate, five, hour int
			want                  float64
		}{
			{"aggregate", 1000, 0, 0, .013}, {"five", 1000, 1000, 0, .013},
			{"hour", 1000, 0, 1000, .026}, {"split", 1000, 600, 400, .0182},
			{"details-only", 0, 600, 400, .0182}, {"contradictory", 1000, 1200, 800, .0182},
			{"remainder", 1000, 200, 400, .013}, {"negative", 1000, -100, 0, .013},
		} {
			t.Run(source+"/"+tc.name, func(t *testing.T) {
				for _, mult := range []float64{1, .25} {
					for _, resolver := range []*ModelPricingResolver{nil, NewModelPricingResolver(nil, bs)} {
						bd, err := bs.CalculateCostUnified(CostInput{Model: "kimi-k3.1-pro", RateMultiplier: mult, Resolver: resolver,
							Tokens: UsageTokens{InputTokens: 1000, OutputTokens: 1000, CacheReadTokens: 1000, CacheCreationTokens: tc.aggregate, CacheCreation5mTokens: tc.five, CacheCreation1hTokens: tc.hour}})
						require.NoError(t, err)
						require.InDelta(t, .013, bd.InputCost, 1e-12)
						require.InDelta(t, .068, bd.OutputCost, 1e-12)
						require.InDelta(t, .0013, bd.CacheReadCost, 1e-12)
						require.InDelta(t, tc.want, bd.CacheCreationCost, 1e-12)
						require.InDelta(t, (.0823+tc.want)*mult, bd.ActualCost, 1e-12)
					}
				}
			})
		}
		for _, at := range []string{"2026-10-26T00:00:00Z", "2026-10-26T12:00:00Z"} {
			bd, err := bs.CalculateCostUnified(CostInput{Model: "kimi-k3.1-pro", RateMultiplier: 1,
				PricingAt: newModelBilling20261022Time(at), Tokens: UsageTokens{InputTokens: 1000000, OutputTokens: 1000},
				Group: &Group{LongContextPricingEnabled: true}, Resolver: NewModelPricingResolver(nil, bs)})
			require.NoError(t, err)
			require.InDelta(t, 13.068, bd.TotalCost, 1e-12)
		}
	}
}

func TestKimi31CustomPricingBoundaries(t *testing.T) {
	for _, model := range append(append([]string{}, kimi31Aliases...), kimi31Unknown...) {
		for _, source := range []string{"channel", "group"} {
			for _, price := range []float64{0, 7e-6} {
				bs := NewBillingService(&config.Config{}, nil)
				cp := &ChannelModelPricing{Models: []string{model}, BillingMode: BillingModeToken,
					InputPrice: &price, OutputPrice: &price, CacheReadPrice: &price, CacheWritePrice: &price}
				group := &Group{ID: 77, LongContextPricingEnabled: true}
				cs := newModelBilling20261022Channel(group.ID, model, cp)
				if source == "group" {
					group.ModelPricing = []ChannelModelPricing{*cp}
				}
				resolver := NewModelPricingResolver(cs, bs)
				bd, err := bs.CalculateCostUnified(CostInput{Ctx: context.Background(), Model: model, GroupID: &group.ID, Group: group,
					RateMultiplier: .25, Resolver: resolver, Tokens: UsageTokens{InputTokens: 1000, OutputTokens: 1000, CacheReadTokens: 1000, CacheCreationTokens: 1000, CacheCreation5mTokens: 600, CacheCreation1hTokens: 400}})
				require.NoError(t, err, source+"/"+model)
				require.InDelta(t, 4000*price, bd.TotalCost, 1e-12)
				require.InDelta(t, 1000*price, bd.ActualCost, 1e-12)
				if source == "channel" {
					legacy, err := bs.calculateCostInternal(model, UsageTokens{InputTokens: 1000, OutputTokens: 1000, CacheReadTokens: 1000, CacheCreationTokens: 1000}, .25, "", cp)
					require.NoError(t, err)
					require.InDelta(t, bd.TotalCost, legacy.TotalCost, 1e-12)
				}
			}
		}
	}
	bs := NewBillingService(&config.Config{}, nil)
	price := 7e-6
	cp := &ChannelModelPricing{Models: []string{"kimi-k3.1-pro"}, InputPrice: &price}
	p, err := bs.GetModelPricingWithChannel("kimi-k3.1-pro", cp)
	require.NoError(t, err)
	require.Equal(t, price, p.InputPricePerToken)
	require.Equal(t, 68e-6, p.OutputPricePerToken)
	require.Equal(t, 26e-6, p.CacheCreation1hPrice)
	for _, mode := range []BillingMode{BillingModeToken, BillingModePerRequest, BillingModeImage} {
		cp := &ChannelModelPricing{Models: []string{"kimi-k3.1-pro"}, BillingMode: mode, PerRequestPrice: &price,
			Intervals: []PricingInterval{{MinTokens: 0, InputPrice: &price, OutputPrice: &price, CacheWritePrice: &price, CacheReadPrice: &price}}}
		cs := newModelBilling20261022Channel(77, "kimi-k3.1-pro", cp)
		id := int64(77)
		got, err := bs.CalculateCostUnified(CostInput{Ctx: context.Background(), Model: "kimi-k3.1-pro", GroupID: &id,
			Tokens: UsageTokens{InputTokens: 1000}, RateMultiplier: .25, RequestCount: 2, Resolver: NewModelPricingResolver(cs, bs)})
		require.NoError(t, err)
		want := price * 2
		if mode == BillingModeToken {
			want = price * 1000
		}
		require.InDelta(t, want, got.TotalCost, 1e-12)
		require.InDelta(t, want*.25, got.ActualCost, 1e-12)
	}
}

func TestKimi31PartialZeroIntervalsAndRequestModes(t *testing.T) {
	for _, model := range []string{"kimi-k3.1-pro", "kimi-k3.1-pro-highspeed"} {
		for _, source := range []string{PricingSourceChannel, PricingSourceGroup} {
			for _, price := range []float64{0, 7e-6} {
				for _, mode := range []BillingMode{BillingModeToken, BillingModePerRequest, BillingModeImage} {
					bs := NewBillingService(&config.Config{}, nil)
					cp := &ChannelModelPricing{Models: []string{model}, BillingMode: mode, InputPrice: &price, PerRequestPrice: &price}
					group := &Group{ID: 77}
					if source == PricingSourceGroup {
						group.ModelPricing = []ChannelModelPricing{*cp}
					}
					cs := newModelBilling20261022Channel(77, model, cp)
					got, err := bs.CalculateCostUnified(CostInput{Ctx: context.Background(), Model: model, GroupID: &group.ID, Group: group,
						Tokens:         UsageTokens{InputTokens: 1000, OutputTokens: 1000, CacheReadTokens: 1000, CacheCreationTokens: 1000},
						RateMultiplier: .25, RequestCount: 2, Resolver: NewModelPricingResolver(cs, bs)})
					require.NoError(t, err)
					want := price * 2
					if mode == BillingModeToken {
						want = price * 1000
						if model == "kimi-k3.1-pro" {
							want += .068 + .0013 + .013
						}
					}
					require.InDelta(t, want, got.TotalCost, 1e-12)
					require.InDelta(t, want*.25, got.ActualCost, 1e-12)
				}
				// Channel intervals own all token prices, including explicit zeros.
				bs := NewBillingService(&config.Config{}, nil)
				cp := &ChannelModelPricing{Models: []string{model}, BillingMode: BillingModeToken,
					Intervals: []PricingInterval{{MinTokens: 0, InputPrice: &price, OutputPrice: &price, CacheWritePrice: &price, CacheReadPrice: &price}}}
				cs := newModelBilling20261022Channel(77, model, cp)
				id := int64(77)
				got, err := bs.CalculateCostUnified(CostInput{Ctx: context.Background(), Model: model, GroupID: &id,
					Tokens: UsageTokens{InputTokens: 1000, OutputTokens: 1000, CacheReadTokens: 1000, CacheCreationTokens: 1000,
						CacheCreation5mTokens: 600, CacheCreation1hTokens: 400},
					RateMultiplier: 1, Resolver: NewModelPricingResolver(cs, bs)})
				require.NoError(t, err)
				require.InDelta(t, 4000*price, got.TotalCost, 1e-12)
			}
		}
	}
}

func TestKimi31CatalogCanonicalMetadata(t *testing.T) {
	body, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	var data map[string]map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(body, &data))
	entry := data["kimi-k3.1-pro"]
	require.NotNil(t, entry)
	require.JSONEq(t, "1048576", string(entry["max_input_tokens"]))
	require.NotContains(t, entry, "max_output_tokens")
	require.NotContains(t, entry, "max_tokens")
	for _, alias := range kimi31Aliases[1:] {
		require.NotContains(t, data, alias)
	}
}
