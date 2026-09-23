//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func newGroupIDOnlyPricingResolver(t *testing.T, model string, inputPrice, outputPrice float64) *ModelPricingResolver {
	t.Helper()

	const groupID = int64(100)
	channelPricing := &ChannelModelPricing{
		Platform:    PlatformOpenAI,
		Models:      []string{model},
		BillingMode: BillingModeToken,
		InputPrice:  &inputPrice,
		OutputPrice: &outputPrice,
	}
	cache := newEmptyChannelCache()
	cache.loadedAt = time.Now()
	cache.channelByGroupID[groupID] = &Channel{
		ID:     1,
		Status: StatusActive,
	}
	cache.groupPlatform[groupID] = PlatformOpenAI
	cache.pricingByGroupModel[channelModelKey{
		groupID:  groupID,
		platform: PlatformOpenAI,
		model:    normalizeChannelPricingModelName(model),
	}] = channelPricing

	channelService := &ChannelService{}
	channelService.cache.Store(cache)
	return NewModelPricingResolver(channelService, NewBillingService(&config.Config{}, nil))
}

func TestGatewayTokenBillingUsesChannelPricingWhenOnlyGroupIDIsHydrated(t *testing.T) {
	inputPrice := 11e-6
	outputPrice := 22e-6
	resolver := newGroupIDOnlyPricingResolver(t, "gpt-6-sol", inputPrice, outputPrice)
	svc := &GatewayService{
		billingService: NewBillingService(&config.Config{}, nil),
		resolver:       resolver,
	}
	groupID := int64(100)

	cost := svc.calculateTokenCost(
		context.Background(),
		&ForwardResult{
			Model: "gpt-6-sol",
			Usage: ClaudeUsage{
				InputTokens:  1_000,
				OutputTokens: 200,
			},
		},
		&APIKey{GroupID: &groupID},
		"gpt-6-sol",
		1,
		1,
		nil,
	)

	require.InDelta(t, 1_000*inputPrice, cost.InputCost, 1e-12)
	require.InDelta(t, 200*outputPrice, cost.OutputCost, 1e-12)
	require.InDelta(t, 1_000*inputPrice+200*outputPrice, cost.TotalCost, 1e-12)
}

func TestOpenAITokenBillingUsesChannelPricingWhenOnlyGroupIDIsHydrated(t *testing.T) {
	inputPrice := 11e-6
	outputPrice := 22e-6
	resolver := newGroupIDOnlyPricingResolver(t, "gpt-6-sol", inputPrice, outputPrice)
	svc := &OpenAIGatewayService{
		billingService: NewBillingService(&config.Config{}, nil),
		resolver:       resolver,
	}
	groupID := int64(100)

	cost, err := svc.calculateOpenAIRecordUsageTokenCost(
		context.Background(),
		&APIKey{GroupID: &groupID},
		&Account{Platform: PlatformOpenAI},
		"gpt-6-sol",
		1,
		1,
		UsageTokens{InputTokens: 1_000, OutputTokens: 200},
		"",
		"",
		0,
	)

	require.NoError(t, err)
	require.InDelta(t, 1_000*inputPrice, cost.InputCost, 1e-12)
	require.InDelta(t, 200*outputPrice, cost.OutputCost, 1e-12)
	require.InDelta(t, 1_000*inputPrice+200*outputPrice, cost.TotalCost, 1e-12)
}
