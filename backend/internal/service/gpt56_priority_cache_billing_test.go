package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestGPT56PriorityCacheWriteTierAndLongContextMatrix(t *testing.T) {
	billing := NewBillingService(&config.Config{}, nil)
	models := []struct {
		name     string
		standard float64
		priority float64
	}{
		{name: "gpt-5.6-sol", standard: 6.25e-6, priority: 12.5e-6},
		{name: "gpt-5.6-terra", standard: 3.125e-6, priority: 6.25e-6},
		{name: "gpt-5.6-luna", standard: 1.25e-6, priority: 2.5e-6},
	}
	tiers := []struct {
		name              string
		serviceTier       string
		basePriceSelector func(standard, priority float64) float64
		tierMultiplier    float64
	}{
		{name: "standard", basePriceSelector: func(standard, _ float64) float64 { return standard }, tierMultiplier: 1},
		{name: "priority", serviceTier: "priority", basePriceSelector: func(_, priority float64) float64 { return priority }, tierMultiplier: 1},
		{name: "flex", serviceTier: "flex", basePriceSelector: func(standard, _ float64) float64 { return standard }, tierMultiplier: 0.5},
	}

	for _, model := range models {
		for _, tier := range tiers {
			t.Run(model.name+"/"+tier.name, func(t *testing.T) {
				basePrice := tier.basePriceSelector(model.standard, model.priority)
				atBoundary := UsageTokens{InputTokens: 271800, CacheCreationTokens: 100, CacheReadTokens: 100}
				cost, err := billing.CalculateCostWithServiceTier(model.name, atBoundary, 1, tier.serviceTier)
				require.NoError(t, err)
				require.InDelta(t, 100*basePrice*tier.tierMultiplier, cost.CacheCreationCost, 1e-12)

				overBoundary := UsageTokens{InputTokens: 271801, CacheCreationTokens: 100, CacheReadTokens: 100}
				cost, err = billing.CalculateCostWithServiceTier(model.name, overBoundary, 1, tier.serviceTier)
				require.NoError(t, err)
				require.InDelta(t, 100*basePrice*2*tier.tierMultiplier, cost.CacheCreationCost, 1e-12)
			})
		}
	}
}

func TestGPT56PriorityCacheWriteChannelAndIntervalOverrides(t *testing.T) {
	billing := NewBillingService(&config.Config{}, nil)
	resolver := NewModelPricingResolver(nil, billing)

	calculate := func(t *testing.T, resolved *ResolvedPricing, serviceTier string) *CostBreakdown {
		t.Helper()
		cost, err := billing.CalculateCostUnified(CostInput{
			Ctx:            context.Background(),
			Model:          "gpt-5.6-sol",
			Tokens:         UsageTokens{CacheCreationTokens: 10},
			RateMultiplier: 1,
			ServiceTier:    serviceTier,
			Resolver:       resolver,
			Resolved:       resolved,
		})
		require.NoError(t, err)
		return cost
	}

	for _, override := range []float64{7e-6, 0} {
		t.Run("flat", func(t *testing.T) {
			resolved := &ResolvedPricing{Mode: BillingModeToken, BasePricing: newOpenAIGPT56FallbackModelPricing(5e-6, 30e-6)}
			resolver.applyTokenOverrides(&ChannelModelPricing{CacheWritePrice: &override}, resolved)
			require.True(t, resolved.BasePricing.CacheCreationPriceExplicit)
			for _, tier := range []string{"", "priority"} {
				require.InDelta(t, 10*override, calculate(t, resolved, tier).CacheCreationCost, 1e-12)
			}
		})

		t.Run("interval", func(t *testing.T) {
			resolved := &ResolvedPricing{Mode: BillingModeToken, BasePricing: newOpenAIGPT56FallbackModelPricing(5e-6, 30e-6)}
			resolver.applyTokenOverrides(&ChannelModelPricing{Intervals: []PricingInterval{{CacheWritePrice: &override}}}, resolved)
			require.Len(t, resolved.Intervals, 1)
			for _, tier := range []string{"", "priority"} {
				require.InDelta(t, 10*override, calculate(t, resolved, tier).CacheCreationCost, 1e-12)
			}
		})
	}

	t.Run("derive missing standard and priority prices", func(t *testing.T) {
		pricing := billing.applyModelSpecificPricingPolicy("gpt-5.6-sol", &ModelPricing{
			InputPricePerToken:         4e-6,
			InputPricePerTokenPriority: 8e-6,
		})
		require.InDelta(t, 5e-6, pricing.CacheCreationPricePerToken, 1e-12)
		require.InDelta(t, 10e-6, pricing.CacheCreationPricePerTokenPriority, 1e-12)
	})
}

func TestGPT56PriorityCacheWritePreservesCacheBreakdown(t *testing.T) {
	billing := NewBillingService(&config.Config{}, nil)
	pricing := &ModelPricing{
		CacheCreationPricePerToken:         5e-6,
		CacheCreationPricePerTokenPriority: 10e-6,
		CacheCreation5mPrice:               2e-6,
		CacheCreation1hPrice:               3e-6,
		SupportsCacheBreakdown:             true,
	}
	tokens := UsageTokens{
		CacheCreationTokens:   10,
		CacheCreation5mTokens: 4,
		CacheCreation1hTokens: 6,
	}

	cost := billing.computeTokenBreakdown(pricing, tokens, 1, "priority", false)
	require.InDelta(t, 4*2e-6+6*3e-6, cost.CacheCreationCost, 1e-12)
	require.NotEqual(t, 10*10e-6, cost.CacheCreationCost)
}
