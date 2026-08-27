package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLegacyLongContextRule_OnlyGemini(t *testing.T) {
	billing := NewBillingService(&config.Config{}, nil)
	rule := billing.LegacyLongContextRule(PlatformGemini)
	require.NotNil(t, rule)
	require.Equal(t, 200000, rule.Threshold)
	require.InDelta(t, 2.0, rule.Multiplier, 1e-12)
	require.Nil(t, billing.LegacyLongContextRule(PlatformOpenAI))
}

func TestCalculateTokenCostForRequest_ExplicitPricingWinsOverLegacyRule(t *testing.T) {
	billing := NewBillingService(&config.Config{}, nil)
	resolver := NewModelPricingResolver(nil, billing)
	group := &Group{ID: 100, Platform: PlatformGemini, LongContextPricingEnabled: true}
	resolved := &ResolvedPricing{
		Mode:   BillingModeToken,
		Source: PricingSourceChannel,
		BasePricing: &ModelPricing{
			InputPricePerToken:  10e-6,
			OutputPricePerToken: 40e-6,
		},
	}

	cost, err := billing.CalculateTokenCostForRequest(TokenCostRequest{
		Ctx:               context.Background(),
		Model:             "gemini-2.5-pro",
		Group:             group,
		Tokens:            UsageTokens{InputTokens: 300000, OutputTokens: 1000},
		RateMultiplier:    1,
		Resolver:          resolver,
		Resolved:          resolved,
		LegacyLongContext: billing.LegacyLongContextRule(PlatformGemini),
	})

	require.NoError(t, err)
	require.InDelta(t, 3.0, cost.InputCost, 1e-12)
	require.False(t, cost.LongContextBillingApplied)
}

func TestCalculateTokenCostForRequest_LegacyRuleFollowsGroupToggle(t *testing.T) {
	billing := NewBillingService(&config.Config{}, nil)
	resolver := NewModelPricingResolver(nil, billing)
	tokens := UsageTokens{InputTokens: 300000, OutputTokens: 1000}
	rule := billing.LegacyLongContextRule(PlatformGemini)

	for _, enabled := range []bool{true, false} {
		group := &Group{ID: 100, Platform: PlatformGemini, LongContextPricingEnabled: enabled}
		groupID := group.ID
		resolved := resolver.Resolve(context.Background(), PricingInput{Model: "gpt-5.4", GroupID: &groupID, Group: group})

		got, err := billing.CalculateTokenCostForRequest(TokenCostRequest{
			Ctx:               context.Background(),
			Model:             "gpt-5.4",
			Group:             group,
			Tokens:            tokens,
			RateMultiplier:    1,
			Resolver:          resolver,
			Resolved:          resolved,
			LegacyLongContext: rule,
		})
		require.NoError(t, err)

		if enabled {
			want, wantErr := billing.CalculateCostWithLongContext("gpt-5.4", tokens, 1, rule.Threshold, rule.Multiplier)
			require.NoError(t, wantErr)
			require.Equal(t, want, got)
			continue
		}

		want, wantErr := billing.CalculateCostUnified(CostInput{
			Ctx:            context.Background(),
			Model:          "gpt-5.4",
			GroupID:        &groupID,
			Group:          group,
			Tokens:         tokens,
			RequestCount:   1,
			RateMultiplier: 1,
			Resolver:       resolver,
			Resolved:       resolved,
		})
		require.NoError(t, wantErr)
		require.Equal(t, want, got)
		require.False(t, got.LongContextBillingApplied)
	}
}

func TestCalculateTokenCostForRequest_NoResolverFallsBackToCatalog(t *testing.T) {
	billing := NewBillingService(&config.Config{}, nil)
	tokens := UsageTokens{InputTokens: 1000, OutputTokens: 10}

	got, err := billing.CalculateTokenCostForRequest(TokenCostRequest{
		Model:          "gpt-5.4",
		Tokens:         tokens,
		RateMultiplier: 1,
	})
	require.NoError(t, err)
	want, err := billing.CalculateCost("gpt-5.4", tokens, 1)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
