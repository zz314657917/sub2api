package service

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func geminiSameAccountMaxAttempts(cfg *config.Config) int {
	if cfg == nil {
		return 5
	}
	return cfg.Gateway.MaxSameAccountRetriesGemini + 1
}

func isGeminiThoughtPart(part map[string]any) bool {
	thought, ok := part["thought"].(bool)
	return ok && thought
}

// These are pinned upstream catalog values, not endpoint capability declarations.
// Return fresh cards so callers cannot mutate a shared fallback.
func geminiUpstreamDefaultPricing(model string) (*LiteLLMModelPricing, bool) {
	model = normalizeModelNameForPricing(strings.ToLower(strings.TrimSpace(model)))
	for _, prefix := range []string{"google/", "antigravity/"} {
		model = strings.TrimPrefix(model, prefix)
	}
	p := &LiteLLMModelPricing{LiteLLMProvider: "vertex_ai-language-models", Mode: "chat", SupportsPromptCaching: true}
	switch model {
	case "gemini-3-flash":
		p.InputCostPerToken, p.OutputCostPerToken, p.CacheReadInputTokenCost = 5e-7, 3e-6, 5e-8
		p.InputCostPerTokenPriority, p.OutputCostPerTokenPriority, p.CacheReadInputTokenCostPriority = 9e-7, 5.4e-6, 9e-8
		p.SupportsServiceTier = true
	case "gemini-3.1-flash-lite":
		p.InputCostPerToken, p.OutputCostPerToken, p.CacheReadInputTokenCost = 2.5e-7, 1.5e-6, 2.5e-8
		p.InputCostPerTokenPriority, p.OutputCostPerTokenPriority, p.CacheReadInputTokenCostPriority = 4.5e-7, 2.7e-6, 4.5e-8
		p.SupportsServiceTier = true
	case "gemini-3.1-flash-lite-preview":
		p.InputCostPerToken, p.OutputCostPerToken, p.CacheReadInputTokenCost = 2.5e-7, 1.5e-6, 2.5e-8
	case "gemini-3.5-flash":
		p.InputCostPerToken, p.OutputCostPerToken, p.CacheReadInputTokenCost = 1.5e-6, 9e-6, 1.5e-7
		p.InputCostPerTokenPriority, p.OutputCostPerTokenPriority, p.CacheReadInputTokenCostPriority = 2.7e-6, 1.62e-5, 2.7e-7
		p.SupportsServiceTier = true
	case "gemini-3-pro-image":
		p.Mode = "image_generation"
		p.InputCostPerToken, p.OutputCostPerToken = 2e-6, 1.2e-5
		p.InputCostPerImage, p.OutputCostPerImage, p.OutputCostPerImageToken = .0011, .134, .00012
		p.SupportsServiceTier = true
	case "gemini-3.1-flash-image", "gemini-3.1-flash-image-preview":
		p.Mode = "image_generation"
		p.InputCostPerToken, p.OutputCostPerToken = 5e-7, 3e-6
		p.InputCostPerImage, p.OutputCostPerImage, p.OutputCostPerImageToken = .00056, .0672, .00006
	case "gemini-3.1-flash-lite-image":
		p.Mode = "image_generation"
		p.InputCostPerToken, p.OutputCostPerToken = 3e-7, 2.5e-6
		p.InputCostPerImage, p.OutputCostPerImage, p.OutputCostPerImageToken = .0003, .034, .00003
	default:
		// Reserve siblings before the generic fuzzy matcher can select these cards.
		for _, base := range []string{"gemini-3-flash", "gemini-3-pro-image", "gemini-3.1-flash", "gemini-3.5-flash"} {
			if strings.HasPrefix(model, base) {
				return nil, true
			}
		}
		return nil, false
	}
	p.CacheReadInputTokenCostExplicit = p.CacheReadInputTokenCost != 0
	p.InputCostPerTokenPriorityExplicit = p.InputCostPerTokenPriority != 0
	p.OutputCostPerTokenPriorityExplicit = p.OutputCostPerTokenPriority != 0
	p.CacheReadInputTokenCostPriorityExplicit = p.CacheReadInputTokenCostPriority != 0
	return p, true
}
