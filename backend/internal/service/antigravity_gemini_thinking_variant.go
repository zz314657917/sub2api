package service

import (
	"encoding/json"
	"strings"
)

var geminiThinkingVariantSuffixes = []string{"-low", "-medium", "-high", "-tiered"}

const (
	geminiThinkingBudgetLowMax    = 1024
	geminiThinkingBudgetMediumMax = 8192
)

type geminiThinkingConfigProbe struct {
	GenerationConfig struct {
		ThinkingConfig *struct {
			ThinkingBudget *json.Number `json:"thinkingBudget"`
			ThinkingLevel  string       `json:"thinkingLevel"`
		} `json:"thinkingConfig"`
	} `json:"generationConfig"`
}

func hasGeminiThinkingVariantSuffix(model string) bool {
	for _, suffix := range geminiThinkingVariantSuffixes {
		if strings.HasSuffix(model, suffix) {
			return true
		}
	}
	return false
}

func geminiThinkingLevelFromBody(body []byte) string {
	if len(body) == 0 {
		return "high"
	}
	var probe geminiThinkingConfigProbe
	if err := json.Unmarshal(body, &probe); err != nil || probe.GenerationConfig.ThinkingConfig == nil {
		return "high"
	}
	tc := probe.GenerationConfig.ThinkingConfig
	switch strings.ToLower(strings.TrimSpace(tc.ThinkingLevel)) {
	case "low":
		return "low"
	case "medium":
		return "medium"
	case "high":
		return "high"
	}
	if tc.ThinkingBudget == nil {
		return "high"
	}
	budget, err := tc.ThinkingBudget.Float64()
	if err != nil {
		return "high"
	}
	switch {
	case budget < 0:
		return "high"
	case budget <= geminiThinkingBudgetLowMax:
		return "low"
	case budget <= geminiThinkingBudgetMediumMax:
		return "medium"
	default:
		return "high"
	}
}

func accountRawModelMappingHasKey(account *Account, key string) bool {
	if account == nil || account.Credentials == nil {
		return false
	}
	raw, _ := account.Credentials["model_mapping"].(map[string]any)
	_, ok := raw[key]
	return ok
}

// resolveGeminiThinkingVariant derives a configured variant only for a bare Gemini model.
// Explicit account mappings, including identity mappings, always take precedence.
func resolveGeminiThinkingVariant(account *Account, requestedModel string, body []byte) (string, bool) {
	if account == nil {
		return "", false
	}
	model := strings.TrimSpace(strings.TrimPrefix(requestedModel, "models/"))
	if !strings.HasPrefix(model, "gemini-") || hasGeminiThinkingVariantSuffix(model) {
		return "", false
	}
	mapping := account.GetModelMapping()
	if len(mapping) == 0 {
		return "", false
	}
	if mapped, matched := resolveRequestedModelInMapping(mapping, model); matched {
		if strings.TrimSpace(mapped) != model || accountRawModelMappingHasKey(account, model) {
			return "", false
		}
	}

	preferred := geminiThinkingLevelFromBody(body)
	order := []string{preferred}
	for _, level := range []string{"high", "medium", "low", "tiered"} {
		if level != preferred {
			order = append(order, level)
		}
	}
	for _, level := range order {
		if mapped, matched := resolveRequestedModelInMapping(mapping, model+"-"+level); matched && strings.TrimSpace(mapped) != "" {
			return mapped, true
		}
	}
	return "", false
}
