package securityaudit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPolicyHistoryValidatesAuthoritativeRecordID(t *testing.T) {
	for _, id := range []string{strings.Repeat("a", 129), strings.Repeat("\u754c", 129), "  "} {
		raw, err := json.Marshal(PolicyHistory{Versions: []PolicyVersionRecord{{PolicyVersion: 1, PolicyID: id}}})
		require.NoError(t, err)
		history, err := parsePolicyHistory(string(raw))
		require.NoError(t, err)
		require.Empty(t, history.Versions)
	}
	id := strings.Repeat("\u754c", 128)
	raw, err := json.Marshal(PolicyHistory{Versions: []PolicyVersionRecord{{PolicyVersion: 1, PolicyID: id}}})
	require.NoError(t, err)
	history, err := parsePolicyHistory(string(raw))
	require.NoError(t, err)
	require.Len(t, history.Versions, 1)
	require.Equal(t, id, history.Versions[0].Rules.PolicyID)
}

func TestPolicyAttributionExcludesDefaultAndMapBaseline(t *testing.T) {
	for _, rules := range []RiskActionRules{
		{Defaults: map[string]RiskPolicyAction{"controversial": {Action: ActionBlock}}},
		{Categories: map[string]Action{"violent": ActionBlock}},
		{Safety: map[string]Action{"controversial": ActionBlock}},
	} {
		result, err := ParseQwen3Guard("Safety: Controversial\nCategories: Violent", AllScannerIDs)
		require.NoError(t, err)
		rules.Rules = []RiskPolicyRule{{ID: "not-a-contributor", Priority: 100, Action: ActionBlock, OWASP: []string{"LLM01"}}}
		ApplyRiskPolicy(result, rules, PolicyMatchContext{})
		require.Equal(t, ActionBlock, result.Action)
		require.Empty(t, result.MatchedRuleID)
		require.Equal(t, []string{"LLM01"}, result.OWASPTags)
	}
}

func TestPolicyAggregateBreaksTiesByContributionPriorityAndID(t *testing.T) {
	makeResult := func(id string, priority int) *NormalizedResult {
		result, err := ParseQwen3Guard("Safety: Controversial\nCategories: Violent", AllScannerIDs)
		require.NoError(t, err)
		ApplyRiskPolicy(result, RiskActionRules{Rules: []RiskPolicyRule{{ID: id, Priority: priority, Action: ActionBlock}}}, PolicyMatchContext{})
		return result
	}
	for _, pair := range [][]*NormalizedResult{
		{makeResult("z", 1), makeResult("a", 100)},
		{makeResult("a", 100), makeResult("z", 1)},
		{makeResult("z", 100), makeResult("a", 100)},
		{makeResult("a", 100), makeResult("z", 100)},
	} {
		result, err := AggregateResults(pair, 0)
		require.NoError(t, err)
		require.Equal(t, "a", result.MatchedRuleID)
		raw, err := json.Marshal(result)
		require.NoError(t, err)
		require.NotContains(t, string(raw), "Priority")
		require.NotContains(t, string(raw), "priority\":100")
	}
}

func TestPolicyAggregateDoesNotBorrowWeakerAttribution(t *testing.T) {
	strong := &NormalizedResult{Action: ActionBlock, RiskLevel: RiskCritical, Decision: EventCritical}
	weak := &NormalizedResult{Action: ActionWarn, RiskLevel: RiskHigh, Decision: EventFlag, MatchedRuleID: "warning"}
	for _, pair := range [][]*NormalizedResult{{strong, weak}, {weak, strong}} {
		result, err := AggregateResults(pair, 0)
		require.NoError(t, err)
		require.Equal(t, ActionBlock, result.Action)
		require.Empty(t, result.MatchedRuleID)
	}
	result, err := AggregateResults([]*NormalizedResult{
		{Action: ActionWarn, RiskLevel: RiskMedium, Decision: EventFlag, MatchedRuleID: "medium"},
		{Action: ActionWarn, RiskLevel: RiskHigh, Decision: EventFlag, MatchedRuleID: "high"},
	}, 0)
	require.NoError(t, err)
	require.Equal(t, RiskHigh, result.RiskLevel)
	require.Equal(t, "high", result.MatchedRuleID)
}
