package service

// issue #5256 回归测试：带后缀模型在渠道只配置基名时，仍使用渠道自定义价格。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	channelPricingExpectedChannelCost  = 0.4
	channelPricingExpectedOfficialCost = 0.2
	channelPricingUnrelatedCost        = 0.9
)

// tokenPricingForModels 构造 token 计费模式的渠道定价；inputPerMillion 单位为 USD/1M token。
func tokenPricingForModels(models []string, inputPerMillion float64) ChannelModelPricing {
	return ChannelModelPricing{
		Platform:        PlatformOpenAI,
		Models:          models,
		BillingMode:     BillingModeToken,
		InputPrice:      float64Ptr(inputPerMillion / 1e6),
		OutputPrice:     float64Ptr(2.4e-6),
		CacheWritePrice: float64Ptr(0.5e-6),
		CacheReadPrice:  float64Ptr(0.04e-6),
	}
}

func newChannelServiceWithPricings(groupID int64, pricings []ChannelModelPricing) *ChannelService {
	ch := Channel{
		ID:           1,
		Name:         "codex-channel",
		Status:       StatusActive,
		ModelPricing: pricings,
		GroupIDs:     []int64{groupID},
	}
	cs := &ChannelService{}
	cs.cache.Store(populateChannelCache([]Channel{ch}, map[int64]string{groupID: PlatformOpenAI}))
	return cs
}

// recordUsageWithChannelPricing 用给定渠道定价跑一次 RecordUsage，返回落库的 UsageLog。
func recordUsageWithChannelPricing(t *testing.T, requestedModel string, subscriptionGroup bool, pricings []ChannelModelPricing) *UsageLog {
	t.Helper()
	const groupID = int64(777)

	usageRepo := &openAIRecordUsageLogRepoStub{inserted: true}
	svc := newOpenAIRecordUsageServiceForTest(usageRepo, &openAIRecordUsageUserRepoStub{}, &openAIRecordUsageSubRepoStub{}, nil)
	cs := newChannelServiceWithPricings(groupID, pricings)
	svc.channelService = cs
	svc.resolver = NewModelPricingResolver(cs, svc.billingService)

	group := &Group{ID: groupID, Platform: PlatformOpenAI, RateMultiplier: 1}
	if subscriptionGroup {
		group.SubscriptionType = SubscriptionTypeSubscription
	}

	err := svc.RecordUsage(context.Background(), &OpenAIRecordUsageInput{
		Result: &OpenAIForwardResult{
			RequestID:    "resp_luna_5256",
			Model:        requestedModel,
			BillingModel: requestedModel,
			Usage: OpenAIUsage{
				InputTokens:  1_000_000,
				OutputTokens: 0,
			},
			Duration: time.Second,
		},
		ChannelUsageFields: ChannelUsageFields{
			OriginalModel:      requestedModel,
			ChannelMappedModel: requestedModel,
		},
		APIKey:  &APIKey{ID: 1, GroupID: i64p(groupID), Group: group},
		User:    &User{ID: 1},
		Account: &Account{ID: 1, Platform: PlatformOpenAI},
	})
	require.NoError(t, err)
	require.NotNil(t, usageRepo.lastLog)
	return usageRepo.lastLog
}

func TestChannelPricing_ExactModelMatch(t *testing.T) {
	log := recordUsageWithChannelPricing(t, "gpt-5.6-luna", false, []ChannelModelPricing{
		tokenPricingForModels([]string{"gpt-5.6-luna"}, channelPricingExpectedChannelCost),
	})
	require.InDelta(t, channelPricingExpectedChannelCost, log.InputCost, 1e-9)
}

func TestChannelPricing_SuffixedModelUsesNormalizedChannelPricing(t *testing.T) {
	log := recordUsageWithChannelPricing(t, "gpt-5.6-luna-high", false, []ChannelModelPricing{
		tokenPricingForModels([]string{"gpt-5.6-luna"}, channelPricingExpectedChannelCost),
	})
	require.InDelta(t, channelPricingExpectedChannelCost, log.InputCost, 1e-9)
}

func TestChannelPricing_DateSuffixedModelUsesNormalizedChannelPricing(t *testing.T) {
	log := recordUsageWithChannelPricing(t, "gpt-5.6-luna-2026-08-01", false, []ChannelModelPricing{
		tokenPricingForModels([]string{"gpt-5.6-luna"}, channelPricingExpectedChannelCost),
	})
	require.InDelta(t, channelPricingExpectedChannelCost, log.InputCost, 1e-9)
}

func TestChannelPricing_ExactVariantWinsOverNormalizedBaseName(t *testing.T) {
	log := recordUsageWithChannelPricing(t, "gpt-5.6-luna-high", false, []ChannelModelPricing{
		tokenPricingForModels([]string{"gpt-5.6-luna-high"}, channelPricingUnrelatedCost),
		tokenPricingForModels([]string{"gpt-5.6-luna"}, channelPricingExpectedChannelCost),
	})
	require.InDelta(t, channelPricingUnrelatedCost, log.InputCost, 1e-9)
}

func TestChannelPricing_SuffixedModelSubscriptionGroup(t *testing.T) {
	log := recordUsageWithChannelPricing(t, "gpt-5.6-luna-high", true, []ChannelModelPricing{
		tokenPricingForModels([]string{"gpt-5.6-luna"}, channelPricingExpectedChannelCost),
	})
	require.InDelta(t, channelPricingExpectedChannelCost, log.InputCost, 1e-9)
}

func TestChannelPricing_UnrelatedChannelModelNotMatched(t *testing.T) {
	log := recordUsageWithChannelPricing(t, "gpt-5.6-luna-high", false, []ChannelModelPricing{
		tokenPricingForModels([]string{"gpt-5.4"}, channelPricingUnrelatedCost),
	})
	require.InDelta(t, channelPricingExpectedOfficialCost, log.InputCost, 1e-9)
}
