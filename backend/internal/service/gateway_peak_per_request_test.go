package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type peakPerRequestUsageRepo struct {
	UsageLogRepository
	lastLog *UsageLog
}

func (r *peakPerRequestUsageRepo) Create(_ context.Context, log *UsageLog) (bool, error) {
	r.lastLog = log
	return true, nil
}

type peakPerRequestUserRepo struct {
	UserRepository
	lastAmount float64
}

func (r *peakPerRequestUserRepo) DeductBalance(_ context.Context, _ int64, amount float64) error {
	r.lastAmount = amount
	return nil
}

func TestGatewayPeakRateDoesNotAffectChannelPerRequestCost(t *testing.T) {
	setTestTimezone(t, "UTC")
	groupID := int64(903)
	price := 0.2
	cache := newEmptyChannelCache()
	cache.pricingByGroupModel[channelModelKey{groupID: groupID, model: "request-priced-model"}] = &ChannelModelPricing{
		BillingMode:     BillingModePerRequest,
		PerRequestPrice: &price,
	}
	cache.channelByGroupID[groupID] = &Channel{ID: groupID, Status: StatusActive}
	cache.groupPlatform[groupID] = ""
	cache.loadedAt = time.Now()
	channelService := &ChannelService{}
	channelService.cache.Store(cache)
	cfg := &config.Config{}
	usageRepo := &peakPerRequestUsageRepo{}
	userRepo := &peakPerRequestUserRepo{}
	billingService := NewBillingService(cfg, nil)
	svc := NewGatewayService(
		nil, nil, usageRepo, nil, userRepo, nil, nil, nil, cfg, nil, nil,
		billingService, nil, &BillingCacheService{}, nil, nil, &DeferredService{},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	svc.resolver = NewModelPricingResolver(channelService, billingService)
	apiKey := &APIKey{
		ID:      801,
		GroupID: i64p(groupID),
		Group: &Group{
			ID:                 groupID,
			RateMultiplier:     1.5,
			SubscriptionType:   SubscriptionTypeStandard,
			PeakRateEnabled:    true,
			PeakStart:          "14:00",
			PeakEnd:            "18:00",
			PeakRateMultiplier: 0.7,
		},
	}

	err := svc.RecordUsage(
		context.Background(),
		&RecordUsageInput{
			Result: &ForwardResult{
				RequestID: "gateway_peak_per_request",
				Model:     "request-priced-model",
				Duration:  time.Second,
			},
			APIKey:           apiKey,
			User:             &User{ID: 601},
			Account:          &Account{ID: 701},
			RequestStartedAt: at(17, 59),
		},
	)

	require.NoError(t, err)
	require.NotNil(t, usageRepo.lastLog)
	require.NotNil(t, usageRepo.lastLog.BillingMode)
	require.Equal(t, string(BillingModePerRequest), *usageRepo.lastLog.BillingMode)
	require.InDelta(t, 1.5, usageRepo.lastLog.RateMultiplier, 1e-12)
	require.InDelta(t, 0.2, usageRepo.lastLog.TotalCost, 1e-12)
	require.InDelta(t, 0.3, usageRepo.lastLog.ActualCost, 1e-12)
	require.InDelta(t, 0.3, userRepo.lastAmount, 1e-12)
}
