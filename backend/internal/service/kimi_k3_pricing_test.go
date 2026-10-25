package service

import (
	"os"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestKimiK3PricingSourcesAndAliases(t *testing.T) {
	body, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	data, err := (&PricingService{}).parsePricingData(body)
	require.NoError(t, err)
	card := data["kimi-k3"]
	require.NotNil(t, card)
	require.Equal(t, 20e-6, card.InputCostPerToken)
	require.Equal(t, 100e-6, card.OutputCostPerToken)
	require.Equal(t, 2e-6, card.CacheReadInputTokenCost)
	require.Equal(t, 20e-6, card.CacheCreationInputTokenCost)
	require.Equal(t, 40e-6, card.CacheCreationInputTokenCostAbove1hr)

	aliases := []string{"kimi-k3", "k3", "k3-256k", "moonshot/kimi-k3", "kimi/k3", "kimi/k3-256k", "K3"}
	staleData := make(map[string]*LiteLLMModelPricing)
	for _, alias := range aliases {
		staleData[alias] = &LiteLLMModelPricing{
			InputCostPerToken: 3e-6, OutputCostPerToken: 15e-6, CacheReadInputTokenCost: 0.3e-6,
		}
	}
	for source, ps := range map[string]*PricingService{
		"fallback": nil,
		"bundle":   {pricingData: data},
		"stale":    {pricingData: staleData},
	} {
		t.Run(source, func(t *testing.T) {
			bs := NewBillingService(&config.Config{}, ps)
			for _, alias := range aliases {
				t.Run(alias, func(t *testing.T) {
					p, err := bs.GetModelPricing(alias)
					require.NoError(t, err)
					require.Equal(t, 20e-6, p.InputPricePerToken)
					require.Equal(t, 100e-6, p.OutputPricePerToken)
					require.Equal(t, 2e-6, p.CacheReadPricePerToken)
					for _, resolver := range []*ModelPricingResolver{nil, NewModelPricingResolver(nil, bs)} {
						for _, tokens := range []UsageTokens{
							{InputTokens: 1000, OutputTokens: 1000, CacheReadTokens: 1000, CacheCreationTokens: 1000},
							{InputTokens: 1000, OutputTokens: 1000, CacheReadTokens: 1000, CacheCreationTokens: 1000,
								CacheCreation5mTokens: 600, CacheCreation1hTokens: 400},
						} {
							bd, err := bs.CalculateCostUnified(CostInput{
								Model: alias, Tokens: tokens, RateMultiplier: 2, Resolver: resolver,
							})
							require.NoError(t, err)
							require.InDelta(t, 0.02, bd.InputCost, 1e-15)
							require.InDelta(t, 0.1, bd.OutputCost, 1e-15)
							require.InDelta(t, 0.002, bd.CacheReadCost, 1e-15)
							writeCost := 0.02
							if tokens.CacheCreation1hTokens > 0 {
								writeCost = 0.028
							}
							require.InDelta(t, writeCost, bd.CacheCreationCost, 1e-15)
							require.InDelta(t, (0.122+writeCost)*2, bd.ActualCost, 1e-15)
						}
					}
				})
			}
		})
	}
	require.Equal(t, 3e-6, staleData["kimi-k3"].InputCostPerToken)
}

func TestKimiK3PricingPolicyIsolation(t *testing.T) {
	bs := NewBillingService(&config.Config{}, nil)
	original := &ModelPricing{InputPricePerToken: 3e-6, OutputPricePerToken: 15e-6, CacheReadPricePerToken: .3e-6}
	got := bs.applyModelSpecificPricingPolicy("k3", original)
	require.NotSame(t, original, got)
	require.Equal(t, 3e-6, original.InputPricePerToken)
	require.Equal(t, 15e-6, original.OutputPricePerToken)
	require.Equal(t, .3e-6, original.CacheReadPricePerToken)
	for _, model := range []string{"kimi-k30", "foo-k3-bar", "kimi-k3[1m]", "kimi-k2.6", "kimi-for-coding"} {
		require.False(t, isKimiK3Model(model), model)
		require.Same(t, original, bs.applyModelSpecificPricingPolicy(model, original), model)
	}
	require.Nil(t, bs.applyModelSpecificPricingPolicy("k3", nil))
}
