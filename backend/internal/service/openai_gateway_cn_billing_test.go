package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestFilterCNProviderBillingModelCandidates(t *testing.T) {
	ctx := context.Background()
	groupID := int64(731)
	cnAccount := &Account{ID: 1, Platform: PlatformKimi}
	apiKey := &APIKey{GroupID: &groupID, Group: &Group{ID: groupID, Platform: PlatformKimi}}
	candidates := []string{"kimi-k2-0905-preview", "claude-sonnet-4", "moonshot-v1-8k"}

	t.Run("filters unpriced Claude candidates", func(t *testing.T) {
		svc := &OpenAIGatewayService{}
		require.Equal(t, []string{"kimi-k2-0905-preview", "moonshot-v1-8k"},
			svc.filterCNProviderBillingModelCandidates(ctx, cnAccount, apiKey, candidates))
		require.Empty(t, svc.filterCNProviderBillingModelCandidates(ctx, cnAccount, apiKey, []string{"claude-sonnet-4"}))
	})

	t.Run("keeps explicitly priced group candidate", func(t *testing.T) {
		groupPricedKey := &APIKey{GroupID: &groupID, Group: &Group{
			ID:       groupID,
			Platform: PlatformKimi,
			ModelPricing: []ChannelModelPricing{{
				Models:     []string{"claude-*"},
				InputPrice: s229BFloat64Ptr(1.5e-6),
			}},
		}}
		svc := &OpenAIGatewayService{resolver: NewModelPricingResolver(nil, newS229BBillingService())}
		require.Equal(t, []string{"claude-sonnet-4"},
			svc.filterCNProviderBillingModelCandidates(ctx, cnAccount, groupPricedKey, []string{"claude-sonnet-4"}))
	})

	t.Run("keeps explicitly priced channel candidate", func(t *testing.T) {
		channelService := newS229BChannelServiceWithCache(&channelCache{
			pricingByGroupModel: map[channelModelKey]*ChannelModelPricing{
				{groupID: groupID, platform: PlatformKimi, model: "claude-sonnet-4"}: {
					Models:     []string{"claude-sonnet-4"},
					InputPrice: s229BFloat64Ptr(1.5e-6),
				},
			},
			wildcardByGroupPlatform: map[channelGroupPlatformKey][]*wildcardPricingEntry{},
			mappingByGroupModel:     map[channelModelKey]string{},
			wildcardMappingByGP:     map[channelGroupPlatformKey][]*wildcardMappingEntry{},
			channelByGroupID:        map[int64]*Channel{groupID: {ID: 1, Status: StatusActive}},
			groupPlatform:           map[int64]string{groupID: PlatformKimi},
			byID:                    map[int64]*Channel{},
		})
		svc := &OpenAIGatewayService{resolver: NewModelPricingResolver(channelService, newS229BBillingService())}
		require.Equal(t, []string{"claude-sonnet-4"},
			svc.filterCNProviderBillingModelCandidates(ctx, cnAccount, apiKey, []string{"claude-sonnet-4"}))
	})

	t.Run("non CN candidates pass through unchanged", func(t *testing.T) {
		svc := &OpenAIGatewayService{}
		for _, platform := range []string{PlatformOpenAI, PlatformGrok, PlatformAnthropic} {
			require.Equal(t, candidates, svc.filterCNProviderBillingModelCandidates(ctx,
				&Account{ID: 2, Platform: platform}, apiKey, candidates), platform)
		}
	})
}

func newS229BBillingService() *BillingService {
	return NewBillingService(&config.Config{}, nil)
}

func newS229BChannelServiceWithCache(cache *channelCache) *ChannelService {
	cache.loadedAt = time.Now()
	service := &ChannelService{}
	service.cache.Store(cache)
	return service
}

func s229BFloat64Ptr(value float64) *float64 {
	return &value
}

func TestCalculateOpenAIRecordUsageCost_EmptyCandidatesIsPricingUnavailable(t *testing.T) {
	svc := &OpenAIGatewayService{}
	apiKey := &APIKey{Group: &Group{ID: 1, Platform: PlatformKimi}}

	for _, candidates := range [][]string{nil, {"  "}} {
		_, err := svc.calculateOpenAIRecordUsageCost(
			context.Background(), nil, apiKey, nil, candidates,
			1, 1, 1, UsageTokens{InputTokens: 100}, "", "", 0,
		)
		require.Error(t, err)
		require.True(t, isUsagePricingUnavailableError(err), err)
	}
}

func TestOpenAIGatewayServiceRecordUsage_CNFilteredCandidatesWriteZeroCostLog(t *testing.T) {
	groupID := int64(732)
	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	userRepo := &openAIRecordUsageUserRepoStub{}
	subRepo := &openAIRecordUsageSubRepoStub{}
	svc := newOpenAIRecordUsageServiceForTest(usageRepo, userRepo, subRepo, nil)
	svc.cfg.RunMode = config.RunModeSimple

	err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{
			RequestID: "cn_claude_unpriced",
			Model:     "claude-sonnet-4",
			Usage:     OpenAIUsage{InputTokens: 20, OutputTokens: 10},
			Duration:  time.Second,
		},
		APIKey: &APIKey{
			ID:      1001,
			GroupID: &groupID,
			Group:   &Group{ID: groupID, Platform: PlatformKimi, RateMultiplier: 1},
		},
		User:    &User{ID: 2001},
		Account: &Account{ID: 3001, Platform: PlatformKimi},
	})

	require.NoError(t, err)
	require.Equal(t, 1, usageRepo.calls)
	require.NotNil(t, usageRepo.lastLog)
	require.Zero(t, usageRepo.lastLog.TotalCost)
	require.Zero(t, usageRepo.lastLog.ActualCost)
	require.Equal(t, 0, userRepo.deductCalls)
	require.Equal(t, 0, subRepo.incrementCalls)
}
