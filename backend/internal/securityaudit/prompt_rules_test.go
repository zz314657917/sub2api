package securityaudit

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRiskActionRulesDefaultAndMonotonicEscalation(t *testing.T) {
	result, err := ParseQwen3Guard("Safety: Controversial\nCategories: Violent", AllScannerIDs)
	require.NoError(t, err)
	require.Equal(t, ActionWarn, result.Action)

	ApplyRiskActionRules(result, RiskActionRules{Safety: map[string]Action{"controversial": ActionBlock}})
	require.Equal(t, ActionBlock, result.Action)
	require.Equal(t, EventCritical, result.Decision)

	ApplyRiskActionRules(result, RiskActionRules{Safety: map[string]Action{"controversial": ActionWarn}, Categories: map[string]Action{"violent": ActionAllow}})
	require.Equal(t, ActionBlock, result.Action, "rules must not weaken an existing block")
}

func TestRiskActionRulesCategoryEscalation(t *testing.T) {
	result, err := ParseQwen3Guard("Safety: Controversial\nCategories: Violent", AllScannerIDs)
	require.NoError(t, err)
	ApplyRiskActionRules(result, RiskActionRules{Categories: map[string]Action{"violent": ActionBlock}})
	require.Equal(t, ActionBlock, result.Action)
	require.Equal(t, RiskCritical, result.RiskLevel)
}

func TestRiskActionRulesValidationRejectsWeakeningAndUnknownValues(t *testing.T) {
	cases := []RiskActionRules{
		{Safety: map[string]Action{"unsafe": ActionWarn}},
		{Safety: map[string]Action{"safe": ActionBlock}},
		{Safety: map[string]Action{"future": ActionBlock}},
		{Categories: map[string]Action{"not-a-scanner": ActionBlock}},
		{Categories: map[string]Action{"pii": ActionAllow}},
	}
	for _, rules := range cases {
		normalizeRiskActionRules(&rules)
		require.Error(t, validateRiskActionRules(rules))
	}
}

func TestRiskActionRulesJSONRoundTripAndLegacyCompatibility(t *testing.T) {
	rules := RiskActionRules{
		Safety:     map[string]Action{"controversial": ActionBlock},
		Categories: map[string]Action{"Prompt Injection": ActionBlock},
	}
	raw, err := json.Marshal(rules)
	require.NoError(t, err)
	var decoded RiskActionRules
	require.NoError(t, json.Unmarshal(raw, &decoded))
	normalizeRiskActionRules(&decoded)
	require.Equal(t, ActionBlock, decoded.Categories["jailbreak"])

	legacy, err := ParseStorageConfig(`{"enabled":false,"config_version":3}`)
	require.NoError(t, err)
	require.Empty(t, legacy.Rules.Safety)
	require.Empty(t, legacy.Rules.Categories)
}

func TestRiskPolicyMatrixMatchesScopePriorityAndExplainsWinner(t *testing.T) {
	groupID := int64(7)
	rules := RiskActionRules{
		PolicyID:      "ops-policy",
		PolicyVersion: 9,
		Defaults: map[string]RiskPolicyAction{
			"safe":          {Action: ActionAllow, RiskLevel: RiskLow},
			"controversial": {Action: ActionWarn, RiskLevel: RiskMedium},
			"unsafe":        {Action: ActionBlock, RiskLevel: RiskCritical},
		},
		Rules: []RiskPolicyRule{
			{ID: "wide-warning", Priority: 10, Categories: []string{"jailbreak"}, Action: ActionWarn, RiskLevel: RiskHigh},
			{ID: "group-block", Priority: 100, Safety: []string{"Controversial"}, Categories: []string{"jailbreak"}, Groups: []int64{7}, Models: []string{"qwen"}, Providers: []string{"openai"}, Action: ActionBlock, RiskLevel: RiskCritical},
		},
	}
	normalizeRiskActionRules(&rules)
	require.NoError(t, validateRiskActionRules(rules))
	result, err := ParseQwen3Guard("Safety: Controversial\nCategories: Jailbreak", AllScannerIDs)
	require.NoError(t, err)
	ApplyRiskPolicy(result, rules, PolicyMatchContext{GroupID: &groupID, Model: "Qwen", Provider: "OpenAI"})
	require.Equal(t, ActionBlock, result.Action)
	require.Equal(t, RiskCritical, result.RiskLevel)
	require.Equal(t, EventCritical, result.Decision)
	require.Equal(t, "ops-policy", result.PolicyID)
	require.Equal(t, 9, result.PolicyVersion)
	// Jailbreak is already a parser-level critical baseline; a matching rule that
	// does not further escalate it must not be claimed as the explanation.
	require.Empty(t, result.MatchedRuleID)
}

func TestRiskPolicyMatrixScopeMissDoesNotApplyRule(t *testing.T) {
	groupID := int64(8)
	rules := RiskActionRules{Rules: []RiskPolicyRule{{ID: "group-block", Groups: []int64{7}, Action: ActionBlock}}}
	normalizeRiskActionRules(&rules)
	require.NoError(t, validateRiskActionRules(rules))
	result, err := ParseQwen3Guard("Safety: Controversial\nCategories: Violent", AllScannerIDs)
	require.NoError(t, err)
	ApplyRiskPolicy(result, rules, PolicyMatchContext{GroupID: &groupID})
	require.Equal(t, ActionWarn, result.Action)
	require.Empty(t, result.MatchedRuleID)
}

func TestRiskPolicyMatrixRejectsWeakeningDefaultsAndDuplicateRules(t *testing.T) {
	cases := []RiskActionRules{
		{Defaults: map[string]RiskPolicyAction{"unsafe": {Action: ActionWarn}}},
		{Rules: []RiskPolicyRule{{ID: "same", Action: ActionBlock}, {ID: "same", Action: ActionBlock}}},
		{Rules: []RiskPolicyRule{{ID: "bad-risk", Action: ActionWarn, RiskLevel: RiskLow}}},
	}
	for _, rules := range cases {
		normalizeRiskActionRules(&rules)
		require.Error(t, validateRiskActionRules(rules))
	}
}

func TestAggregateResultsPreservesMatchedRuleID(t *testing.T) {
	result, err := AggregateResults([]*NormalizedResult{
		{Decision: EventPass, RiskLevel: RiskLow, Action: ActionAllow, PolicyID: "policy", PolicyVersion: 2},
		{Decision: EventCritical, RiskLevel: RiskCritical, Action: ActionBlock, MatchedRuleID: "jailbreak-block", PolicyID: "policy", PolicyVersion: 2},
	}, 0)
	require.NoError(t, err)
	require.Equal(t, "jailbreak-block", result.MatchedRuleID)
}

func TestPolicyShadowEvaluateIsMonotonicAndAddsOWASPMetadata(t *testing.T) {
	result, err := ParseQwen3Guard("Safety: Controversial\nCategories: Violent", AllScannerIDs)
	require.NoError(t, err)
	original := result.Action
	shadow, err := ShadowEvaluate(*result, RiskActionRules{Rules: []RiskPolicyRule{{ID: "owasp-violent", Priority: 10, Categories: []string{"violent"}, Action: ActionBlock, RiskLevel: RiskCritical, OWASP: []string{"LLM01"}}}}, PolicyMatchContext{})
	require.NoError(t, err)
	require.Equal(t, original, result.Action)
	require.True(t, shadow.ActionChanged)
	require.True(t, shadow.WouldEscalate)
	require.Equal(t, ActionBlock, shadow.Candidate.Action)
	require.Equal(t, []string{"LLM01"}, shadow.Candidate.OWASPTags)
}

func TestRiskPolicyValidationRejectsUnknownActionsAndBounds(t *testing.T) {
	tooManyTags := make([]string, maxOWASPTagsPerRule+1)
	for i := range tooManyTags {
		tooManyTags[i] = "LLM" + fmt.Sprintf("%02d", i+1)
	}
	cases := []RiskActionRules{
		{Rules: []RiskPolicyRule{{ID: "bad", Action: Action("Bogus"), RiskLevel: RiskCritical}}},
		{Rules: []RiskPolicyRule{{ID: strings.Repeat("x", 129), Action: ActionBlock}}},
		{Rules: []RiskPolicyRule{{ID: "tags", Action: ActionBlock, OWASP: tooManyTags}}},
	}
	for _, rules := range cases {
		normalizeRiskActionRules(&rules)
		require.Error(t, validateRiskActionRules(rules))
	}
}

func TestRiskPolicyUnsafeFloorAppliesWithoutRules(t *testing.T) {
	result, err := ParseQwen3Guard("Safety: Unsafe\nCategories: Violent", []string{"pii"})
	require.NoError(t, err)
	result.Action, result.Decision, result.RiskLevel = ActionAllow, EventPass, RiskLow
	ApplyRiskPolicy(result, RiskActionRules{}, PolicyMatchContext{})
	require.Equal(t, ActionBlock, result.Action)
	require.Equal(t, EventCritical, result.Decision)
	require.Equal(t, RiskCritical, result.RiskLevel)
}

func TestRiskPolicyMatchedRuleExplainsEscalationOnly(t *testing.T) {
	result, err := ParseQwen3Guard("Safety: Controversial\nCategories: Violent", AllScannerIDs)
	require.NoError(t, err)
	rules := RiskActionRules{Rules: []RiskPolicyRule{
		{ID: "warn", Priority: 100, Action: ActionWarn, RiskLevel: RiskMedium},
		{ID: "block", Priority: 1, Action: ActionBlock, RiskLevel: RiskCritical},
	}}
	normalizeRiskActionRules(&rules)
	require.NoError(t, validateRiskActionRules(rules))
	ApplyRiskPolicy(result, rules, PolicyMatchContext{})
	require.Equal(t, ActionBlock, result.Action)
	require.Equal(t, "block", result.MatchedRuleID)
}
