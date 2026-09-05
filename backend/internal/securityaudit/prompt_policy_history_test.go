package securityaudit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPolicyHistoryNormalizesBoundsAndDropsInvalidSnapshots(t *testing.T) {
	raw, err := json.Marshal(PolicyHistory{ActiveVersion: 99, Versions: []PolicyVersionRecord{
		{PolicyVersion: 2, PolicyID: " keep ", Rules: RiskActionRules{Defaults: map[string]RiskPolicyAction{
			"safe":   {Action: ActionAllow, RiskLevel: RiskLow},
			"unsafe": {Action: ActionBlock, RiskLevel: RiskCritical},
		}}},
		{PolicyVersion: 0, PolicyID: "invalid"},
		{PolicyVersion: 2, PolicyID: "duplicate"},
	}})
	require.NoError(t, err)
	history, err := parsePolicyHistory(string(raw))
	require.NoError(t, err)
	require.Len(t, history.Versions, 1)
	require.Equal(t, 2, history.Versions[0].PolicyVersion)
	require.Equal(t, "keep", history.Versions[0].PolicyID)
	require.Equal(t, 0, history.ActiveVersion)
}

func TestAppendPolicyVersionIsBoundedAndDoesNotStorePromptData(t *testing.T) {
	history := defaultPolicyHistory()
	for version := 1; version <= maxPromptPolicyHistory+4; version++ {
		history = appendPolicyVersion(history, RiskActionRules{
			PolicyID: "audit-policy", PolicyVersion: version,
			Defaults: map[string]RiskPolicyAction{"unsafe": {Action: ActionBlock, RiskLevel: RiskCritical}},
		}, int64(version), 42, time.Unix(int64(version), 0))
	}
	require.Len(t, history.Versions, maxPromptPolicyHistory)
	require.Equal(t, maxPromptPolicyHistory+4, history.ActiveVersion)
	raw, err := marshalPolicyHistory(history)
	require.NoError(t, err)
	require.NotContains(t, raw, "prompt")
	require.NotContains(t, raw, "token")
}

func TestPolicyHistoryClearsActiveVersionOutsideRetainedWindow(t *testing.T) {
	history := PolicyHistory{ActiveVersion: 1}
	for version := 1; version <= maxPromptPolicyHistory+1; version++ {
		history.Versions = append(history.Versions, PolicyVersionRecord{PolicyVersion: version, PolicyID: "test-policy"})
	}
	raw, err := json.Marshal(history)
	require.NoError(t, err)
	parsed, err := parsePolicyHistory(string(raw))
	require.NoError(t, err)
	require.Len(t, parsed.Versions, maxPromptPolicyHistory)
	require.Zero(t, parsed.ActiveVersion)
	_, found := findPolicyVersion(parsed, 1)
	require.False(t, found)
	history.ActiveVersion = maxPromptPolicyHistory + 1
	raw, err = json.Marshal(history)
	require.NoError(t, err)
	parsed, err = parsePolicyHistory(string(raw))
	require.NoError(t, err)
	require.Equal(t, maxPromptPolicyHistory+1, parsed.ActiveVersion)
}

func TestListPolicyVersionsTreatsMissingSettingAsEmpty(t *testing.T) {
	manager := &ConfigManager{settings: policyHistorySettingStub{err: service.ErrSettingNotFound}}
	history, err := manager.ListPolicyVersions(context.Background())
	require.NoError(t, err)
	require.Empty(t, history.Versions)
}

func TestBuildPolicyPreviewIncludesRedactedShadowExamples(t *testing.T) {
	preview, err := buildPolicyPreview(RiskActionRules{
		PolicyID: "preview-policy",
		Rules:    []RiskPolicyRule{{ID: "jailbreak-block", Priority: 100, Categories: []string{"jailbreak"}, Action: ActionBlock, RiskLevel: RiskCritical, OWASP: []string{"LLM01"}}},
	})
	require.NoError(t, err)
	require.Len(t, preview.Examples, 4)
	var jailbreak PolicyPreviewExample
	for _, example := range preview.Examples {
		if example.Name == "controversial_jailbreak" {
			jailbreak = example
		}
	}
	require.Equal(t, ActionBlock, jailbreak.CandidateAction)
	require.Empty(t, jailbreak.MatchedRuleID)
	require.Equal(t, []string{"LLM01"}, jailbreak.OWASPTags)
	require.False(t, jailbreak.WouldEscalate)
	raw, marshalErr := json.Marshal(preview)
	require.NoError(t, marshalErr)
	require.NotContains(t, string(raw), "prompt")
}

type policyHistorySettingStub struct{ err error }

func (s policyHistorySettingStub) Get(context.Context, string) (*service.Setting, error) {
	return nil, s.err
}
func (s policyHistorySettingStub) GetValue(context.Context, string) (string, error) { return "", s.err }
func (s policyHistorySettingStub) Set(context.Context, string, string) error        { return nil }
func (s policyHistorySettingStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	return nil, nil
}
func (s policyHistorySettingStub) SetMultiple(context.Context, map[string]string) error { return nil }
func (s policyHistorySettingStub) GetAll(context.Context) (map[string]string, error)    { return nil, nil }
func (s policyHistorySettingStub) Delete(context.Context, string) error                 { return nil }
