package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestOpenAIGatewayServiceEstimateOpenAIImagesCost_UsesVoucherPreflightAmount(t *testing.T) {
	groupID := int64(9)
	channelService := NewChannelService(nil, nil, nil, nil)
	cache := newEmptyChannelCache()
	cache.loadedAt = time.Now()
	cache.channelByGroupID[groupID] = &Channel{ID: 1, Status: StatusActive}
	cache.groupPlatform[groupID] = PlatformOpenAI
	cache.pricingByGroupModel[channelModelKey{groupID: groupID, platform: PlatformOpenAI, model: "gpt-image-2"}] = &ChannelModelPricing{
		ID:              1,
		Platform:        PlatformOpenAI,
		Models:          []string{"gpt-image-2"},
		BillingMode:     BillingModeImage,
		PerRequestPrice: floatPtrForOpenAIImagesEstimate(1),
	}
	channelService.cache.Store(cache)
	billingService := NewBillingService(&config.Config{}, nil)
	svc := NewOpenAIGatewayService(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
		nil,
		nil,
		billingService,
		nil,
		nil,
		nil,
		&DeferredService{},
		nil,
		nil,
		NewModelPricingResolver(channelService, billingService),
		channelService,
		nil,
		nil,
		nil,
	)

	cost, billingModel, err := svc.EstimateOpenAIImagesCost(
		context.Background(),
		&APIKey{ID: 1, UserID: 7, GroupID: &groupID, Group: &Group{ID: groupID, RateMultiplier: 1, ImageRateMultiplier: 1}},
		&User{ID: 7},
		&Account{ID: 11, Type: AccountTypeAPIKey},
		&OpenAIImagesRequest{Model: "gpt-image-2", N: 3, SizeTier: ImageBillingSize1K, Size: "1K"},
		ChannelUsageFields{
			OriginalModel:      "gpt-image-2",
			ChannelMappedModel: "gpt-image-2",
			BillingModelSource: BillingModelSourceChannelMapped,
		},
	)

	require.NoError(t, err)
	require.Equal(t, "gpt-image-2", billingModel)
	require.NotNil(t, cost)
	require.InDelta(t, 3, cost.ActualCost, 1e-12)
}

func floatPtrForOpenAIImagesEstimate(v float64) *float64 {
	return &v
}
