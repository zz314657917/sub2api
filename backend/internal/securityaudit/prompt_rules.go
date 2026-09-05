package securityaudit

import (
	"encoding/json"
	"sort"
	"strings"
	"unicode/utf8"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	maxPolicyTextLength = 128
	maxPolicyJSONBytes  = 256 * 1024
	maxPolicyRules      = 256
	maxOWASPTagsPerRule = 16
	maxOWASPTagLength   = 64
)

type RiskPolicyAction struct {
	Action    Action    `json:"action"`
	RiskLevel RiskLevel `json:"risk_level,omitempty"`
}

type RiskPolicyRule struct {
	ID          string    `json:"id"`
	Priority    int       `json:"priority"`
	Safety      []string  `json:"safety,omitempty"`
	Categories  []string  `json:"categories,omitempty"`
	Groups      []int64   `json:"groups,omitempty"`
	Models      []string  `json:"models,omitempty"`
	Providers   []string  `json:"providers,omitempty"`
	Action      Action    `json:"action"`
	RiskLevel   RiskLevel `json:"risk_level,omitempty"`
	MessageCode string    `json:"message_code,omitempty"`
	OWASP       []string  `json:"owasp_tags,omitempty"`
}

// S292 map fields remain readable; new policies should use Defaults and Rules.
type RiskActionRules struct {
	PolicyID      string                      `json:"policy_id,omitempty"`
	PolicyVersion int                         `json:"policy_version,omitempty"`
	Defaults      map[string]RiskPolicyAction `json:"defaults,omitempty"`
	Rules         []RiskPolicyRule            `json:"rules,omitempty"`
	Safety        map[string]Action           `json:"safety,omitempty"`
	Categories    map[string]Action           `json:"categories,omitempty"`
}

type PolicyMatchContext struct {
	GroupID  *int64 `json:"group_id,omitempty"`
	Model    string `json:"model,omitempty"`
	Provider string `json:"provider,omitempty"`
}

func normalizeRiskActionRules(rules *RiskActionRules) {
	if rules == nil {
		return
	}
	rules.PolicyID = strings.TrimSpace(rules.PolicyID)
	if rules.PolicyVersion < 0 {
		rules.PolicyVersion = 0
	}
	rules.Safety = normalizeSafetyRules(rules.Safety)
	rules.Categories = normalizeCategoryRules(rules.Categories)
	if len(rules.Defaults) > 0 {
		normalized := make(map[string]RiskPolicyAction, len(rules.Defaults))
		for key, value := range rules.Defaults {
			normalized[strings.ToLower(strings.TrimSpace(key))] = normalizePolicyAction(value)
		}
		rules.Defaults = normalized
	}
	for index := range rules.Rules {
		rule := &rules.Rules[index]
		rule.ID = strings.TrimSpace(rule.ID)
		rule.Safety = normalizeSafetyValues(rule.Safety)
		rule.Categories = normalizeCategoryValues(rule.Categories)
		rule.Models = normalizeStringValues(rule.Models)
		rule.Providers = normalizeStringValues(rule.Providers)
		rule.Action = canonicalAction(rule.Action)
		rule.RiskLevel = canonicalRiskLevel(rule.RiskLevel)
		rule.MessageCode = strings.TrimSpace(rule.MessageCode)
		rule.OWASP = normalizeOWASPValues(rule.OWASP)
	}
	sort.SliceStable(rules.Rules, func(i, j int) bool {
		if rules.Rules[i].Priority != rules.Rules[j].Priority {
			return rules.Rules[i].Priority > rules.Rules[j].Priority
		}
		return rules.Rules[i].ID < rules.Rules[j].ID
	})
}

func normalizeOWASPValues(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		key := strings.ToUpper(strings.TrimSpace(value))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func normalizeSafetyRules(values map[string]Action) map[string]Action {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]Action, len(values))
	for key, action := range values {
		result[strings.ToLower(strings.TrimSpace(key))] = canonicalAction(action)
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

func normalizeSafetyValues(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		key := strings.ToLower(strings.TrimSpace(value))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func normalizeCategoryValues(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		key := NormalizeCategory(value)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func normalizeStringValues(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		key := strings.ToLower(strings.TrimSpace(value))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func normalizePolicyAction(value RiskPolicyAction) RiskPolicyAction {
	value.Action = canonicalAction(value.Action)
	value.RiskLevel = canonicalRiskLevel(value.RiskLevel)
	return value
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

func canonicalRiskLevel(value RiskLevel) RiskLevel {
	switch strings.ToLower(strings.TrimSpace(string(value))) {
	case "low":
		return RiskLow
	case "medium":
		return RiskMedium
	case "high":
		return RiskHigh
	case "critical":
		return RiskCritical
	default:
		return value
	}
}

func validateRiskActionRules(rules RiskActionRules) error {
	if utf8.RuneCountInString(strings.TrimSpace(rules.PolicyID)) > maxPolicyTextLength {
		return badRules("prompt_audit_invalid_policy_id", "策略 ID 长度超出限制")
	}
	if len(rules.Rules) > maxPolicyRules {
		return badRules("prompt_audit_invalid_rule_count", "策略规则数量超出限制")
	}
	for safety, action := range rules.Safety {
		if safety != "safe" && safety != "controversial" && safety != "unsafe" {
			return badRules("prompt_audit_invalid_rule_safety", "提示词审计安全等级规则无效")
		}
		if err := validateSafetyAction(safety, action); err != nil {
			return err
		}
		if !isKnownAction(action) {
			return badRules("prompt_audit_invalid_rule_action", "策略动作无效")
		}
	}
	for safety, value := range rules.Defaults {
		if safety != "safe" && safety != "controversial" && safety != "unsafe" {
			return badRules("prompt_audit_invalid_rule_safety", "提示词审计默认安全等级无效")
		}
		if err := validateSafetyAction(safety, value.Action); err != nil {
			return err
		}
		if err := validateRiskForAction(value.Action, value.RiskLevel); err != nil {
			return err
		}
		if !isKnownAction(value.Action) || !isKnownRiskLevel(value.RiskLevel) {
			return badRules("prompt_audit_invalid_rule_value", "策略动作或风险等级无效")
		}
	}
	for category, action := range rules.Categories {
		if _, ok := ScannerCatalog[category]; !ok {
			return badRules("prompt_audit_invalid_rule_category", "提示词审计分类规则无效")
		}
		if action != ActionWarn && action != ActionBlock {
			return badRules("prompt_audit_invalid_rule_action", "分类规则只能设置为 Warn 或 Block")
		}
		if !isKnownAction(action) {
			return badRules("prompt_audit_invalid_rule_action", "策略动作无效")
		}
	}
	seenIDs := map[string]struct{}{}
	for _, rule := range rules.Rules {
		if rule.ID == "" {
			return badRules("prompt_audit_invalid_rule_id", "策略规则 ID 不能为空")
		}
		if utf8.RuneCountInString(rule.ID) > maxPolicyTextLength {
			return badRules("prompt_audit_invalid_rule_id", "策略规则 ID 长度超出限制")
		}
		if _, exists := seenIDs[rule.ID]; exists {
			return badRules("prompt_audit_duplicate_rule_id", "策略规则 ID 不能重复")
		}
		seenIDs[rule.ID] = struct{}{}
		for _, safety := range rule.Safety {
			if safety != "safe" && safety != "controversial" && safety != "unsafe" {
				return badRules("prompt_audit_invalid_rule_safety", "策略规则安全等级无效")
			}
		}
		for _, category := range rule.Categories {
			if _, ok := ScannerCatalog[category]; !ok {
				return badRules("prompt_audit_invalid_rule_category", "策略规则分类无效")
			}
		}
		for _, groupID := range rule.Groups {
			if groupID <= 0 {
				return badRules("prompt_audit_invalid_rule_group", "策略规则分组 ID 无效")
			}
		}
		if err := validateRiskForAction(rule.Action, rule.RiskLevel); err != nil {
			return err
		}
		if !isKnownAction(rule.Action) || !isKnownRiskLevel(rule.RiskLevel) {
			return badRules("prompt_audit_invalid_rule_value", "策略动作或风险等级无效")
		}
		if rule.Action == ActionAllow {
			return badRules("prompt_audit_invalid_rule_action", "策略规则不能配置为 Allow")
		}
		for _, tag := range rule.OWASP {
			if !strings.HasPrefix(tag, "LLM") && !strings.HasPrefix(tag, "OWASP") {
				return badRules("prompt_audit_invalid_owasp_tag", "OWASP 标签格式无效")
			}
			if utf8.RuneCountInString(tag) > maxOWASPTagLength {
				return badRules("prompt_audit_invalid_owasp_tag", "OWASP 标签长度超出限制")
			}
		}
		if len(rule.OWASP) > maxOWASPTagsPerRule {
			return badRules("prompt_audit_invalid_owasp_tag", "OWASP 标签数量超出限制")
		}
	}
	encoded, err := json.Marshal(rules)
	if err != nil {
		return badRules("prompt_audit_invalid_rules", "策略规则无法序列化")
	}
	if len(encoded) > maxPolicyJSONBytes {
		return badRules("prompt_audit_invalid_rules_size", "策略规则大小超出限制")
	}
	return nil
}

func isKnownAction(action Action) bool {
	return action == ActionAllow || action == ActionWarn || action == ActionBlock
}

func isKnownRiskLevel(risk RiskLevel) bool {
	return risk == "" || risk == RiskLow || risk == RiskMedium || risk == RiskHigh || risk == RiskCritical
}

func validateSafetyAction(safety string, action Action) error {
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
	return nil
}

func validateRiskForAction(action Action, risk RiskLevel) error {
	if risk == "" {
		return nil
	}
	if action == ActionBlock && risk != RiskHigh && risk != RiskCritical {
		return badRules("prompt_audit_invalid_rule_risk", "Block 规则风险等级至少为 high")
	}
	if action == ActionWarn && risk != RiskMedium && risk != RiskHigh && risk != RiskCritical {
		return badRules("prompt_audit_invalid_rule_risk", "Warn 规则风险等级至少为 medium")
	}
	return nil
}

func badRules(code, message string) error {
	return infraerrors.BadRequest(code, message)
}

func ApplyRiskActionRules(result *NormalizedResult, rules RiskActionRules) {
	ApplyRiskPolicy(result, rules, PolicyMatchContext{})
}

// ShadowEvaluate applies a candidate policy to a copy and never mutates the
// supplied result. It is intended for dry-run previews and cannot affect the
// request decision path.
func ShadowEvaluate(result NormalizedResult, rules RiskActionRules, context PolicyMatchContext) (PolicyShadowResult, error) {
	if err := validateRiskActionRules(rules); err != nil {
		return PolicyShadowResult{}, err
	}
	candidate := result
	candidate.Categories = append([]string(nil), result.Categories...)
	candidate.OWASPTags = append([]string(nil), result.OWASPTags...)
	ApplyRiskPolicy(&candidate, rules, context)
	shadow := PolicyShadowResult{
		Current: result, Candidate: candidate,
		ActionChanged: result.Action != candidate.Action,
		RiskChanged:   result.RiskLevel != candidate.RiskLevel,
	}
	shadow.WouldEscalate = actionRank(candidate.Action) > actionRank(result.Action) ||
		(actionRank(candidate.Action) == actionRank(result.Action) && riskRank(candidate.RiskLevel) > riskRank(result.RiskLevel))
	return shadow, nil
}

func ApplyRiskPolicy(result *NormalizedResult, rules RiskActionRules, context PolicyMatchContext) {
	if result == nil {
		return
	}
	policyID := strings.TrimSpace(rules.PolicyID)
	if policyID == "" {
		policyID = result.PolicyID
	}
	if policyID == "" {
		policyID = "priority"
	}
	result.PolicyID = policyID
	if rules.PolicyVersion > 0 {
		result.PolicyVersion = rules.PolicyVersion
	} else if result.PolicyVersion < 1 {
		result.PolicyVersion = 1
	}

	current := RiskPolicyAction{Action: result.Action, RiskLevel: result.RiskLevel}
	if strings.EqualFold(strings.TrimSpace(result.Safety), "unsafe") || len(result.UnknownCategories) > 0 {
		current = strongerPolicyAction(current, RiskPolicyAction{Action: ActionBlock, RiskLevel: RiskCritical})
	}
	for _, category := range result.Categories {
		if action, ok := rules.Categories[NormalizeCategory(category)]; ok {
			current = strongerPolicyAction(current, RiskPolicyAction{Action: action})
		}
	}
	if action, ok := rules.Safety[strings.ToLower(strings.TrimSpace(result.Safety))]; ok {
		current = strongerPolicyAction(current, RiskPolicyAction{Action: action})
	}
	if defaultAction, ok := rules.Defaults[strings.ToLower(strings.TrimSpace(result.Safety))]; ok {
		current = strongerPolicyAction(current, defaultAction)
	}
	base := current
	result.MatchedRuleID = ""
	result.matchedRulePriority = 0
	matched := make([]RiskPolicyRule, 0, len(rules.Rules))
	for _, rule := range rules.Rules {
		if ruleMatches(rule, result, context) {
			matched = append(matched, rule)
			result.OWASPTags = appendUniqueStrings(result.OWASPTags, rule.OWASP...)
			current = strongerPolicyAction(current, RiskPolicyAction{Action: rule.Action, RiskLevel: rule.RiskLevel})
		}
	}
	applyPolicyAction(result, current)
	// Only identify a rule when it actually contributes the final escalation;
	// a higher-priority warning must not be reported for a lower-priority block.
	if len(matched) > 0 {
		winner := RiskPolicyRule{}
		var winnerAction RiskPolicyAction
		for _, rule := range matched {
			candidate := withMinimumRisk(RiskPolicyAction{Action: rule.Action, RiskLevel: rule.RiskLevel})
			if actionRank(candidate.Action) != actionRank(result.Action) || riskRank(candidate.RiskLevel) != riskRank(result.RiskLevel) {
				continue
			}
			if winner.ID == "" || comparePolicyRule(rule, winner) < 0 {
				winner, winnerAction = rule, candidate
			}
		}
		if winner.ID != "" && (actionRank(winnerAction.Action) > actionRank(base.Action) || riskRank(winnerAction.RiskLevel) > riskRank(base.RiskLevel)) {
			result.MatchedRuleID = winner.ID
			result.matchedRulePriority = winner.Priority
		}
	}
}

func appendUniqueStrings(values []string, additions ...string) []string {
	seen := make(map[string]struct{}, len(values)+len(additions))
	for _, value := range values {
		if value != "" {
			seen[value] = struct{}{}
		}
	}
	for _, value := range additions {
		if value != "" {
			seen[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func ruleMatches(rule RiskPolicyRule, result *NormalizedResult, context PolicyMatchContext) bool {
	if len(rule.Safety) > 0 && !containsString(rule.Safety, strings.ToLower(strings.TrimSpace(result.Safety))) {
		return false
	}
	if len(rule.Categories) > 0 && !intersects(rule.Categories, result.Categories) {
		return false
	}
	if len(rule.Groups) > 0 && (context.GroupID == nil || !containsInt64(rule.Groups, *context.GroupID)) {
		return false
	}
	if len(rule.Models) > 0 && !containsString(rule.Models, strings.ToLower(strings.TrimSpace(context.Model))) {
		return false
	}
	if len(rule.Providers) > 0 && !containsString(rule.Providers, strings.ToLower(strings.TrimSpace(context.Provider))) {
		return false
	}
	return true
}

func strongerPolicyAction(current, candidate RiskPolicyAction) RiskPolicyAction {
	candidate = normalizePolicyAction(candidate)
	if candidate.Action == ActionAllow || actionRank(candidate.Action) < actionRank(current.Action) {
		return current
	}
	if actionRank(candidate.Action) > actionRank(current.Action) {
		return withMinimumRisk(candidate)
	}
	if riskRank(candidate.RiskLevel) > riskRank(current.RiskLevel) {
		return candidate
	}
	return current
}

func withMinimumRisk(value RiskPolicyAction) RiskPolicyAction {
	if value.Action == ActionBlock && riskRank(value.RiskLevel) < riskRank(RiskCritical) {
		value.RiskLevel = RiskCritical
	}
	if value.Action == ActionWarn && riskRank(value.RiskLevel) < riskRank(RiskMedium) {
		value.RiskLevel = RiskMedium
	}
	return value
}

func applyPolicyAction(result *NormalizedResult, value RiskPolicyAction) {
	value = withMinimumRisk(value)
	result.Action = value.Action
	if riskRank(value.RiskLevel) > riskRank(result.RiskLevel) {
		result.RiskLevel = value.RiskLevel
	}
	if result.Action == ActionBlock {
		result.Decision, result.RiskLevel = EventCritical, RiskCritical
		return
	}
	if result.Action == ActionWarn && result.Decision != EventCritical {
		result.Decision, result.RiskLevel = EventFlag, value.RiskLevel
	}
}

func comparePolicyRule(left, right RiskPolicyRule) int {
	if left.Priority != right.Priority {
		if left.Priority > right.Priority {
			return -1
		}
		return 1
	}
	return strings.Compare(left.ID, right.ID)
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

func riskRank(risk RiskLevel) int {
	switch risk {
	case RiskCritical:
		return 4
	case RiskHigh:
		return 3
	case RiskMedium:
		return 2
	case RiskLow:
		return 1
	default:
		return 0
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func intersects(values []string, categories []string) bool {
	for _, value := range values {
		if containsString(categories, value) {
			return true
		}
	}
	return false
}

func containsInt64(values []int64, target int64) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func cloneRiskActionRules(rules RiskActionRules) RiskActionRules {
	copy := RiskActionRules{PolicyID: rules.PolicyID, PolicyVersion: rules.PolicyVersion}
	if len(rules.Defaults) > 0 {
		copy.Defaults = make(map[string]RiskPolicyAction, len(rules.Defaults))
		for key, action := range rules.Defaults {
			copy.Defaults[key] = action
		}
	}
	if len(rules.Rules) > 0 {
		copy.Rules = make([]RiskPolicyRule, len(rules.Rules))
		for index, rule := range rules.Rules {
			copy.Rules[index] = rule
			copy.Rules[index].Safety = append([]string(nil), rule.Safety...)
			copy.Rules[index].Categories = append([]string(nil), rule.Categories...)
			copy.Rules[index].Groups = append([]int64(nil), rule.Groups...)
			copy.Rules[index].Models = append([]string(nil), rule.Models...)
			copy.Rules[index].Providers = append([]string(nil), rule.Providers...)
			copy.Rules[index].OWASP = append([]string(nil), rule.OWASP...)
		}
	}
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
