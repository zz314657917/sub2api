package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func newS220PricingBillingService() *BillingService {
	return NewBillingService(&config.Config{}, nil)
}

func s220FloatPtr(v float64) *float64 {
	return &v
}

func TestCalculateCostUnified_GroupLongContextToggleUsesPresetLadder(t *testing.T) {
	svc := newS220PricingBillingService()
	resolver := NewModelPricingResolver(nil, svc)
	tokens := UsageTokens{InputTokens: 250000, OutputTokens: 1000}

	disabled, err := svc.CalculateCostUnified(CostInput{
		Model: "grok-4.5", Group: &Group{LongContextPricingEnabled: false}, Tokens: tokens, RateMultiplier: 1, Resolver: resolver,
	})
	require.NoError(t, err)
	enabled, err := svc.CalculateCostUnified(CostInput{
		Model: "grok-4.5", Group: &Group{LongContextPricingEnabled: true}, Tokens: tokens, RateMultiplier: 1, Resolver: resolver,
	})
	require.NoError(t, err)
	require.False(t, disabled.LongContextBillingApplied)
	require.True(t, enabled.LongContextBillingApplied)
	require.InDelta(t, disabled.InputCost*2, enabled.InputCost, 1e-12)
	require.InDelta(t, disabled.OutputCost*2, enabled.OutputCost, 1e-12)
}

func TestResolve_GroupPricingOverridesChannel(t *testing.T) {
	svc := newS220PricingBillingService()
	resolver := NewModelPricingResolver(nil, svc)
	inputPrice, outputPrice := 1e-6, 2e-6
	resolved := resolver.Resolve(context.Background(), PricingInput{
		Model: "claude-sonnet-4",
		Group: &Group{ModelPricing: []ChannelModelPricing{{
			Models: []string{"claude-sonnet-*"}, BillingMode: BillingModeToken,
			InputPrice: &inputPrice, OutputPrice: &outputPrice,
		}}},
	})
	require.Equal(t, PricingSourceGroup, resolved.Source)
	require.InDelta(t, inputPrice, resolved.BasePricing.InputPricePerToken, 1e-12)
	require.InDelta(t, outputPrice, resolved.BasePricing.OutputPricePerToken, 1e-12)
}

func TestResolve_GroupLongContextUsesPresetNotCustomIntervals(t *testing.T) {
	svc := newS220PricingBillingService()
	resolver := NewModelPricingResolver(nil, svc)
	price := 1e-6
	max := 200000
	group := &Group{LongContextPricingEnabled: true, ModelPricing: []ChannelModelPricing{{
		Models: []string{"grok-4.5"}, BillingMode: BillingModeToken, InputPrice: &price,
		Intervals: []PricingInterval{{MinTokens: 0, MaxTokens: &max, InputPrice: s220FloatPtr(9e-6)}},
	}}}
	resolved := resolver.Resolve(context.Background(), PricingInput{Model: "grok-4.5", Group: group})
	require.Empty(t, resolved.Intervals)
	require.InDelta(t, price, resolver.GetIntervalPricing(resolved, 250000).InputPricePerToken, 1e-12)
	require.Equal(t, 200000, resolved.BasePricing.LongContextInputThreshold)
}
