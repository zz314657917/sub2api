//go:build unit

package service

import (
	"os"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestNewModelPricingCatalogFallbackAndContext(t *testing.T) {
	data, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	catalog := &PricingService{}
	catalog.pricingData, err = catalog.parsePricingData(data)
	require.NoError(t, err)

	sources := map[string]*BillingService{
		"billing fallback": NewBillingService(&config.Config{}, nil),
		"pricing fallback": NewBillingService(&config.Config{}, &PricingService{pricingData: map[string]*LiteLLMModelPricing{
			"gpt-6":         {InputCostPerToken: 10e-6, OutputCostPerToken: 50e-6},
			"claude-opus-5": {InputCostPerToken: 5e-6, OutputCostPerToken: 25e-6},
		}}),
		"catalog": NewBillingService(&config.Config{}, catalog),
	}
	for source, svc := range sources {
		for _, tc := range []struct {
			model                      string
			input, output, write, read float64
		}{
			{"gpt-6-sol", 2e-6, 10e-6, 2.5e-6, 0.2e-6},
			{"gpt-6-luna", 0.1e-6, 0.5e-6, 0.125e-6, 0.01e-6},
		} {
			t.Run(source+"/"+tc.model, func(t *testing.T) {
				for _, n := range []int{271999, 272000, 272001} {
					for tier, mult := range map[string]float64{"": 1, "priority": 2, "fast": 2, "ultrafast": 2, "flex": 0.5} {
						tokens := UsageTokens{InputTokens: n - 3000, CacheReadTokens: 2000, CacheCreationTokens: 1000, OutputTokens: 500}
						cost, err := svc.CalculateCostWithServiceTier(tc.model, tokens, 1, tier)
						require.NoError(t, err)
						im, om := 1.0, 1.0
						if n > 272000 {
							im, om = 2, 1.5
						}
						require.InDelta(t, float64(tokens.InputTokens)*tc.input*im*mult, cost.InputCost, 1e-10)
						require.InDelta(t, 1000*tc.write*im*mult, cost.CacheCreationCost, 1e-10)
						require.InDelta(t, 2000*tc.read*im*mult, cost.CacheReadCost, 1e-10)
						require.InDelta(t, 500*tc.output*om*mult, cost.OutputCost, 1e-10)
						require.Equal(t, n > 272000, cost.LongContextBillingApplied)
					}
				}
			})
		}
		t.Run(source+"/opus", func(t *testing.T) {
			tokens := UsageTokens{InputTokens: 300000, OutputTokens: 500, CacheReadTokens: 1000,
				CacheCreationTokens: 1000, CacheCreation5mTokens: 400, CacheCreation1hTokens: 600}
			for tier, mult := range map[string]float64{"": 1, "fast": 2, "priority": 2} {
				cost, err := svc.CalculateCostWithServiceTier("claude-opus-5-5", tokens, 1, tier)
				require.NoError(t, err)
				require.InDelta(t, 300000*4e-6*mult, cost.InputCost, 1e-10)
				require.InDelta(t, (400*5e-6+600*8e-6)*mult, cost.CacheCreationCost, 1e-10)
				require.InDelta(t, 1000*0.2e-6*mult, cost.CacheReadCost, 1e-10)
				require.InDelta(t, 500*20e-6*mult, cost.OutputCost, 1e-10)
				require.False(t, cost.LongContextBillingApplied)
			}
		})
	}
}

func TestNewModelPricingUnderscoreLunaFallback(t *testing.T) {
	svc := NewBillingService(&config.Config{}, nil)
	for _, model := range []string{"GPT_6_LUNA_XHIGH", "openai/GPT_6_LUNA_XHIGH"} {
		t.Run(model, func(t *testing.T) {
			prices, err := svc.GetModelPricing(model)
			require.NoError(t, err)
			require.InDelta(t, 0.1e-6, prices.InputPricePerToken, 1e-12)
			require.InDelta(t, 0.5e-6, prices.OutputPricePerToken, 1e-12)
			require.InDelta(t, 0.125e-6, prices.CacheCreationPricePerToken, 1e-12)
			require.InDelta(t, 0.01e-6, prices.CacheReadPricePerToken, 1e-12)
			require.InDelta(t, 0.2e-6, prices.InputPricePerTokenPriority, 1e-12)
			require.Equal(t, 272000, prices.LongContextInputThreshold)
		})
	}
}

func TestNewModelPricingChannelOverridesAndFamilyIsolation(t *testing.T) {
	svc := NewBillingService(&config.Config{}, nil)
	for _, model := range []string{"gpt-6-sol", "gpt-6-luna", "claude-opus-5-5"} {
		t.Run(model, func(t *testing.T) {
			zero := 0.0
			prices, err := svc.GetModelPricingWithChannel(model, &ChannelModelPricing{
				InputPrice: &zero, OutputPrice: &zero, CacheWritePrice: &zero, CacheReadPrice: &zero})
			require.NoError(t, err)
			cost := svc.computeTokenBreakdown(prices, UsageTokens{InputTokens: 300000, CacheReadTokens: 1000,
				CacheCreationTokens: 1000, OutputTokens: 1000}, 1, "priority", true)
			require.Zero(t, cost.TotalCost)
			fresh, err := svc.GetModelPricing(model)
			require.NoError(t, err)
			require.Positive(t, fresh.InputPricePerToken)
		})
	}
	for _, tc := range []struct {
		model string
		input float64
	}{
		{"claude-opus-5", 5e-6}, {"gpt-6", 10e-6},
		{"claude-opus-5-5-thinking", 4e-6}, {"openai/gpt-6-sol-max", 2e-6},
		{"gpt-6-luna-openai-compact", 0.1e-6},
	} {
		prices, err := svc.GetModelPricing(tc.model)
		require.NoError(t, err)
		require.InDelta(t, tc.input, prices.InputPricePerToken, 1e-12, tc.model)
	}
	for _, name := range []string{"gpt-6-solitude", "gpt-6-luna-preview", "gpt-6-omni"} {
		_, err := svc.GetModelPricing(name)
		require.ErrorIs(t, err, ErrModelPricingUnavailable)
	}
}

func TestNewModelPricingExplicitZeroCacheWrite(t *testing.T) {
	parsed := &PricingService{}
	var err error
	parsed.pricingData, err = parsed.parsePricingData([]byte(`{
		"gpt-6-sol":{"input_cost_per_token":0.000002,"output_cost_per_token":0.00001,
			"input_cost_per_token_priority":0.000004,"output_cost_per_token_priority":0.00002,
			"cache_creation_input_token_cost":0},
		"gpt-6-luna":{"input_cost_per_token":0.0000001,"output_cost_per_token":0.0000005,
			"input_cost_per_token_priority":0.0000002,"output_cost_per_token_priority":0.000001,
			"cache_creation_input_token_cost":0.000000125,"cache_creation_input_token_cost_priority":0},
		"claude-opus-5-5":{"input_cost_per_token":0.000004,"output_cost_per_token":0.00002,
			"cache_creation_input_token_cost":0,"cache_creation_input_token_cost_above_1hr":0}
	}`))
	require.NoError(t, err)
	billing := NewBillingService(&config.Config{}, parsed)
	for _, tc := range []struct {
		model, tier string
		want        float64
	}{
		{"gpt-6-sol", "", 0}, {"gpt-6-sol", "priority", 0},
		{"gpt-6-luna", "", 1000 * 0.125e-6}, {"gpt-6-luna", "priority", 0},
		{"claude-opus-5-5", "", 0}, {"claude-opus-5-5", "fast", 0},
	} {
		cost, err := billing.CalculateCostWithServiceTier(tc.model, UsageTokens{CacheCreationTokens: 1000}, 1, tc.tier)
		require.NoError(t, err)
		require.InDelta(t, tc.want, cost.CacheCreationCost, 1e-12, tc.model+"/"+tc.tier)
	}
	// Explicit free 1h writes must not be replaced with the ordinary 5m rate.
	parsed.pricingData["claude-opus-5-5"].CacheCreationInputTokenCost = 5e-6
	cost, err := billing.CalculateCost("claude-opus-5-5", UsageTokens{
		CacheCreationTokens: 1000, CacheCreation5mTokens: 400, CacheCreation1hTokens: 600}, 1)
	require.NoError(t, err)
	require.InDelta(t, 400*5e-6, cost.CacheCreationCost, 1e-12)

	cost, err = billing.CalculateCostWithServiceTier("gpt-6-luna", UsageTokens{
		InputTokens: 100, OutputTokens: 10, CacheCreationTokens: 1000}, 1, "priority")
	require.NoError(t, err)
	require.InDelta(t, 100*0.2e-6, cost.InputCost, 1e-12)
	require.InDelta(t, 10*1e-6, cost.OutputCost, 1e-12)
	require.Zero(t, cost.CacheCreationCost)
}

func TestNewModelPricingAliasesRetainExplicitOverrides(t *testing.T) {
	exact := &LiteLLMModelPricing{InputCostPerToken: 9e-6, OutputCostPerToken: 19e-6}
	canonical := &LiteLLMModelPricing{InputCostPerToken: 3e-6, OutputCostPerToken: 13e-6}
	svc := &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"gpt-6-sol": canonical, "gpt-6-sol-max": exact,
		"gpt-6-luna": canonical, "claude-opus-5-5": canonical,
		"claude-opus-5": {InputCostPerToken: 5e-6},
		"gpt-6-astra":   {InputCostPerToken: 10e-6},
	}}
	require.Same(t, exact, svc.GetModelPricing("gpt-6-sol-max"))
	require.Same(t, exact, svc.GetModelPricing("openai/gpt-6-sol-max"))
	for _, model := range []string{"openai/gpt-6-luna-openai-compact", "claude-opus-5-5-thinking"} {
		require.Same(t, canonical, svc.GetModelPricing(model), model)
	}
	pricing := NewBillingService(&config.Config{}, svc)
	got, err := pricing.GetModelPricing("gpt-6-sol-max")
	require.NoError(t, err)
	require.InDelta(t, 9e-6, got.InputPricePerToken, 1e-12)
	require.InDelta(t, 11.25e-6, got.CacheCreationPricePerToken, 1e-12)
	require.Equal(t, 272000, got.LongContextInputThreshold)
	require.Same(t, svc.pricingData["claude-opus-5"], svc.GetModelPricing("claude-opus-5"))
}

func TestNewModelPricingPartialPriorityUsesWholeTierMultiplier(t *testing.T) {
	svc := NewBillingService(&config.Config{}, &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"gpt-6-sol": {InputCostPerToken: 3e-6, OutputCostPerToken: 12e-6,
			InputCostPerTokenPriority: 10e-6, CacheCreationInputTokenCost: 3.75e-6,
			CacheReadInputTokenCost: 0.3e-6},
	}})
	tokens := UsageTokens{InputTokens: 100, OutputTokens: 20, CacheReadTokens: 10, CacheCreationTokens: 5}
	normal, err := svc.CalculateCost("gpt-6-sol", tokens, 1)
	require.NoError(t, err)
	priority, err := svc.CalculateCostWithServiceTier("gpt-6-sol", tokens, 1, "priority")
	require.NoError(t, err)
	require.InDelta(t, normal.TotalCost*2, priority.TotalCost, 1e-12)
	require.InDelta(t, normal.InputCost*2, priority.InputCost, 1e-12)
}
