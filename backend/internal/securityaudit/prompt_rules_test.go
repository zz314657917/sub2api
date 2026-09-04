package securityaudit

import (
	"encoding/json"
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
