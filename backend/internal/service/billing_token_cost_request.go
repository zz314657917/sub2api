package service

import "context"

// LegacyLongContextRule 平台级的边际长上下文规则。
// 只有 Gemini 原生入口在没有显式分组/渠道定价时使用该规则。
type LegacyLongContextRule struct {
	Threshold  int
	Multiplier float64
}

const (
	geminiLegacyLongContextThreshold  = 200000
	geminiLegacyLongContextMultiplier = 2.0
)

// LegacyLongContextRule 返回平台的旧长上下文规则；无规则的平台返回 nil。
func (s *BillingService) LegacyLongContextRule(platform string) *LegacyLongContextRule {
	if platform != PlatformGemini {
		return nil
	}
	return &LegacyLongContextRule{
		Threshold:  geminiLegacyLongContextThreshold,
		Multiplier: geminiLegacyLongContextMultiplier,
	}
}

// TokenCostRequest 是网关 token 计费的统一路径选择输入。
type TokenCostRequest struct {
	Ctx               context.Context
	Model             string
	Group             *Group
	Tokens            UsageTokens
	RateMultiplier    float64
	Resolver          *ModelPricingResolver
	Resolved          *ResolvedPricing
	LegacyLongContext *LegacyLongContextRule
}

func legacyLongContextApplies(resolved *ResolvedPricing, group *Group, rule *LegacyLongContextRule) bool {
	if rule == nil || rule.Threshold <= 0 || rule.Multiplier <= 1 {
		return false
	}
	if resolved != nil && (resolved.Source == PricingSourceGroup || resolved.Source == PricingSourceChannel) {
		return false
	}
	return group == nil || group.LongContextPricingEnabled
}

// CalculateTokenCostForRequest selects one token billing path. Explicit
// group/channel pricing wins; Gemini's legacy marginal rule is used next;
// built-in pricing then uses the resolver so the group toggle remains effective.
func (s *BillingService) CalculateTokenCostForRequest(req TokenCostRequest) (*CostBreakdown, error) {
	if req.Resolved != nil && (req.Resolved.Source == PricingSourceGroup || req.Resolved.Source == PricingSourceChannel) {
		return s.CalculateCostUnified(s.tokenCostInput(req))
	}
	if legacyLongContextApplies(req.Resolved, req.Group, req.LegacyLongContext) {
		return s.CalculateCostWithLongContext(
			req.Model,
			req.Tokens,
			req.RateMultiplier,
			req.LegacyLongContext.Threshold,
			req.LegacyLongContext.Multiplier,
		)
	}
	if req.Resolver != nil && req.Group != nil {
		return s.CalculateCostUnified(s.tokenCostInput(req))
	}
	return s.CalculateCost(req.Model, req.Tokens, req.RateMultiplier)
}

func (s *BillingService) tokenCostInput(req TokenCostRequest) CostInput {
	input := CostInput{
		Ctx:            req.Ctx,
		Model:          req.Model,
		Group:          req.Group,
		Tokens:         req.Tokens,
		RequestCount:   1,
		RateMultiplier: req.RateMultiplier,
		Resolver:       req.Resolver,
		Resolved:       req.Resolved,
	}
	if req.Group != nil {
		groupID := req.Group.ID
		input.GroupID = &groupID
	}
	return input
}
