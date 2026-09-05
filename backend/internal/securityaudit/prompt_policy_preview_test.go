package securityaudit

import (
	"context"
	"net/http"
	"strings"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestGuardPolicyComparisonUsesIndependentBaseline(t *testing.T) {
	group := int64(42)
	rules := RiskActionRules{Rules: []RiskPolicyRule{{
		ID: "scoped", Categories: []string{"violent"}, Groups: []int64{group},
		Models: []string{"test-model"}, Providers: []string{"test-provider"},
		Action: ActionBlock, RiskLevel: RiskCritical, OWASP: []string{"LLM01"},
	}}}
	active := ActiveConfig{Scanners: AllScannerIDs, Rules: rules}
	match := PolicyMatchContext{GroupID: &group, Model: "test-model", Provider: "test-provider"}
	output := "Safety: Controversial\nCategories: Violent"
	result, err := compareGuardPolicies(output, active, RiskActionRules{}, match)
	require.NoError(t, err)
	require.Equal(t, ActionBlock, result.Current.Action)
	require.Equal(t, "scoped", result.Current.MatchedRuleID)
	require.Equal(t, ActionWarn, result.Candidate.Action)
	require.Empty(t, result.Candidate.MatchedRuleID)
	require.Empty(t, result.Candidate.OWASPTags)
	require.False(t, result.WouldEscalate)
	require.True(t, result.ActionChanged)

	active.Rules = RiskActionRules{}
	result, err = compareGuardPolicies(output, active, rules, match)
	require.NoError(t, err)
	require.Equal(t, ActionWarn, result.Current.Action)
	require.Equal(t, ActionBlock, result.Candidate.Action)
	require.True(t, result.WouldEscalate)
	match.Provider = "other-provider"
	result, err = compareGuardPolicies(output, active, rules, match)
	require.NoError(t, err)
	require.Equal(t, ActionWarn, result.Candidate.Action)
	require.False(t, result.ActionChanged)
}

func TestGuardPolicyComparisonUsesActiveScannersAndFloors(t *testing.T) {
	for _, tc := range []struct {
		name, output string
		scanners     []string
		action       Action
	}{
		{"enabled jailbreak", "Safety: Controversial\nCategories: Jailbreak", AllScannerIDs, ActionBlock},
		{"disabled jailbreak", "Safety: Controversial\nCategories: Jailbreak", []string{"violent"}, ActionWarn},
		{"disabled unsafe", "Safety: Unsafe\nCategories: Jailbreak", []string{"violent"}, ActionBlock},
		{"unknown", "Safety: Safe\nCategories: future-category", nil, ActionBlock},
		{"safe", "Safety: Safe\nCategories: None", AllScannerIDs, ActionAllow},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := compareGuardPolicies(tc.output, ActiveConfig{Scanners: tc.scanners}, RiskActionRules{}, PolicyMatchContext{})
			require.NoError(t, err)
			require.Equal(t, tc.action, result.Current.Action)
			require.Equal(t, tc.action, result.Candidate.Action)
			require.False(t, result.ActionChanged)
		})
	}
}

func TestShadowPolicyGuardOutputValidationAndLegacy(t *testing.T) {
	service := &PromptService{config: &fakeConfigStore{active: true, cfg: ActiveConfig{Scanners: AllScannerIDs}}}
	for _, output := range []string{"", " ", "synthetic-private-canary", strings.Repeat("x", 4097)} {
		_, err := service.ShadowPolicy(context.Background(), PolicyShadowRequest{GuardOutput: &output})
		require.Equal(t, 400, infraerrors.Code(err))
		require.NotContains(t, err.Error(), "synthetic-private-canary")
	}
	output := "Safety: Controversial\nCategories: Jailbreak"
	result, err := service.ShadowPolicy(context.Background(), PolicyShadowRequest{
		GuardOutput: &output, CurrentResult: NormalizedResult{Action: ActionAllow},
	})
	require.NoError(t, err)
	require.Equal(t, ActionBlock, result.Current.Action)
	legacy := NormalizedResult{Safety: "Controversial", Categories: []string{"jailbreak"}, Action: ActionWarn, RiskLevel: RiskMedium, Decision: EventFlag}
	result, err = service.ShadowPolicy(context.Background(), PolicyShadowRequest{CurrentResult: legacy})
	require.NoError(t, err)
	require.Equal(t, legacy, result.Current)
	require.Equal(t, ActionWarn, result.Candidate.Action)
}

func TestPolicyPreviewsRequireActiveSnapshot(t *testing.T) {
	output := "Safety: Safe\nCategories: None"
	service := &PromptService{config: &fakeConfigStore{}}
	_, err := service.ShadowPolicy(context.Background(), PolicyShadowRequest{GuardOutput: &output})
	require.Equal(t, 503, infraerrors.Code(err))
	manager := &ConfigManager{}
	_, err = manager.PreviewPolicy(context.Background(), RiskActionRules{})
	require.Equal(t, 503, infraerrors.Code(err))
}

func TestPolicyPreviewShowsRemovalOfActiveEscalation(t *testing.T) {
	active := ActiveConfig{Scanners: AllScannerIDs, Rules: RiskActionRules{Rules: []RiskPolicyRule{{
		ID: "violent-block", Categories: []string{"violent"}, Action: ActionBlock,
	}}}}
	manager := &ConfigManager{}
	manager.snapshot.Store(&activeConfigSnapshot{active: active})
	preview, err := manager.PreviewPolicy(context.Background(), RiskActionRules{})
	require.NoError(t, err)
	for _, example := range preview.Examples {
		if example.Name == "controversial_violent" {
			require.Equal(t, ActionBlock, example.CurrentAction)
			require.Equal(t, ActionWarn, example.CandidateAction)
			require.Empty(t, example.MatchedRuleID)
			return
		}
	}
	t.Fatal("missing violent preview example")
}

func TestShadowPolicyHTTPAcceptsGuardOutputWithoutLegacyResult(t *testing.T) {
	service := &PromptService{config: &fakeConfigStore{active: true, cfg: ActiveConfig{Scanners: AllScannerIDs}}}
	router := promptAdminRouter(service)
	response := promptAdminRequest(t, router, http.MethodPost, "/admin/prompt-audit/policy/shadow", map[string]any{
		"guard_output": "Safety: Controversial\nCategories: Jailbreak", "rules": map[string]any{},
	})
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Contains(t, response.Body.String(), `"action":"Block"`)
	response = promptAdminRequest(t, router, http.MethodPost, "/admin/prompt-audit/policy/shadow", map[string]any{
		"guard_output": "private-invalid-output", "rules": map[string]any{},
	})
	require.Equal(t, http.StatusBadRequest, response.Code)
	require.NotContains(t, response.Body.String(), "private-invalid-output")
}
