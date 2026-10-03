//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestTokenCostInput_GroupIDPrecedenceAndFallback(t *testing.T) {
	billing := NewBillingService(&config.Config{}, nil)
	group := &Group{ID: 200}
	explicitGroupID := int64(100)

	input := billing.tokenCostInput(TokenCostRequest{
		GroupID: &explicitGroupID,
		Group:   group,
	})
	require.NotNil(t, input.GroupID)
	require.Equal(t, int64(100), *input.GroupID)

	input = billing.tokenCostInput(TokenCostRequest{Group: group})
	require.NotNil(t, input.GroupID)
	require.Equal(t, int64(200), *input.GroupID)

	input = billing.tokenCostInput(TokenCostRequest{})
	require.Nil(t, input.GroupID)
}

func TestCalculateTokenCostForRequest_GroupIDOnlyUsesChannelPricing(t *testing.T) {
	inputPrice := 11e-6
	outputPrice := 22e-6
	resolver := newGroupIDOnlyPricingResolver(t, "gpt-6-sol", inputPrice, outputPrice)
	billing := NewBillingService(&config.Config{}, nil)
	groupID := int64(100)

	got, err := billing.CalculateTokenCostForRequest(TokenCostRequest{
		Ctx:            context.Background(),
		Model:          "gpt-6-sol",
		GroupID:        &groupID,
		Tokens:         UsageTokens{InputTokens: 1_000, OutputTokens: 200},
		RateMultiplier: 1,
		Resolver:       resolver,
	})

	require.NoError(t, err)
	require.InDelta(t, 1_000*inputPrice, got.InputCost, 1e-12)
	require.InDelta(t, 200*outputPrice, got.OutputCost, 1e-12)
	require.InDelta(t, 1_000*inputPrice+200*outputPrice, got.TotalCost, 1e-12)
}

func TestCalculateTokenCostForRequest_AbsentGroupIDFallsBackToCatalog(t *testing.T) {
	billing := NewBillingService(&config.Config{}, nil)
	resolver := NewModelPricingResolver(nil, billing)
	tokens := UsageTokens{InputTokens: 1_000, OutputTokens: 10}

	got, err := billing.CalculateTokenCostForRequest(TokenCostRequest{
		Model:          "gpt-5.4",
		Tokens:         tokens,
		RateMultiplier: 1,
		Resolver:       resolver,
	})
	require.NoError(t, err)

	want, err := billing.CalculateCost("gpt-5.4", tokens, 1)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
