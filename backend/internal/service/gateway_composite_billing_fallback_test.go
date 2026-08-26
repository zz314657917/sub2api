package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestCompositeBillableModel(t *testing.T) {
	billingService := NewBillingService(&config.Config{}, nil)
	svc := &GatewayService{billingService: billingService}
	apiKey := &APIKey{Group: &Group{Platform: PlatformComposite}}
	ctx := context.Background()

	// 无显式别名定价时，即使别名会触发家族兜底，也必须按实际 Opus 模型计费。
	require.Equal(t, "claude-opus-4-7",
		svc.compositeBillableModel(ctx, apiKey, "all/claude", "claude-opus-4-7"))
	require.Equal(t, "claude-sonnet-4",
		svc.compositeBillableModel(ctx, apiKey, "team/best", "claude-sonnet-4"))

	// 计费模型本来就是具体模型时不改变。
	require.Equal(t, "claude-sonnet-4",
		svc.compositeBillableModel(ctx, apiKey, "claude-sonnet-4", "claude-sonnet-4"))

	// 缺少具体模型时保持原值，交给后续通用兜底或既有零成本路径处理。
	require.Equal(t, "all/claude",
		svc.compositeBillableModel(ctx, apiKey, "all/claude", ""))

	// 本地分组定价是管理员显式别名价格，优先级与渠道定价一致，不能被回退覆盖。
	price := 2e-6
	apiKey.Group.ModelPricing = []ChannelModelPricing{{
		Models:     []string{"all/claude"},
		InputPrice: &price,
	}}
	svc.resolver = NewModelPricingResolver(nil, billingService)
	require.Equal(t, "all/claude",
		svc.compositeBillableModel(ctx, apiKey, "all/claude", "claude-opus-4-7"))
}

func TestBillableModelWithFallback(t *testing.T) {
	svc := &GatewayService{billingService: NewBillingService(&config.Config{}, nil)}
	apiKey := &APIKey{}
	ctx := context.Background()

	// 完全无价的别名回退到具体转发模型。
	require.Equal(t, "claude-sonnet-4",
		svc.billableModelWithFallback(ctx, apiKey, "team/best", "", "claude-sonnet-4"))

	// 已定价模型不回退，候选被忽略。
	require.Equal(t, "claude-sonnet-4",
		svc.billableModelWithFallback(ctx, apiKey, "claude-sonnet-4", "claude-opus-4"))

	// 所有候选都无价时保持原值，走既有 warn + 零成本路径。
	require.Equal(t, "team/best",
		svc.billableModelWithFallback(ctx, apiKey, "team/best", "another/alias", ""))

	// 空计费模型有可定价候选时，使用候选。
	require.Equal(t, "claude-sonnet-4",
		svc.billableModelWithFallback(ctx, apiKey, "", "claude-sonnet-4"))
}

func TestHasResolvableTokenPricing(t *testing.T) {
	svc := &GatewayService{billingService: NewBillingService(&config.Config{}, nil)}
	apiKey := &APIKey{}
	ctx := context.Background()

	require.True(t, svc.hasResolvableTokenPricing(ctx, "claude-sonnet-4", apiKey))
	// 含家族词的别名会被价格表家族兜底解析为有价，compositeBillableModel 必须先拦截它。
	require.True(t, svc.hasResolvableTokenPricing(ctx, "all/claude", apiKey))
	require.False(t, svc.hasResolvableTokenPricing(ctx, "team/best", apiKey))
	require.False(t, svc.hasResolvableTokenPricing(ctx, "", apiKey))

	// billingService 缺失时 fail-closed，不能误判为有价。
	empty := &GatewayService{}
	require.False(t, empty.hasResolvableTokenPricing(ctx, "claude-sonnet-4", apiKey))
}
