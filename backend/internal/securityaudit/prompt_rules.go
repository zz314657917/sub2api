package securityaudit

import (
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// RiskActionRules controls only risk escalation. It is intentionally separate
// from scanner selection so policy changes do not require changing the guard
// transport or prompt retention pipeline.
type RiskActionRules struct {
	Safety     map[string]Action `json:"safety,omitempty"`
	Categories map[string]Action `json:"categories,omitempty"`
}

func normalizeRiskActionRules(rules *RiskActionRules) {
	if rules == nil {
		return
	}
	rules.Safety = normalizeSafetyRules(rules.Safety)
	rules.Categories = normalizeCategoryRules(rules.Categories)
}

func normalizeSafetyRules(values map[string]Action) map[string]Action {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]Action, len(values))
	for key, action := range values {
		key = strings.ToLower(strings.TrimSpace(key))
		if key != "safe" && key != "controversial" && key != "unsafe" {
			result[key] = action
			continue
		}
		result[key] = canonicalAction(action)
	}
	return result
}

func normalizeCategoryRules(values map[string]Action) map[string]Action {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]Action, len(values))
	for key, action := range values {
		result[NormalizeCategory(key)] = canonicalAction(action)
	}
	return result
}

func canonicalAction(value Action) Action {
	switch strings.ToLower(strings.TrimSpace(string(value))) {
	case "allow":
		return ActionAllow
	case "warn":
		return ActionWarn
	case "block":
		return ActionBlock
	default:
		return value
	}
}

func validateRiskActionRules(rules RiskActionRules) error {
	for safety, action := range rules.Safety {
		if safety != "safe" && safety != "controversial" && safety != "unsafe" {
			return badRules("prompt_audit_invalid_rule_safety", "提示词审计安全等级规则无效")
		}
		switch safety {
		case "safe":
			if action != ActionAllow {
				return badRules("prompt_audit_invalid_rule_action", "Safe 规则只能设置为 Allow")
			}
		case "controversial":
			if action != ActionWarn && action != ActionBlock {
				return badRules("prompt_audit_invalid_rule_action", "Controversial 规则只能设置为 Warn 或 Block")
			}
		case "unsafe":
			if action != ActionBlock {
				return badRules("prompt_audit_invalid_rule_action", "Unsafe 规则必须设置为 Block")
			}
		}
	}
	for category, action := range rules.Categories {
		if _, ok := ScannerCatalog[category]; !ok {
			return badRules("prompt_audit_invalid_rule_category", "提示词审计分类规则无效")
		}
		if action != ActionWarn && action != ActionBlock {
			return badRules("prompt_audit_invalid_rule_action", "分类规则只能设置为 Warn 或 Block")
		}
	}
	return nil
}

func badRules(code, message string) error {
	return infraerrors.BadRequest(code, message)
}

// ApplyRiskActionRules applies only monotonic upgrades. Zero-value rules are
// a compatibility mode for configurations written before rules existed.
func ApplyRiskActionRules(result *NormalizedResult, rules RiskActionRules) {
	if result == nil {
		return
	}
	action := result.Action
	if override, ok := rules.Safety[strings.ToLower(strings.TrimSpace(result.Safety))]; ok {
		action = higherAction(action, override)
	}
	for _, category := range result.Categories {
		if override, ok := rules.Categories[NormalizeCategory(category)]; ok {
			action = higherAction(action, override)
		}
	}
	if action == result.Action {
		return
	}
	result.Action = action
	if action == ActionBlock {
		result.Decision, result.RiskLevel = EventCritical, RiskCritical
		return
	}
	if action == ActionWarn && result.Decision != EventCritical {
		result.Decision, result.RiskLevel = EventFlag, RiskMedium
	}
}

func higherAction(current, candidate Action) Action {
	if actionRank(candidate) > actionRank(current) {
		return candidate
	}
	return current
}

func actionRank(action Action) int {
	switch action {
	case ActionBlock:
		return 2
	case ActionWarn:
		return 1
	default:
		return 0
	}
}

func cloneRiskActionRules(rules RiskActionRules) RiskActionRules {
	copy := RiskActionRules{}
	if len(rules.Safety) > 0 {
		copy.Safety = make(map[string]Action, len(rules.Safety))
		for key, action := range rules.Safety {
			copy.Safety[key] = action
		}
	}
	if len(rules.Categories) > 0 {
		copy.Categories = make(map[string]Action, len(rules.Categories))
		for key, action := range rules.Categories {
			copy.Categories[key] = action
		}
	}
	return copy
}
