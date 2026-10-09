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
	GroupID           *int64
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
// group/channel pricing wins, followed by catalog whole-session metadata.
// Gemini's legacy marginal rule is retained only for metadata-free cards.
func (s *BillingService) CalculateTokenCostForRequest(req TokenCostRequest) (*CostBreakdown, error) {
	// Resolve before selecting the legacy path so GroupID-only gates and catalog
	// metadata work even when the gateway did not pre-resolve pricing.
	if req.Resolved == nil && req.Resolver != nil {
		input := s.tokenCostInput(req)
		req.Resolved = req.Resolver.Resolve(req.Ctx, PricingInput{
			Model: req.Model, GroupID: input.GroupID, Group: req.Group,
		})
	}
	if req.Resolved != nil && (req.Resolved.Source == PricingSourceGroup || req.Resolved.Source == PricingSourceChannel) {
		return s.CalculateCostUnified(s.tokenCostInput(req))
	}
	pricing := (*ModelPricing)(nil)
	if req.Resolved != nil {
		pricing = req.Resolved.BasePricing
	}
	if pricing == nil {
		pricing, _ = s.GetModelPricing(req.Model)
	}
	if pricing != nil && pricing.LongContextMetadataPresent {
		return s.CalculateCostUnified(s.tokenCostInput(req))
	}
	if req.Resolved != nil && !req.Resolved.longContextPricingEnabled {
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
	return s.CalculateCostUnified(s.tokenCostInput(req))
}

func (s *BillingService) tokenCostInput(req TokenCostRequest) CostInput {
	input := CostInput{
		Ctx:            req.Ctx,
		Model:          req.Model,
		GroupID:        req.GroupID,
		Group:          req.Group,
		Tokens:         req.Tokens,
		RequestCount:   1,
		RateMultiplier: req.RateMultiplier,
		Resolver:       req.Resolver,
		Resolved:       req.Resolved,
	}
	if input.GroupID == nil && req.Group != nil {
		groupID := req.Group.ID
		input.GroupID = &groupID
	}
	return input
}
