package service

import "strings"

// kimi31ProPricingKey deliberately accepts only the published standard model and
// finite user spellings. Reserved variants cannot inherit a fuzzy default card.
func kimi31ProPricingKey(model string) (known, reserved bool) {
	model = strings.ToLower(strings.TrimSpace(model))
	bare := model
	if strings.HasPrefix(bare, "kimi/") {
		bare = strings.TrimPrefix(bare, "kimi/")
	} else if strings.HasPrefix(bare, "moonshot/") {
		bare = strings.TrimPrefix(bare, "moonshot/")
	}
	switch bare {
	case "kimi-k3.1-pro", "kimi3.1pro", "kimi-k3-1-pro", "kimi-k3.1pro", "k3.1-pro":
		return true, true
	}
	// Inspect the last segment for isolation, never for accepting unknown providers
	// or nested prefixes as aliases.
	segment := model[strings.LastIndex(model, "/")+1:]
	return false, strings.HasPrefix(segment, "kimi-k3.1") ||
		strings.HasPrefix(segment, "kimi-k3-1") ||
		strings.HasPrefix(segment, "kimi3.1") ||
		strings.HasPrefix(segment, "k3.1")
}

// CNY and site credits are intentionally 1:1 for this domestic standard card.
// A new value on every call avoids mutating shared dynamic/default cards.
func kimi31ProDefaultPricing() *LiteLLMModelPricing {
	return &LiteLLMModelPricing{
		InputCostPerToken: 13e-6, OutputCostPerToken: 68e-6,
		CacheReadInputTokenCost:     1.3e-6,
		CacheCreationInputTokenCost: 13e-6, CacheCreationInputTokenCostAbove1hr: 26e-6,
		CacheReadInputTokenCostExplicit: true, CacheCreationInputTokenCostExplicit: true,
		CacheCreationInputTokenCostAbove1hrExplicit: true,
		LongContextMetadataPresent:                  true,
		LiteLLMProvider:                             "moonshot", Mode: "chat", SupportsPromptCaching: true,
	}
}
