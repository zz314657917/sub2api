package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

type codexManifestHTTPStub struct {
	body    string
	status  int
	etag    string
	lastReq *http.Request
}

func (s *codexManifestHTTPStub) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	s.lastReq = req
	status := s.status
	if status == 0 {
		status = http.StatusOK
	}
	header := make(http.Header)
	if s.etag != "" {
		header.Set("ETag", s.etag)
	}
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(s.body)),
	}, nil
}

func TestFetchCodexModelsManifestAPIKeyFallsBackForStaleHeaderVersion(t *testing.T) {
	upstream := &codexManifestHTTPStub{body: `{"models":[{"slug":"gpt-5.6-sol"}]}`}
	service := &OpenAIGatewayService{
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
		httpUpstream: upstream,
	}
	account := &Account{ID: 4, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://provider.example/v1"}}

	manifest, err := service.FetchCodexModelsManifest(context.Background(), account, "0.125.0", "")

func (r splitCodexModelsAccountRepo) ListSchedulableByGroupID(_ context.Context, groupID int64) ([]Account, error) {
	return append([]Account(nil), r.schedulable[groupID]...), nil
}

func (r splitCodexModelsAccountRepo) ListByGroup(_ context.Context, groupID int64) ([]Account, error) {
	accounts := r.all[groupID]
	if accounts == nil {
		accounts = r.catalog[groupID]
	}
	return append([]Account(nil), accounts...), nil
}

func (r splitCodexModelsAccountRepo) ListModelAvailabilityCandidates(_ context.Context, groupID *int64, _ []string, _ bool) ([]Account, error) {
	if groupID == nil {
		return nil, nil
	}
	return append([]Account(nil), r.catalog[*groupID]...), nil
}

func newCodexCatalogMappedAccount(
	id int64,
	target string,
	displayName string,
	levels []string,
	modalities []string,
	contextWindow int64,
	schedulable bool,
	extraMapping map[string]any,
) Account {
	reasoning := true
	mapping := map[string]any{"my-coder": target}
	for key, value := range extraMapping {
		mapping[key] = value
	}
	account := Account{
		ID:          id,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Schedulable: schedulable,
		Credentials: map[string]any{
			"base_url":      fmt.Sprintf("https://provider-%d.example/v1", id),
			"model_mapping": mapping,
		},
	}
	models := map[string]UpstreamModelMetadata{
		target: {
			ID:                       target,
			DisplayName:              displayName,
			Description:              displayName + " upstream",
			Reasoning:                &reasoning,
			SupportedReasoningLevels: levels,
			InputModalities:          modalities,
			ContextWindow:            contextWindow,
		},
	}
	for _, value := range extraMapping {
		exclusive, _ := value.(string)
		if exclusive == "" || exclusive == target {
			continue
		}
		models[exclusive] = UpstreamModelMetadata{
			ID:                       exclusive,
			DisplayName:              "Exclusive Model",
			Description:              "Only mapped on the unschedulable account",
			Reasoning:                &reasoning,
			SupportedReasoningLevels: []string{"high"},
			InputModalities:          []string{"text", "image"},
			ContextWindow:            1_000_000,
		}
	}
	account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: models})
	return account
}

func TestFilterCodexModelIDsForGroupOmitsWildcardKeys(t *testing.T) {
	t.Parallel()

	got := FilterCodexModelIDsForGroup(
		[]string{"deepseek-v4-pro", "foo-*", "  bar-*  ", "gpt-5.5"},
		&Group{Platform: PlatformDeepseek},
	)
	require.Equal(t, []string{"deepseek-v4-pro", "gpt-5.5"}, got)
}

func decodeCodexManifestModels(t *testing.T, body []byte) []map[string]any {
	t.Helper()

	var envelope struct {
		Models []map[string]any `json:"models"`
	}
	require.NoError(t, json.Unmarshal(body, &envelope))
	return envelope.Models
}

func codexManifestModelSlugs(t *testing.T, body []byte) []string {
	t.Helper()

	models := decodeCodexManifestModels(t, body)
	slugs := make([]string, 0, len(models))
	for _, model := range models {
		slug, ok := model["slug"].(string)
		require.True(t, ok)
		slugs = append(slugs, slug)
	}
	return slugs
}

func requireCompleteConfiguredCodexModel(t *testing.T, model map[string]any, slug string) {
	t.Helper()

	require.Equal(t, slug, model["slug"])
	require.NotEmpty(t, model["display_name"])
	require.NotEmpty(t, model["description"])
	require.Equal(t, "unified_exec", model["shell_type"])
	require.Equal(t, "list", model["visibility"])
	require.Equal(t, true, model["supported_in_api"])
	require.NotNil(t, model["priority"])
	require.Equal(t, []any{}, model["additional_speed_tiers"])
	require.IsType(t, []any{}, model["service_tiers"])
	require.Contains(t, model, "default_service_tier")
	require.Contains(t, model, "availability_nux")
	require.Contains(t, model, "upgrade")
	require.Contains(t, model, "default_verbosity")
	require.Contains(t, model, "apply_patch_tool_type")
	require.Contains(t, model, "auto_compact_token_limit")
	require.Contains(t, model, "comp_hash")
	require.Contains(t, model, "auto_review_model_override")
	require.Contains(t, model, "model_specialty")
	require.Contains(t, model, "tool_mode")
	require.Contains(t, model, "multi_agent_version")
	require.Equal(t, true, model["supports_reasoning_summary_parameter"])
	require.Contains(t, model, "include_skills_usage_instructions")
	require.Contains(t, model, "include_plugin_usage_instructions")
	require.Contains(t, model, "include_apps_usage_instructions")
	require.Contains(t, model, "supports_image_detail_original")
	require.Contains(t, model, "node_repl_auto_review_required")
	require.Contains(t, model, "node_repl_disabled")
	require.Contains(t, model, "truncation_policy")
	require.Contains(t, model, "supports_parallel_tool_calls")
	require.Contains(t, model, "experimental_supported_tools")
	modelMessages, ok := model["model_messages"].(map[string]any)
	require.True(t, ok)
	require.NotEmpty(t, modelMessages["instructions_template"])
	for _, key := range []string{
		"instructions_variables",
		"approvals",
		"collaboration_modes",
		"auto_review",
		"permissions",
		"multi_agent",
		"token_budget",
		"guardian_v2",
	} {
		require.Contains(t, modelMessages, key)
	}
}

func effortsFromManifestModel(t *testing.T, model map[string]any) []string {
	t.Helper()

	levels, ok := model["supported_reasoning_levels"].([]any)
	require.True(t, ok)
	efforts := make([]string, 0, len(levels))
	for _, rawLevel := range levels {
		level, ok := rawLevel.(map[string]any)
		require.True(t, ok)
		effort, ok := level["effort"].(string)
		require.True(t, ok)
		efforts = append(efforts, effort)
	}
	return efforts
}

func TestNewConfiguredCodexModelDescriptorUsesProviderMetadataAndSafeFallback(t *testing.T) {
	t.Parallel()

	deepSeek := newConfiguredCodexModelDescriptor("deepseek-v4-pro")
	require.Equal(t, "DeepSeek V4 Pro", deepSeek.DisplayName)
	require.Equal(t, int64(1_000_000), deepSeek.ContextWindow)
	require.Equal(t, int64(1_000_000), deepSeek.MaxContextWindow)
	require.NotNil(t, deepSeek.DefaultReasoningLevel)
	require.Equal(t, "high", *deepSeek.DefaultReasoningLevel)
	require.Equal(t, []configuredCodexReasoningLevel{
		{Effort: "low", Description: "Fast responses with lighter reasoning"},
		{Effort: "high", Description: "Greater reasoning depth for coding and agent tasks"},
		{Effort: "max", Description: "Maximum reasoning depth for complex tasks"},
	}, deepSeek.SupportedReasoningLevels)
	require.True(t, deepSeek.SupportsParallelToolCalls)
	require.Equal(t, []string{"text"}, deepSeek.InputModalities)

	grok := newConfiguredCodexModelDescriptor("grok-4.6")
	require.Equal(t, "Grok 4.6", grok.DisplayName)
	require.Equal(t, int64(500_000), grok.ContextWindow)
	require.Equal(t, int64(500_000), grok.MaxContextWindow)
	require.NotNil(t, grok.DefaultReasoningLevel)
	require.Equal(t, "high", *grok.DefaultReasoningLevel)
	require.Equal(t, []string{"low", "medium", "high", "xhigh"}, effortsFromConfiguredCodexLevels(grok.SupportedReasoningLevels))
	require.True(t, grok.SupportsParallelToolCalls)
	require.Equal(t, []string{"text"}, grok.InputModalities)
	require.NotContains(t, grok.SupportedReasoningLevels, configuredCodexReasoningLevel{Effort: "none"})
	require.NotContains(t, grok.SupportedReasoningLevels, configuredCodexReasoningLevel{Effort: "max"})

	grokAlias := newConfiguredCodexModelDescriptor("xai/grok-4.6-latest")
	require.Equal(t, "Grok 4.6", grokAlias.DisplayName)
	require.NotNil(t, grokAlias.DefaultReasoningLevel)
	require.Equal(t, "high", *grokAlias.DefaultReasoningLevel)
	require.Equal(t, []string{"low", "medium", "high", "xhigh"}, effortsFromConfiguredCodexLevels(grokAlias.SupportedReasoningLevels))

	grok45 := newConfiguredCodexModelDescriptor("grok-4.5")
	require.Equal(t, []string{"low", "medium", "high"}, effortsFromConfiguredCodexLevels(grok45.SupportedReasoningLevels))

	grokNonReasoning := newConfiguredCodexModelDescriptor("grok-4.20-0309-non-reasoning")
	require.Equal(t, "Grok 4.20 Non Reasoning", grokNonReasoning.DisplayName)
	require.NotNil(t, grokNonReasoning.DefaultReasoningLevel)
	require.Equal(t, "none", *grokNonReasoning.DefaultReasoningLevel)
	require.Equal(t, []configuredCodexReasoningLevel{
		{Effort: "none", Description: configuredCodexReasoningLevelDescription("none")},
	}, grokNonReasoning.SupportedReasoningLevels)

	claude := newConfiguredCodexModelDescriptor("claude-opus-4-6")
	require.Equal(t, "Claude Opus 4.6", claude.DisplayName)
	require.NotNil(t, claude.DefaultReasoningLevel)
	require.Equal(t, "medium", *claude.DefaultReasoningLevel)
	require.Equal(t, []string{"low", "medium", "high", "max"}, effortsFromConfiguredCodexLevels(claude.SupportedReasoningLevels))
	require.NotContains(t, claude.SupportedReasoningLevels, configuredCodexReasoningLevel{Effort: "xhigh"})
	require.NotContains(t, claude.SupportedReasoningLevels, configuredCodexReasoningLevel{Effort: "none"})
	require.True(t, claude.SupportsParallelToolCalls)

	claudeOpus5 := newConfiguredCodexModelDescriptor("claude-opus-5")
	require.Equal(t, []string{"low", "medium", "high", "xhigh", "max"}, effortsFromConfiguredCodexLevels(claudeOpus5.SupportedReasoningLevels))

	providerQualifiedClaude := newConfiguredCodexModelDescriptor("anthropic/claude-sonnet-4-6")
	require.Equal(t, "Claude Sonnet 4.6", providerQualifiedClaude.DisplayName)
	require.Equal(t, []string{"low", "medium", "high", "max"}, effortsFromConfiguredCodexLevels(providerQualifiedClaude.SupportedReasoningLevels))

	claudeHaiku := newConfiguredCodexModelDescriptor("claude-haiku-4-5-20251001")
	require.Equal(t, "Claude Haiku 4.5", claudeHaiku.DisplayName)
	require.NotNil(t, claudeHaiku.DefaultReasoningLevel)
	require.Equal(t, "none", *claudeHaiku.DefaultReasoningLevel)
	require.Equal(t, []string{"none"}, effortsFromConfiguredCodexLevels(claudeHaiku.SupportedReasoningLevels))

	gpt56 := newConfiguredCodexModelDescriptor("gpt-5.6-sol")
	require.Equal(t, "GPT-5.6 Sol", gpt56.DisplayName)
	require.Equal(t, "OpenAI GPT coding model routed through Sub2API.", gpt56.Description)
	require.NotNil(t, gpt56.DefaultReasoningLevel)
	require.Equal(t, "low", *gpt56.DefaultReasoningLevel)
	require.Equal(t, configuredCodexGPTReasoningLevels("gpt-5.6-sol"), gpt56.SupportedReasoningLevels)
	require.Equal(t, []string{"low", "medium", "high", "xhigh", "max", "ultra"}, effortsFromConfiguredCodexLevels(gpt56.SupportedReasoningLevels))
	require.True(t, gpt56.SupportsParallelToolCalls)
	require.True(t, gpt56.SupportVerbosity)
	require.Equal(t, []string{"text"}, gpt56.InputModalities)
	require.Equal(t, int64(872_000), gpt56.MaxContextWindow)
	require.Equal(t, configuredCodexTruncationPolicy{Mode: "tokens", Limit: 10_000}, gpt56.TruncationPolicy)
	require.NotNil(t, gpt56.DefaultVerbosity)
	require.Equal(t, "low", *gpt56.DefaultVerbosity)
	require.True(t, gpt56.SupportsReasoningSummaryParameter)
	require.Equal(t, "none", gpt56.DefaultReasoningSummary)

	gpt56Luna := newConfiguredCodexModelDescriptor("gpt-5.6-luna")
	require.Equal(t, []string{"low", "medium", "high", "xhigh", "max"}, effortsFromConfiguredCodexLevels(gpt56Luna.SupportedReasoningLevels))
	require.Equal(t, "medium", *gpt56Luna.DefaultReasoningLevel)

	astra := newConfiguredCodexModelDescriptor("gpt-6-astra")
	require.Equal(t, "GPT-6 Astra", astra.DisplayName)
	require.NotNil(t, astra.DefaultReasoningLevel)
	require.Equal(t, "medium", *astra.DefaultReasoningLevel)
	require.Equal(t, []string{"low", "medium", "high", "xhigh", "max"}, effortsFromConfiguredCodexLevels(astra.SupportedReasoningLevels))
	require.NotContains(t, astra.SupportedReasoningLevels, configuredCodexReasoningLevel{Effort: "none"})
	require.NotContains(t, astra.SupportedReasoningLevels, configuredCodexReasoningLevel{Effort: "minimal"})
	require.Len(t, astra.ServiceTiers, 1)
	require.Equal(t, "priority", astra.ServiceTiers[0].ID)
	require.True(t, astra.SupportsParallelToolCalls)
	require.True(t, astra.SupportVerbosity)
	require.Equal(t, int64(1_050_000), astra.ContextWindow)
	require.Equal(t, int64(1_050_000), astra.MaxContextWindow)
	require.Equal(t, configuredCodexTruncationPolicy{Mode: "tokens", Limit: 10_000}, astra.TruncationPolicy)
	gpt6 := newConfiguredCodexModelDescriptor("gpt-6")
	require.Equal(t, "GPT-6 (Astra)", gpt6.DisplayName)
	require.Equal(t, []string{"low", "medium", "high", "xhigh", "max"}, effortsFromConfiguredCodexLevels(gpt6.SupportedReasoningLevels))
	require.Equal(t, int64(1_050_000), gpt6.ContextWindow)

	gpt55 := newConfiguredCodexModelDescriptor("gpt-5.5")
	require.Equal(t, "GPT-5.5", gpt55.DisplayName)
	require.NotNil(t, gpt55.DefaultReasoningLevel)
	require.Equal(t, "medium", *gpt55.DefaultReasoningLevel)
	require.Equal(t, []string{"low", "medium", "high", "xhigh"}, effortsFromConfiguredCodexLevels(gpt55.SupportedReasoningLevels))
	require.NotContains(t, gpt55.SupportedReasoningLevels, configuredCodexReasoningLevel{Effort: "max"})
	require.NotNil(t, gpt55.DefaultVerbosity)
	require.Equal(t, "low", *gpt55.DefaultVerbosity)

	gpt4o := newConfiguredCodexModelDescriptor("gpt-4o")
	require.Equal(t, "gpt-4o", gpt4o.DisplayName)
	require.NotNil(t, gpt4o.DefaultReasoningLevel)
	require.Equal(t, "none", *gpt4o.DefaultReasoningLevel)
	require.Equal(t, []string{"none"}, effortsFromConfiguredCodexLevels(gpt4o.SupportedReasoningLevels))
	require.True(t, gpt4o.SupportsParallelToolCalls)

	image := newConfiguredCodexModelDescriptor("gpt-image-2")
	require.Equal(t, "gpt-image-2", image.DisplayName)
	require.NotNil(t, image.DefaultReasoningLevel)
	require.Equal(t, "none", *image.DefaultReasoningLevel)
	require.Equal(t, []string{"none"}, effortsFromConfiguredCodexLevels(image.SupportedReasoningLevels))

	custom := newConfiguredCodexModelDescriptor("company-coding-model")
	require.Equal(t, "company-coding-model", custom.DisplayName)
	require.Equal(t, int64(272_000), custom.ContextWindow)
	require.NotNil(t, custom.DefaultReasoningLevel)
	require.Equal(t, "none", *custom.DefaultReasoningLevel)
	require.Equal(t, []configuredCodexReasoningLevel{
		{Effort: "none", Description: configuredCodexReasoningLevelDescription("none")},
	}, custom.SupportedReasoningLevels)
	require.False(t, custom.SupportsParallelToolCalls)
	require.NotEmpty(t, custom.ModelMessages.InstructionsTemplate)
	require.Equal(t, "auto", custom.DefaultReasoningSummary)
	require.Equal(t, configuredCodexTruncationPolicy{Mode: "bytes", Limit: 10_000}, custom.TruncationPolicy)
}

func effortsFromConfiguredCodexLevels(levels []configuredCodexReasoningLevel) []string {
	efforts := make([]string, 0, len(levels))
	for _, level := range levels {
		efforts = append(efforts, level.Effort)
	}
	return efforts
}

// Scenario: 无推理模型可直接选中。
func TestBuildCodexModelsManifestUsesSingleNoneReasoningChoiceForCustomModel(t *testing.T) {
	t.Parallel()

	body, err := BuildCodexModelsManifest([]string{"company-coding-model"})
	require.NoError(t, err)
	require.NotNil(t, manifest)
	require.Equal(t, "0.125.0", upstream.lastReq.URL.Query().Get("client_version"))
	require.Equal(t, codexCLIVersion, upstream.lastReq.Header.Get("Version"))
}

// Scenario: 已知推理模型保留真实档位。
func TestBuildCodexModelsManifestKeepsKnownReasoningChoices(t *testing.T) {
	t.Parallel()

	body, err := BuildCodexModelsManifest([]string{"gpt-5.6-sol"})
	require.NoError(t, err)
	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 1)
	require.Equal(t, "low", models[0]["default_reasoning_level"])
	levels, ok := models[0]["supported_reasoning_levels"].([]any)
	require.True(t, ok)
	require.Len(t, levels, 6)
	firstLevel, ok := levels[0].(map[string]any)
	require.True(t, ok)
	require.NotEqual(t, "none", firstLevel["effort"])
}

// Scenario: 支持 Fast 的 GPT 型号在目录中声明 priority service tier。
func TestBuildCodexModelsManifestAdvertisesPriorityServiceTierForFastGPTModels(t *testing.T) {
	t.Parallel()

	body, err := BuildCodexModelsManifest([]string{
		"gpt-5.4-mini",
		"gpt-5.5",
		"gpt-5.6-terra",
	})
	require.NoError(t, err)
	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 3)

	for _, model := range models {
		require.Equal(t, []any{
			map[string]any{
				"id":          "priority",
				"name":        "Fast",
				"description": "Priority processing for lower latency.",
			},
		}, model["service_tiers"])
		require.Nil(t, model["default_service_tier"])
	}
}

// Scenario: GPT-5.6 Sol 在 Fast 之外额外声明 ultrafast service tier。
func TestBuildCodexModelsManifestAdvertisesUltrafastServiceTierForSol(t *testing.T) {
	t.Parallel()

	body, err := BuildCodexModelsManifest([]string{
		"gpt-5.6-sol",
		"gpt-5.6",
	})
	require.NoError(t, err)
	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 2)

	wantTiers := []any{
		map[string]any{
			"id":          "priority",
			"name":        "Fast",
			"description": "Priority processing for lower latency.",
		},
		map[string]any{
			"id":          "ultrafast",
			"name":        "Ultrafast",
			"description": "Ultra-low latency processing.",
		},
	}
	for _, model := range models {
		require.Equal(t, wantTiers, model["service_tiers"])
		require.Nil(t, model["default_service_tier"])
	}
}

// Scenario: 未明确支持 Fast 的型号不推测 service tier。
func TestBuildCodexModelsManifestLeavesServiceTiersEmptyForOtherModels(t *testing.T) {
	t.Parallel()

	body, err := BuildCodexModelsManifest([]string{
		"gpt-4o",
		"claude-opus-4-6",
		"deepseek-v4-pro",
		"company-coding-model",
	})
	require.NoError(t, err)
	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 4)

	for _, model := range models {
		require.Equal(t, []any{}, model["service_tiers"])
		require.Nil(t, model["default_service_tier"])
	}
}

// Scenario: 专用图片生成模型不进入 Codex 主模型目录。
func TestBuildCodexModelsManifestOmitsDedicatedImageModels(t *testing.T) {
	t.Parallel()

	body, err := BuildCodexModelsManifest([]string{
		"grok-4.6",
		"gpt-image-1",
		"gpt-image-1.5",
		"gpt-image-2",
		"openai/gpt-image-2",
		"gemini-2.5-flash-image",
		"gemini-3.1-flash-image-preview",
		"gemini-3-pro-image",
		"google/gemini-3-pro-image",
		"google/models/gemini-2.5-flash-image-preview",
		"grok-imagine-image",
		"grok-imagine-video",
		"xai/grok-imagine-image-quality",
		"grok-4.5",
	})
	require.NoError(t, err)
	models := decodeCodexManifestModels(t, body)
	slugs := make([]string, 0, len(models))
	for _, model := range models {
		slug, _ := model["slug"].(string)
		slugs = append(slugs, slug)
	}
	require.Equal(t, []string{"grok-4.6", "grok-4.5"}, slugs)
}

func TestBuildCodexModelsManifestForGroupAdvertisesOfficialGrokResponsesImageInput(t *testing.T) {
	t.Parallel()

	const groupID int64 = 701
	svc := &GatewayService{
		accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
			groupID: {{
				ID:       1,
				Platform: PlatformGrok,
				Type:     AccountTypeOAuth,
				Credentials: map[string]any{
					"access_token": "token",
				},
			}},
		}},
	}

	body, err := svc.BuildCodexModelsManifestForGroup(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformComposite},
		"",
		[]string{"grok-4.5"},
	)
	require.NoError(t, err)

	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 1)
	require.Equal(t, []any{"text", "image"}, models[0]["input_modalities"])
	require.Equal(t, "Grok 4.5", models[0]["display_name"])
	require.Equal(t, []string{"low", "medium", "high"}, effortsFromManifestModel(t, models[0]))
}

func TestBuildCodexModelsManifestForGroupAdvertisesOfficialOpenAIResponsesImageInput(t *testing.T) {
	t.Parallel()

	const groupID int64 = 702
	svc := &GatewayService{
		accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
			groupID: {{
				ID:       2,
				Platform: PlatformOpenAI,
				Type:     AccountTypeOAuth,
			}},
		}},
	}

	body, err := svc.BuildCodexModelsManifestForGroup(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformComposite},
		"",
		[]string{"gpt-5.6-sol"},
	)
	require.NoError(t, err)

	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 1)
	require.Equal(t, []any{"text", "image"}, models[0]["input_modalities"])
}

func TestBuildCodexModelsManifestForGroupUsesProviderImageCapabilities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		model      string
		accounts   []Account
		modalities []any
	}{
		{
			name:  "official Grok 4.6",
			model: "grok-4.6",
			accounts: []Account{{
				ID: 10, Platform: PlatformGrok, Type: AccountTypeOAuth,
			}},
			modalities: []any{"text", "image"},
		},
		{
			name:  "official Grok Build vision host",
			model: "grok-build-0.1",
			accounts: []Account{{
				ID: 11, Platform: PlatformGrok, Type: AccountTypeOAuth,
			}},
			modalities: []any{"text", "image"},
		},
		{
			name:  "official Grok 4.20 vision model",
			model: "grok-4.20-0309-reasoning",
			accounts: []Account{{
				ID: 23, Platform: PlatformGrok, Type: AccountTypeOAuth,
			}},
			modalities: []any{"text", "image"},
		},
		{
			name:  "Grok 3 Mini is text only",
			model: "grok-3-mini",
			accounts: []Account{{
				ID: 24, Platform: PlatformGrok, Type: AccountTypeOAuth,
			}},
			modalities: []any{"text"},
		},
		{
			name:  "Grok Composer has only Chat image bridge",
			model: "grok-composer-2.5-fast",
			accounts: []Account{{
				ID: 12, Platform: PlatformGrok, Type: AccountTypeOAuth,
			}},
			modalities: []any{"text"},
		},
		{
			name:  "custom Grok host",
			model: "grok-4.5",
			accounts: []Account{{
				ID: 13, Platform: PlatformGrok, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"base_url": "https://relay.example.test/v1"},
			}},
			modalities: []any{"text"},
		},
		{
			name:  "malformed Grok host",
			model: "grok-4.5",
			accounts: []Account{{
				ID: 19, Platform: PlatformGrok, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"base_url": "::invalid::url"},
			}},
			modalities: []any{"text"},
		},
		{
			name:  "mixed official and custom Grok candidates",
			model: "grok-4.5",
			accounts: []Account{
				{ID: 14, Platform: PlatformGrok, Type: AccountTypeOAuth},
				{
					ID: 15, Platform: PlatformGrok, Type: AccountTypeAPIKey,
					Credentials: map[string]any{"base_url": "https://relay.example.test/v1"},
				},
			},
			modalities: []any{"text"},
		},
		{
			name:  "DeepSeek V4",
			model: "deepseek-v4-pro",
			accounts: []Account{{
				ID: 16, Platform: PlatformDeepseek, Type: AccountTypeAPIKey,
			}},
			modalities: []any{"text"},
		},
		{
			name:  "official OpenAI API key",
			model: "gpt-5.6-sol",
			accounts: []Account{{
				ID: 17, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
			}},
			modalities: []any{"text", "image"},
		},
		{
			name:  "official OpenAI legacy text model",
			model: "gpt-3.5-turbo",
			accounts: []Account{{
				ID: 20, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
			}},
			modalities: []any{"text"},
		},
		{
			name:  "custom OpenAI-compatible host",
			model: "gpt-5.6-sol",
			accounts: []Account{{
				ID: 18, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"base_url": "https://openai-compatible.example.test/v1"},
			}},
			modalities: []any{"text", "image"},
		},
		{
			name:  "custom OpenAI-compatible host with unknown model",
			model: "company-coding-model",
			accounts: []Account{{
				ID: 29, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
				Credentials: map[string]any{"base_url": "https://openai-compatible.example.test/v1"},
			}},
			modalities: []any{"text"},
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			groupID := int64(710 + i)
			svc := &GatewayService{accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
				groupID: tt.accounts,
			}}}
			body, err := svc.BuildCodexModelsManifestForGroup(
				context.Background(),
				&Group{ID: groupID, Platform: PlatformComposite},
				"",
				[]string{tt.model},
			)
			require.NoError(t, err)
			models := decodeCodexManifestModels(t, body)
			require.Len(t, models, 1)
			require.Equal(t, tt.modalities, models[0]["input_modalities"])
		})
	}
}

func TestBuildCodexModelsManifestForGroupPrefersSyncedOpenAIImageCapabilities(t *testing.T) {
	t.Parallel()

	newAccount := func(id int64, modalities []string) Account {
		account := Account{
			ID: id, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
			Credentials: map[string]any{"base_url": "https://openai-compatible.example.test/v1"},
		}
		if modalities != nil {
			account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
				"gpt-5.6-sol": {ID: "gpt-5.6-sol", InputModalities: modalities},
			}})
		}
		return account
	}

	tests := []struct {
		name       string
		accounts   []Account
		modalities []any
	}{
		{
			name: "explicit text-only snapshot narrows local fallback",
			accounts: []Account{
				newAccount(30, nil),
				newAccount(31, []string{"text"}),
			},
			modalities: []any{"text"},
		},
		{
			name: "explicit multimodal snapshot preserves local fallback",
			accounts: []Account{
				newAccount(32, nil),
				newAccount(33, []string{"text", "image"}),
			},
			modalities: []any{"text", "image"},
		},
	}

	for i, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			groupID := int64(760 + i)
			svc := &GatewayService{accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
				groupID: tt.accounts,
			}}}
			body, err := svc.BuildCodexModelsManifestForGroup(
				context.Background(),
				&Group{ID: groupID, Platform: PlatformOpenAI},
				"",
				[]string{"gpt-5.6-sol"},
			)
			require.NoError(t, err)
			models := decodeCodexManifestModels(t, body)
			require.Len(t, models, 1)
			require.Equal(t, tt.modalities, models[0]["input_modalities"])
		})
	}
}

func TestBuildCodexModelsManifestForGroupUsesExplicitCompositeResponsesRouteModel(t *testing.T) {
	t.Parallel()

	const groupID int64 = 730
	routeRepo := compositeRouteRepoStub{routes: []CompositeModelRoute{{
		ID:             1,
		GroupID:        groupID,
		PublicModel:    "vision-alias",
		MatchType:      CompositeRouteMatchExact,
		TargetPlatform: PlatformGrok,
		UpstreamModel:  "grok-4.5",
		Endpoint:       CompositeRouteEndpointResponses,
		Enabled:        true,
	}}}
	svc := &GatewayService{
		accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
			groupID: {{ID: 20, Platform: PlatformGrok, Type: AccountTypeOAuth}},
		}},
		compositeResolver: NewCompositeRouteResolver(routeRepo),
	}

	body, err := svc.BuildCodexModelsManifestForGroup(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformComposite},
		"",
		[]string{"vision-alias"},
	)
	require.NoError(t, err)

	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 1)
	require.Equal(t, []any{"text", "image"}, models[0]["input_modalities"])
}

func TestBuildCodexModelsManifestForGroupUsesAccountMappingOwnershipAndMappedModel(t *testing.T) {
	t.Parallel()

	const groupID int64 = 731
	svc := &GatewayService{accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
		groupID: {{
			ID:       21,
			Platform: PlatformGrok,
			Type:     AccountTypeOAuth,
			Credentials: map[string]any{
				"model_mapping": map[string]any{"vision-alias": "grok-4.5"},
			},
		}},
	}}}

	body, err := svc.BuildCodexModelsManifestForGroup(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformComposite},
		"",
		[]string{"vision-alias"},
	)
	require.NoError(t, err)

	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 1)
	require.Equal(t, []any{"text", "image"}, models[0]["input_modalities"])
}

// Scenario: a Composite exact alias inherits metadata from its unique mapped target model.
func TestBuildCodexModelsManifestForGroupUsesMappedTargetMetadataForCompositeAlias(t *testing.T) {
	t.Parallel()

	const groupID int64 = 733
	svc := &GatewayService{accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
		groupID: {{
			ID:       23,
			Platform: PlatformAnthropic,
			Type:     AccountTypeAPIKey,
			Credentials: map[string]any{
				"model_mapping": map[string]any{"reasoning-alias": "claude-opus-4-8"},
			},
		}},
	}}}

	body, err := svc.BuildCodexModelsManifestForGroup(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformComposite},
		"",
		[]string{"reasoning-alias"},
	)
	require.NoError(t, err)

	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 1)
	require.Equal(t, "reasoning-alias", models[0]["slug"])
	require.Equal(t, "reasoning-alias", models[0]["display_name"])
	require.Equal(t, "Custom model routed through Sub2API.", models[0]["description"])
	require.Equal(t, []string{"low", "medium", "high", "xhigh", "max"}, effortsFromManifestModel(t, models[0]))
}

// Scenario: conflicting targets on the same platform keep the public alias but do not guess capabilities.
func TestBuildCodexModelsManifestForGroupUsesSafeFallbackForConflictingAliasTargets(t *testing.T) {
	t.Parallel()

	const groupID int64 = 734
	svc := &GatewayService{accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
		groupID: {
			{
				ID:       24,
				Platform: PlatformAnthropic,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"model_mapping": map[string]any{"shared-alias": "claude-opus-4-8"},
				},
			},
			{
				ID:       25,
				Platform: PlatformAnthropic,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"model_mapping": map[string]any{"shared-alias": "claude-haiku-4-5-20251001"},
				},
			},
		},
	}}}

	body, err := svc.BuildCodexModelsManifestForGroup(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformComposite},
		"",
		[]string{"shared-alias"},
	)
	require.NoError(t, err)

	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 1)
	require.Equal(t, "shared-alias", models[0]["slug"])
	require.Equal(t, "shared-alias", models[0]["display_name"])
	require.Equal(t, "Custom model routed through Sub2API.", models[0]["description"])
	require.Empty(t, effortsFromManifestModel(t, models[0]))
}

// Scenario: a media-only target remains hidden even when exposed through an ordinary alias.
func TestBuildCodexModelsManifestForGroupOmitsDedicatedMediaTargetAlias(t *testing.T) {
	t.Parallel()

	const groupID int64 = 735
	svc := &GatewayService{accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
		groupID: {{
			ID:       26,
			Platform: PlatformOpenAI,
			Type:     AccountTypeAPIKey,
			Credentials: map[string]any{
				"model_mapping": map[string]any{"creative-alias": "gpt-image-2"},
			},
		}},
	}}}

	body, err := svc.BuildCodexModelsManifestForGroup(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformComposite},
		"",
		[]string{"creative-alias"},
	)
	require.NoError(t, err)
	require.Empty(t, decodeCodexManifestModels(t, body))
}

func TestBuildCodexModelsManifestForGroupLoadsAccountsOnce(t *testing.T) {
	t.Parallel()

	const groupID int64 = 732
	repo := &countingCodexModelsAccountRepo{accounts: []Account{{
		ID:       22,
		Platform: PlatformGrok,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"vision-alias-a": "grok-4.5",
				"vision-alias-b": "grok-4.6",
			},
		},
	}}}
	svc := &GatewayService{accountRepo: repo}
	_, err := svc.BuildCodexModelsManifestForGroup(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformComposite},
		"",
		[]string{"vision-alias-a", "vision-alias-b", "deepseek-v4-pro"},
	)
	require.NoError(t, err)
	require.Equal(t, int32(1), repo.calls.Load())
	require.NotNil(t, repo.groupID)
	require.Equal(t, groupID, *repo.groupID)
	require.False(t, repo.includeGrouped)
	require.Contains(t, repo.platforms, PlatformOpenAI)
	require.Contains(t, repo.platforms, PlatformGrok)
	require.Contains(t, repo.platforms, PlatformDeepseek)
	require.NotContains(t, repo.platforms, PlatformComposite)
}

func TestBuildCodexModelsManifestForGroupUsesFallbackWhenTextOnlyPlatformHasNoSnapshot(t *testing.T) {
	t.Parallel()

	repo := &countingCodexModelsAccountRepo{}
	svc := &GatewayService{accountRepo: repo}
	body, err := svc.BuildCodexModelsManifestForGroup(
		context.Background(),
		&Group{ID: 733, Platform: PlatformDeepseek},
		"",
		[]string{"deepseek-v4-pro"},
	)
	require.NoError(t, err)
	require.Equal(t, int32(1), repo.calls.Load())

	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 1)
	require.Equal(t, []any{"text"}, models[0]["input_modalities"])
}

func TestBuildCodexModelsManifestForGroupFallsBackWhenCapabilityLookupFails(t *testing.T) {
	t.Parallel()

	repo := &countingCodexModelsAccountRepo{err: errors.New("account repository unavailable")}
	svc := &GatewayService{accountRepo: repo}
	body, err := svc.BuildCodexModelsManifestForGroup(
		context.Background(),
		&Group{ID: 734, Platform: PlatformComposite},
		"",
		[]string{"gpt-5.6-sol", "grok-4.5"},
	)
	require.NoError(t, err)
	require.Equal(t, int32(1), repo.calls.Load())

	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 2)
	require.Equal(t, []any{"text"}, models[0]["input_modalities"])
	require.Equal(t, []any{"text"}, models[1]["input_modalities"])
}

func TestMergeGroupConfiguredCodexModelsInjectsCurrentGroupAliases(t *testing.T) {
	t.Parallel()

	const groupID int64 = 71
	svc := &OpenAIGatewayService{accountRepo: codexModelsVisibilityAccountRepo{
		byGroup: map[int64][]Account{
			groupID: {
				{
					Platform: PlatformOpenAI,
					Credentials: map[string]any{
						"model_mapping": map[string]any{
							"deepseek-4-pro": "deepseek-v4-pro",
						},
					},
				},
			},
			72: {
				{
					Platform: PlatformOpenAI,
					Credentials: map[string]any{
						"model_mapping": map[string]any{"other-group-model": "upstream-model"},
					},
				},
			},
		},
	}}
	manifest := &CodexModelsManifest{
		Body: []byte(`{"models":[{"slug":"gpt-5.6","display_name":"GPT-5.6","unknown":{"kept":true}}],"metadata":{"version":1}}`),
	}

	err := svc.MergeGroupConfiguredCodexModels(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformOpenAI},
		manifest,
		"",
	)
	require.NoError(t, err)
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 2)
	require.Equal(t, "gpt-5.6", models[0]["slug"])
	require.Equal(t, map[string]any{"kept": true}, models[0]["unknown"])
	requireCompleteConfiguredCodexModel(t, models[1], "deepseek-4-pro")
	require.EqualValues(t, 1_000_000, models[1]["context_window"])
	require.EqualValues(t, 1_000_000, models[1]["max_context_window"])
	require.Equal(t, "high", models[1]["default_reasoning_level"])
	require.Len(t, models[1]["supported_reasoning_levels"], 3)
	require.NotContains(t, string(manifest.Body), "other-group-model")
	require.Equal(t, codexModelsManifestBodyETag(manifest.Body), manifest.ETag)
}

// Scenario: OpenAI 分组存在账号模型配置时直接生成本地 Codex 清单。
func TestBuildGroupConfiguredCodexModelsManifestUsesAdministratorConfiguration(t *testing.T) {
	t.Parallel()

	const groupID int64 = 77
	reasoning := true
	arkAccount := Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Credentials: map[string]any{
			"model_mapping": map[string]any{
				"glm-5.3":     "glm-5.3",
				"gpt-image-2": "gpt-image-2",
			},
		},
	}
	arkAccount.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
		"glm-5.3": {
			ID:                       "glm-5.3",
			DisplayName:              "GLM 5.3",
			Description:              "Ark coding model",
			Reasoning:                &reasoning,
			DefaultReasoningLevel:    "medium",
			SupportedReasoningLevels: []string{"low", "medium", "high"},
			InputModalities:          []string{"text"},
			ContextWindow:            1_000_000,
		},
	}})
	svc := &OpenAIGatewayService{accountRepo: codexModelsVisibilityAccountRepo{
		byGroup: map[int64][]Account{
			groupID: {
				{
					Platform: PlatformOpenAI,
					Type:     AccountTypeOAuth,
				},
				arkAccount,
			},
		},
	}}
	group := &Group{ID: groupID, Platform: PlatformOpenAI}

	manifest, configured, err := svc.BuildGroupConfiguredCodexModelsManifest(context.Background(), group, "")
	require.NoError(t, err)
	require.True(t, configured)
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 1)
	require.Equal(t, "glm-5.3", models[0]["slug"])
	require.Equal(t, "GLM 5.3", models[0]["display_name"])
	require.Equal(t, []string{"low", "medium", "high"}, effortsFromManifestModel(t, models[0]))
	require.Equal(t, "medium", models[0]["default_reasoning_level"])
	require.Equal(t, []any{"text"}, models[0]["input_modalities"])
	require.EqualValues(t, 1_000_000, models[0]["context_window"])
	require.Equal(t, codexModelsManifestBodyETag(manifest.Body), manifest.ETag)

	notModified, configured, err := svc.BuildGroupConfiguredCodexModelsManifest(
		context.Background(),
		group,
		"W/"+manifest.ETag,
	)
	require.NoError(t, err)
	require.True(t, configured)
	require.True(t, notModified.NotModified)
	require.Empty(t, notModified.Body)
	require.Equal(t, manifest.ETag, notModified.ETag)
}

// Scenario: OpenAI 通配映射展开组内精确选择，但不发布通配符 slug。
func TestBuildGroupConfiguredCodexModelsManifestExpandsSelectedModelCoveredByWildcardMapping(t *testing.T) {
	t.Parallel()

	const groupID int64 = 80
	svc := &OpenAIGatewayService{accountRepo: codexModelsVisibilityAccountRepo{
		byGroup: map[int64][]Account{
			groupID: {{
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Credentials: map[string]any{
					"model_mapping": map[string]any{"gpt-*": "gpt-5.6-sol"},
				},
			}},
		},
	}}
	group := &Group{
		ID:       groupID,
		Platform: PlatformOpenAI,
		ModelsListConfig: GroupModelsListConfig{
			Enabled: true,
			Models:  []string{"gpt-5.6"},
		},
	}

	manifest, configured, err := svc.BuildGroupConfiguredCodexModelsManifest(context.Background(), group, "")
	require.NoError(t, err)
	require.True(t, configured)
	require.Equal(t, []string{"gpt-5.6"}, codexManifestModelSlugs(t, manifest.Body))
	require.NotContains(t, string(manifest.Body), "gpt-*")
}

// Scenario: OpenAI 配置目录对仅因瞬态状态退出当前调度池的账号取能力交集，
// 且不发布其独有模型。持久 schedulable 仍为 true。
func TestBuildGroupConfiguredCodexModelsManifestIntersectsTransientlyUnschedulableMappedAccounts(t *testing.T) {
	t.Parallel()

	const groupID int64 = 79
	schedulable := newCodexCatalogMappedAccount(
		41,
		"gpt-5.6-sol",
		"GPT-5.6 Sol",
		[]string{"low", "medium", "high", "xhigh"},
		[]string{"text", "image"},
		1_000_000,
		true,
		nil,
	)
	transientlyUnschedulable := newCodexCatalogMappedAccount(
		42,
		"glm-5.3",
		"GLM 5.3",
		[]string{"low", "medium", "high"},
		[]string{"text"},
		272_000,
		true,
		map[string]any{"exclusive-model": "exclusive-upstream"},
	)
	svc := &OpenAIGatewayService{accountRepo: splitCodexModelsAccountRepo{
		schedulable: map[int64][]Account{groupID: {schedulable}},
		catalog:     map[int64][]Account{groupID: {schedulable, transientlyUnschedulable}},
	}}

	manifest, configured, err := svc.BuildGroupConfiguredCodexModelsManifest(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformOpenAI},
		"",
	)
	require.NoError(t, err)
	require.True(t, configured)
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 1)
	require.Equal(t, "my-coder", models[0]["slug"])
	require.Equal(t, "my-coder", models[0]["display_name"])
	require.Equal(t, []string{"low", "medium", "high"}, effortsFromManifestModel(t, models[0]))
	require.Equal(t, []any{"text"}, models[0]["input_modalities"])
	require.EqualValues(t, 272_000, models[0]["context_window"])
}

// Scenario: 管理员持久禁用的账号不能继续收窄 Codex 能力目录。
func TestBuildGroupConfiguredCodexModelsManifestIgnoresPersistentlyDisabledMappedAccounts(t *testing.T) {
	t.Parallel()

	const groupID int64 = 80
	enabled := newCodexCatalogMappedAccount(
		51,
		"gpt-5.6-sol",
		"GPT-5.6 Sol",
		[]string{"low", "medium", "high", "xhigh"},
		[]string{"text", "image"},
		1_000_000,
		true,
		nil,
	)
	disabled := newCodexCatalogMappedAccount(
		52,
		"glm-5.3",
		"GLM 5.3",
		[]string{"low", "medium", "high"},
		[]string{"text"},
		272_000,
		false,
		nil,
	)
	svc := &OpenAIGatewayService{accountRepo: splitCodexModelsAccountRepo{
		schedulable: map[int64][]Account{groupID: {enabled}},
		catalog:     map[int64][]Account{groupID: {enabled}},
		all:         map[int64][]Account{groupID: {enabled, disabled}},
	}}

	manifest, configured, err := svc.BuildGroupConfiguredCodexModelsManifest(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformOpenAI},
		"",
	)
	require.NoError(t, err)
	require.True(t, configured)
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 1)
	require.Equal(t, "my-coder", models[0]["slug"])
	require.Equal(t, []string{"low", "medium", "high", "xhigh"}, effortsFromManifestModel(t, models[0]))
	require.Equal(t, []any{"text", "image"}, models[0]["input_modalities"])
	require.EqualValues(t, 1_000_000, models[0]["context_window"])
}

// Scenario: 没有管理员模型配置时保留现有上游发现路径。
func TestBuildGroupConfiguredCodexModelsManifestFallsThroughWithoutConfiguration(t *testing.T) {
	t.Parallel()

	const groupID int64 = 78
	svc := &OpenAIGatewayService{accountRepo: codexModelsVisibilityAccountRepo{
		byGroup: map[int64][]Account{
			groupID: {{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}},
		},
	}}

	manifest, configured, err := svc.BuildGroupConfiguredCodexModelsManifest(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformOpenAI},
		"",
	)
	require.NoError(t, err)
	require.False(t, configured)
	require.Nil(t, manifest)
}

func TestMergeGroupConfiguredCodexModelsFiltersAutoReviewByDefault(t *testing.T) {
	t.Parallel()

	const groupID int64 = 74
	svc := &OpenAIGatewayService{accountRepo: codexModelsVisibilityAccountRepo{}}
	manifest := &CodexModelsManifest{
		Body: []byte(`{"models":[{"slug":"codex-auto-review","visibility":"list"},{"slug":"codex-auto-future","visibility":"list"},{"slug":"gpt-image-2","visibility":"list"},{"slug":"gpt-5.6","visibility":"list"}]}`),
	}

	require.NoError(t, svc.MergeGroupConfiguredCodexModels(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformOpenAI},
		manifest,
		"",
	))
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 1)
	require.Equal(t, "gpt-5.6", models[0]["slug"])
	require.Equal(t, codexModelsManifestBodyETag(manifest.Body), manifest.ETag)
}

// Scenario: OpenAI 账号映射不启用 Auto Review。
func TestMergeGroupConfiguredCodexModelsFiltersAccountMappedAutoReviewByDefault(t *testing.T) {
	t.Parallel()

	const groupID int64 = 75
	svc := &OpenAIGatewayService{accountRepo: codexModelsVisibilityAccountRepo{
		byGroup: map[int64][]Account{
			groupID: {
				{
					Platform: PlatformOpenAI,
					Credentials: map[string]any{
						"model_mapping": map[string]any{
							openai.CodexUsageProbeModel: openai.CodexUsageProbeModel,
						},
					},
				},
			},
		},
	}}
	manifest := &CodexModelsManifest{
		Body: []byte(`{"models":[{"slug":"codex-auto-review","visibility":"hide","model_messages":{"auto_review":{"enabled":true}}},{"slug":"gpt-5.6","visibility":"list"}]}`),
	}

	require.NoError(t, svc.MergeGroupConfiguredCodexModels(
		context.Background(),
		&Group{ID: groupID, Platform: PlatformOpenAI},
		manifest,
		"",
	))
	require.Equal(t, []string{"gpt-5.6"}, codexManifestModelSlugs(t, manifest.Body))
}

// Scenario: 启用的分组自定义列表允许 Auto Review。
func TestMergeGroupConfiguredCodexModelsKeepsExplicitAutoReviewSelection(t *testing.T) {
	t.Parallel()

	const groupID int64 = 76
	svc := &OpenAIGatewayService{accountRepo: codexModelsVisibilityAccountRepo{}}
	manifest := &CodexModelsManifest{
		Body: []byte(`{"models":[{"slug":"codex-auto-review","visibility":"list"},{"slug":"gpt-5.6","visibility":"list"}]}`),
	}
	group := &Group{
		ID:       groupID,
		Platform: PlatformOpenAI,
		ModelsListConfig: GroupModelsListConfig{
			Enabled: true,
			Models:  []string{openai.CodexUsageProbeModel},
		},
	}

	require.NoError(t, svc.MergeGroupConfiguredCodexModels(context.Background(), group, manifest, ""))
	require.Equal(t, []string{"codex-auto-review"}, codexManifestModelSlugs(t, manifest.Body))
}

func TestMergeGroupConfiguredCodexModelsHonorsCustomListAndFinalETag(t *testing.T) {
	t.Parallel()

	const groupID int64 = 73
	svc := &OpenAIGatewayService{accountRepo: codexModelsVisibilityAccountRepo{
		byGroup: map[int64][]Account{
			groupID: {
				{
					Platform: PlatformOpenAI,
					Credentials: map[string]any{
						"model_mapping": map[string]any{
							"deepseek-4-pro": "deepseek-v4-pro",
							"hidden-alias":   "hidden-upstream",
						},
					},
				},
			},
		},
	}}
	group := &Group{
		ID:       groupID,
		Platform: PlatformOpenAI,
		ModelsListConfig: GroupModelsListConfig{
			Enabled: true,
			Models:  []string{"deepseek-4-pro"},
		},
	}
	upstreamBody := []byte(`{"models":[{"slug":"gpt-5.6","display_name":"GPT-5.6"}]}`)
	manifest := &CodexModelsManifest{Body: upstreamBody}

	require.NoError(t, svc.MergeGroupConfiguredCodexModels(context.Background(), group, manifest, ""))
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 1)
	requireCompleteConfiguredCodexModel(t, models[0], "deepseek-4-pro")

	finalETag := manifest.ETag
	second := &CodexModelsManifest{Body: upstreamBody}
	require.NoError(t, svc.MergeGroupConfiguredCodexModels(context.Background(), group, second, finalETag))
	require.True(t, second.NotModified)
	require.Empty(t, second.Body)
	require.Equal(t, finalETag, second.ETag)
}

type codexModelsBlockingBody struct {
	ctx         context.Context
	readStarted chan struct{}
	startedOnce *sync.Once
	release     <-chan struct{}
	body        *strings.Reader
}

func (b *codexModelsBlockingBody) Read(p []byte) (int, error) {
	b.startedOnce.Do(func() { close(b.readStarted) })
	select {
	case <-b.release:
		return b.body.Read(p)
	case <-b.ctx.Done():
		return 0, b.ctx.Err()
	}
}

func (b *codexModelsBlockingBody) Close() error { return nil }

func (s *codexModelsHTTPUpstreamStub) Do(req *http.Request, proxyURL string, accountID int64, accountConcurrency int) (*http.Response, error) {
	return s.do(req, proxyURL, accountID, accountConcurrency)
}

func (s *codexModelsHTTPUpstreamStub) DoWithTLS(req *http.Request, proxyURL string, accountID int64, accountConcurrency int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return s.Do(req, proxyURL, accountID, accountConcurrency)
}

func TestFetchCodexModelsManifestOAuthPassesThroughVerbatim(t *testing.T) {
	manifestBody := `{"models":[{"slug":"gpt-5.6-sol","use_responses_lite":true}]}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccountID = r.Header.Get("chatgpt-account-id")
		gotOriginator = r.Header.Get("Originator")
		gotClientVersion = r.URL.Query().Get("client_version")
		w.Header().Set("ETag", `W/"abc123"`)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(manifestBody))
	}))
	defer server.Close()

	original := chatgptCodexModelsURL
	chatgptCodexModelsURL = server.URL
	defer func() { chatgptCodexModelsURL = original }()

	s := &OpenAIGatewayService{}
	manifest, err := s.FetchCodexModelsManifest(context.Background(), newCodexModelsTestAccount(), "0.137.0", "")
	if err != nil {
		t.Fatalf("FetchCodexModelsManifest returned error: %v", err)
	}

	if string(manifest.Body) != manifestBody {
		t.Errorf("body not passed through verbatim: got %q", manifest.Body)
	}
	if manifest.ETag != `W/"abc123"` {
		t.Errorf("etag not passed through: got %q", manifest.ETag)
	}
	if gotAuth != "Bearer test-access-token" {
		t.Errorf("authorization header: got %q", gotAuth)
	}
	if gotAccountID != "acc-123" {
		t.Errorf("chatgpt-account-id header: got %q", gotAccountID)
	}
	if gotOriginator != openai.CodexDefaultOriginator {
		t.Errorf("originator header: got %q", gotOriginator)
	}
	if gotClientVersion != "0.137.0" {
		t.Errorf("client_version query: got %q", gotClientVersion)
	}
}

func TestFetchCodexModelsManifestAgentIdentityUsesAssertionWithoutOAuthToken(t *testing.T) {
	key, privateKey := newTestAgentIdentityKey(t)
	account := &Account{
		ID:       3,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"auth_mode":          OpenAIAuthModeAgentIdentity,
			"agent_runtime_id":   key.runtimeID,
			"agent_private_key":  privateKey,
			"task_id":            key.taskID,
			"chatgpt_account_id": "acc-agent",
		},
	}

	var gotAuth, gotAccountID string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccountID = r.Header.Get("chatgpt-account-id")
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	defer server.Close()

	original := chatgptCodexModelsURL
	chatgptCodexModelsURL = server.URL
	defer func() { chatgptCodexModelsURL = original }()

	s := &OpenAIGatewayService{}
	manifest, err := s.FetchCodexModelsManifest(context.Background(), account, "0.137.0", "")
	if err != nil {
		t.Fatalf("FetchCodexModelsManifest returned error: %v", err)
	}
	if string(manifest.Body) != `{"models":[]}` {
		t.Fatalf("unexpected manifest body: %q", manifest.Body)
	}
	if !strings.HasPrefix(gotAuth, "AgentAssertion ") {
		t.Fatalf("authorization scheme: got %q", strings.SplitN(gotAuth, " ", 2)[0])
	}
	if gotAccountID != "acc-agent" {
		t.Fatalf("chatgpt-account-id header: got %q", gotAccountID)
	}
}

func TestFetchCodexModelsManifestAgentIdentityRecoversInvalidTaskOnce(t *testing.T) {
	key, privateKey := newTestAgentIdentityKey(t)
	account := &Account{
		ID:       4,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"auth_mode":          OpenAIAuthModeAgentIdentity,
			"agent_runtime_id":   key.runtimeID,
			"agent_private_key":  privateKey,
			"task_id":            "task-models-old",
			"chatgpt_account_id": "acc-agent-recovery",
		},
	}
	repo := &stubQuotaAccountRepo{accounts: map[int64]*Account{account.ID: account}}
	modelsCalls := 0
	registerCalls := 0
	var assertions []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		if strings.Contains(r.URL.Path, "/task/register") {
			registerCalls++
			_, _ = w.Write([]byte(`{"task_id":"task-models-new"}`))
			return
		}
		modelsCalls++
		assertions = append(assertions, r.Header.Get("Authorization"))
		if modelsCalls == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"invalid_task_id"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	defer server.Close()

	originalModelsURL := chatgptCodexModelsURL
	chatgptCodexModelsURL = server.URL
	t.Cleanup(func() { chatgptCodexModelsURL = originalModelsURL })
	originalAuthBase := openAIAgentIdentityAuthAPIBaseURL
	openAIAgentIdentityAuthAPIBaseURL = server.URL
	t.Cleanup(func() { openAIAgentIdentityAuthAPIBaseURL = originalAuthBase })

	s := &OpenAIGatewayService{accountRepo: repo}
	manifest, err := s.FetchCodexModelsManifest(context.Background(), account, "0.137.0", "")
	require.NoError(t, err)
	require.Equal(t, `{"models":[]}`, string(manifest.Body))
	require.Equal(t, 2, modelsCalls)
	require.Equal(t, 1, registerCalls)
	require.Len(t, assertions, 2)
	require.Equal(t, "task-models-old", decodeAgentAssertionTask(t, assertions[0]))
	require.Equal(t, "task-models-new", decodeAgentAssertionTask(t, assertions[1]))
}

func TestFetchCodexModelsManifestAgentIdentityRedactsUpstreamErrors(t *testing.T) {
	key, privateKey := newTestAgentIdentityKey(t)
	account := &Account{
		ID:       5,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"auth_mode":          OpenAIAuthModeAgentIdentity,
			"agent_runtime_id":   key.runtimeID,
			"agent_private_key":  privateKey,
			"task_id":            key.taskID,
			"chatgpt_account_id": "acc-agent-redaction",
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, `{"error":"%s %s %s AgentAssertion leaked"}`, key.runtimeID, key.taskID, privateKey)
	}))
	defer server.Close()
	original := chatgptCodexModelsURL
	chatgptCodexModelsURL = server.URL
	t.Cleanup(func() { chatgptCodexModelsURL = original })

	s := &OpenAIGatewayService{}
	_, err := s.FetchCodexModelsManifest(context.Background(), account, "0.137.0", "")
	require.Error(t, err)
	require.NotContains(t, err.Error(), key.runtimeID)
	require.NotContains(t, err.Error(), key.taskID)
	require.NotContains(t, err.Error(), privateKey)
	require.NotContains(t, err.Error(), "AgentAssertion leaked")
	require.Contains(t, err.Error(), "[redacted]")
}

func TestFetchCodexModelsManifestDefaultClientVersion(t *testing.T) {
	var gotClientVersion string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClientVersion = r.URL.Query().Get("client_version")
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	defer server.Close()

	original := chatgptCodexModelsURL
	chatgptCodexModelsURL = server.URL
	defer func() { chatgptCodexModelsURL = original }()

	s := &OpenAIGatewayService{}
	if _, err := s.FetchCodexModelsManifest(context.Background(), newCodexModelsTestAccount(), "", ""); err != nil {
		t.Fatalf("FetchCodexModelsManifest returned error: %v", err)
	}
	if gotClientVersion != CodexCanonicalClientVersion() {
		t.Errorf("default client_version: got %q, want %q", gotClientVersion, CodexCanonicalClientVersion())
	}
}

func TestFetchCodexModelsManifestNotModified(t *testing.T) {
	var gotIfNoneMatch string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotIfNoneMatch = r.Header.Get("If-None-Match")
		w.Header().Set("ETag", `W/"abc123"`)
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	original := chatgptCodexModelsURL
	chatgptCodexModelsURL = server.URL
	defer func() { chatgptCodexModelsURL = original }()

	s := &OpenAIGatewayService{}
	manifest, err := s.FetchCodexModelsManifest(context.Background(), newCodexModelsTestAccount(), "0.137.0", `W/"abc123"`)
	if err != nil {
		t.Fatalf("FetchCodexModelsManifest returned error: %v", err)
	}
	if !manifest.NotModified {
		t.Error("expected NotModified to be true")
	}
	if gotIfNoneMatch != `W/"abc123"` {
		t.Errorf("if-none-match header: got %q", gotIfNoneMatch)
	}
}

func TestFetchCodexModelsManifestUpstreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"detail":"boom"}`, http.StatusInternalServerError)
	}))
	defer server.Close()

	original := chatgptCodexModelsURL
	chatgptCodexModelsURL = server.URL
	defer func() { chatgptCodexModelsURL = original }()

	s := &OpenAIGatewayService{}
	if _, err := s.FetchCodexModelsManifest(context.Background(), newCodexModelsTestAccount(), "0.137.0", ""); err == nil {
		t.Fatal("expected error for upstream 500, got nil")
	}
}

func TestFetchCodexModelsManifestMissingToken(t *testing.T) {
	account := newCodexModelsTestAccount()
	delete(account.Credentials, "access_token")

	s := &OpenAIGatewayService{}
	if _, err := s.FetchCodexModelsManifest(context.Background(), account, "0.137.0", ""); err == nil {
		t.Fatal("expected error for missing access token, got nil")
	}
}

func TestFetchCodexModelsManifestAPIKeyCustomUpstream(t *testing.T) {
	manifestBody := `{"models":[{"slug":"deepseek-v4-pro"}]}`
	var gotRequest *http.Request
	var gotProxyURL string
	var gotAccountID int64
	var gotConcurrency int
	upstream := &codexModelsHTTPUpstreamStub{do: func(req *http.Request, proxyURL string, accountID int64, accountConcurrency int) (*http.Response, error) {
		gotRequest = req
		gotProxyURL = proxyURL
		gotAccountID = accountID
		gotConcurrency = accountConcurrency
		header := make(http.Header)
		header.Set("ETag", `W/"api-key-manifest"`)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     header,
			Body:       io.NopCloser(strings.NewReader(manifestBody)),
		}, nil
	}}

	s := newCodexModelsAPIKeyTestService(upstream)
	manifest, err := s.FetchCodexModelsManifest(
		context.Background(),
		newCodexModelsAPIKeyTestAccount("https://upstream.example/v1"),
		"0.144.0",
		"",
	)
	if err != nil {
		t.Fatalf("FetchCodexModelsManifest returned error: %v", err)
	}

	if gotRequest == nil {
		t.Fatal("expected request to custom API key upstream")
	}
	if gotRequest.Method != http.MethodGet {
		t.Errorf("method: got %q", gotRequest.Method)
	}
	if gotRequest.URL.String() != "https://upstream.example/v1/models?client_version=0.144.0" {
		t.Errorf("request URL: got %q", gotRequest.URL.String())
	}
	if gotRequest.Header.Get("Authorization") != "Bearer sk-upstream" {
		t.Errorf("authorization header: got %q", gotRequest.Header.Get("Authorization"))
	}
	if gotRequest.Header.Get("Originator") != openai.CodexDefaultOriginator {
		t.Errorf("originator header: got %q", gotRequest.Header.Get("Originator"))
	}
	if gotRequest.Header.Get("Version") != "0.144.0" {
		t.Errorf("version header must match the client_version query param: got %q", gotRequest.Header.Get("Version"))
	}
	if gotRequest.Header.Get("User-Agent") != CodexCanonicalUserAgent() {
		t.Errorf("user-agent header: got %q", gotRequest.Header.Get("User-Agent"))
	}
	if gotRequest.Header.Get("chatgpt-account-id") != "" {
		t.Errorf("chatgpt-account-id must not be sent to API key upstream: got %q", gotRequest.Header.Get("chatgpt-account-id"))
	}
	if gotProxyURL != "" || gotAccountID != 2 || gotConcurrency != 3 {
		t.Errorf("upstream routing metadata: proxy=%q account_id=%d concurrency=%d", gotProxyURL, gotAccountID, gotConcurrency)
	}
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 1)
	requireCompleteConfiguredCodexModel(t, models[0], "deepseek-v4-pro")
	require.Equal(t, "DeepSeek V4 Pro", models[0]["display_name"])
	require.EqualValues(t, 1_000_000, models[0]["context_window"])
	require.Equal(t, codexModelsManifestBodyETag(manifest.Body), manifest.ETag)
	require.Equal(t, `W/"api-key-manifest"`, manifest.upstreamETag)
}

// Scenario: 完整上游清单没有 ETag 时，最终正文仍生成强 ETag 并支持 304。
func TestFetchCodexModelsManifestAPIKeyCompleteBodyWithoutUpstreamETagUsesFinalBodyETag(t *testing.T) {
	completeBody, err := BuildCodexModelsManifest([]string{"custom-complete-model"})
	require.NoError(t, err)

	var calls atomic.Int32
	upstream := &codexModelsHTTPUpstreamStub{do: func(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(completeBody)),
		}, nil
	}}
	svc := newCodexModelsAPIKeyTestService(upstream)
	svc.accountRepo = codexModelsVisibilityAccountRepo{}
	account := newCodexModelsAPIKeyTestAccount("https://upstream.example/v1")
	group := &Group{ID: 82, Platform: PlatformOpenAI}

	first, err := svc.FetchCodexModelsManifest(context.Background(), account, "0.150.0", "")
	require.NoError(t, err)
	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(first, account))
	require.NoError(t, svc.MergeGroupConfiguredCodexModels(context.Background(), group, first, ""))
	require.Equal(t, codexModelsManifestBodyETag(first.Body), first.ETag)
	require.NotEmpty(t, first.ETag)

	second, err := svc.FetchCodexModelsManifest(context.Background(), account, "0.150.0", "")
	require.NoError(t, err)
	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(second, account))
	require.NoError(t, svc.MergeGroupConfiguredCodexModels(context.Background(), group, second, first.ETag))
	require.True(t, second.NotModified)
	require.Empty(t, second.Body)
	require.Equal(t, int32(1), calls.Load())
}

func TestFetchCodexModelsManifestAPIKeyConvertsStandardOpenAIModelList(t *testing.T) {
	upstreamBody := `{"object":"list","data":[{"id":"gpt-5.6","object":"model"},{"id":"  ","object":"model"},{"id":"gpt-5.6-codex","object":"model"}]}`
	upstream := &codexModelsHTTPUpstreamStub{do: func(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
		header := make(http.Header)
		header.Set("ETag", `W/"openai-list"`)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     header,
			Body:       io.NopCloser(strings.NewReader(upstreamBody)),
		}, nil
	}}

	s := newCodexModelsAPIKeyTestService(upstream)
	manifest, err := s.FetchCodexModelsManifest(
		context.Background(),
		newCodexModelsAPIKeyTestAccount("https://upstream.example/v1"),
		"0.144.0",
		"",
	)
	if err != nil {
		t.Fatalf("FetchCodexModelsManifest returned error: %v", err)
	}
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 2)
	requireCompleteConfiguredCodexModel(t, models[0], "gpt-5.6")
	requireCompleteConfiguredCodexModel(t, models[1], "gpt-5.6-codex")
	require.Equal(t, []any{"text", "image"}, models[0]["input_modalities"])
	require.Equal(t, []any{"text", "image"}, models[1]["input_modalities"])
	require.Equal(t, codexModelsManifestBodyETag(manifest.Body), manifest.ETag)
	require.Equal(t, `W/"openai-list"`, manifest.upstreamETag)
}

func TestConvertOpenAIModelListToCodexManifestUsesCompleteDescriptors(t *testing.T) {
	upstreamBody := `{"object":"list","data":[{"id":"gpt-5.5","object":"model"}]}`

	converted := convertOpenAIModelListToCodexManifest([]byte(upstreamBody))
	models := decodeCodexManifestModels(t, converted)

	require.Len(t, models, 1)
	requireCompleteConfiguredCodexModel(t, models[0], "gpt-5.5")
	require.Equal(t, "GPT-5.5", models[0]["display_name"])
	require.Equal(t, "medium", models[0]["default_reasoning_level"])
	require.Len(t, models[0]["supported_reasoning_levels"], 4)
}

func TestCompleteAPIKeyCodexModelsManifestForClientPreservesProviderMetadata(t *testing.T) {
	t.Parallel()

	svc := &OpenAIGatewayService{}
	manifest := &CodexModelsManifest{
		Body: []byte(`{"models":[{"slug":"grok-4.6","description":"Provider supplied","service_tiers":[{"id":"provider-priority","name":"Provider Fast","description":"Provider supplied tier."}],"model_messages":{"auto_review":{"enabled":true}},"truncation_policy":{"mode":"tokens"},"unknown":{"kept":true}}],"metadata":{"source":"upstream"}}`),
	}
	account := newCodexModelsAPIKeyTestAccount("https://upstream.example/v1")

	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(manifest, account))
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 1)
	requireCompleteConfiguredCodexModel(t, models[0], "grok-4.6")
	require.Equal(t, "Provider supplied", models[0]["description"])
	require.Equal(t, map[string]any{"kept": true}, models[0]["unknown"])
	require.Equal(t, []any{
		map[string]any{
			"id":          "provider-priority",
			"name":        "Provider Fast",
			"description": "Provider supplied tier.",
		},
	}, models[0]["service_tiers"])
	require.Equal(t, []any{"text"}, models[0]["input_modalities"])
	require.Equal(t, []string{"low", "medium", "high", "xhigh"}, effortsFromManifestModel(t, models[0]))
	modelMessages, ok := models[0]["model_messages"].(map[string]any)
	require.True(t, ok)
	require.NotEmpty(t, modelMessages["instructions_template"])
	require.Equal(t, map[string]any{"enabled": true}, modelMessages["auto_review"])
	require.Equal(t, map[string]any{"mode": "tokens", "limit": float64(10_000)}, models[0]["truncation_policy"])
	require.Equal(t, codexModelsManifestBodyETag(manifest.Body), manifest.ETag)

	var envelope map[string]any
	require.NoError(t, json.Unmarshal(manifest.Body, &envelope))
	require.Equal(t, map[string]any{"source": "upstream"}, envelope["metadata"])
}

func TestCompleteAPIKeyCodexModelsManifestForClientUsesKnownGPTImageFallback(t *testing.T) {
	t.Parallel()

	svc := &OpenAIGatewayService{}
	manifest := &CodexModelsManifest{Body: []byte(`{"models":[
		{"slug":"gpt-5.6-sol"},
		{"slug":"company-coding-model"},
		{"slug":"gpt-4o","input_modalities":["text"]}
	]}`)}
	account := newCodexModelsAPIKeyTestAccount("https://openai-compatible.example.test/v1")

	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(manifest, account))
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 3)

	bySlug := make(map[string]map[string]any, len(models))
	for _, model := range models {
		slug, ok := model["slug"].(string)
		require.True(t, ok)
		bySlug[slug] = model
	}
	require.Equal(t, []any{"text", "image"}, bySlug["gpt-5.6-sol"]["input_modalities"])
	require.Equal(t, []any{"text"}, bySlug["company-coding-model"]["input_modalities"])
	require.Equal(t, []any{"text"}, bySlug["gpt-4o"]["input_modalities"])
}

// Scenario: 标准 /models 型号列表优先使用已同步账号能力，再使用本地 descriptor 兜底。
func TestCompleteAPIKeyCodexModelsManifestForClientUsesSyncedMetadataForConvertedModelList(t *testing.T) {
	t.Parallel()

	reasoning := true
	account := newCodexModelsAPIKeyTestAccount("https://upstream.example/v1")
	account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
		"future-reasoner": {
			ID: "future-reasoner", DisplayName: "Future Reasoner", Description: "Synced upstream capability",
			Reasoning: &reasoning, DefaultReasoningLevel: "ultra",
			SupportedReasoningLevels: []string{"low", "high", "ultra"},
			InputModalities:          []string{"text", "image"},
			ContextWindow:            999_000,
		},
	}})
	upstream := &codexModelsHTTPUpstreamStub{do: func(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"object":"list","data":[{"id":"future-reasoner","object":"model"}]}`)),
		}, nil
	}}
	svc := newCodexModelsAPIKeyTestService(upstream)

	manifest, err := svc.FetchCodexModelsManifest(context.Background(), account, "0.150.0", "")
	require.NoError(t, err)
	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(manifest, account))

	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 1)
	require.Equal(t, "Future Reasoner", models[0]["display_name"])
	require.Equal(t, "Synced upstream capability", models[0]["description"])
	require.Equal(t, "ultra", models[0]["default_reasoning_level"])
	require.Equal(t, []string{"low", "high", "ultra"}, effortsFromManifestModel(t, models[0]))
	require.Equal(t, []any{"text", "image"}, models[0]["input_modalities"])
	require.EqualValues(t, 999_000, models[0]["context_window"])
	require.EqualValues(t, 999_000, models[0]["max_context_window"])
}

func TestCompleteAPIKeyCodexModelsManifestForClientFillsMissingProviderFieldsWithoutOverwritingExplicitMetadata(t *testing.T) {
	t.Parallel()

	reasoning := true
	account := newCodexModelsAPIKeyTestAccount("https://upstream.example/v1")
	account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
		"provider-model": {
			ID: "provider-model", DisplayName: "Synced Display", Description: "Synced description",
			Reasoning: &reasoning, DefaultReasoningLevel: "ultra",
			SupportedReasoningLevels: []string{"high", "ultra"},
			InputModalities:          []string{"text", "image"},
			ContextWindow:            999_000,
		},
	}})
	manifest := &CodexModelsManifest{Body: []byte(`{"models":[{
		"slug":"provider-model",
		"description":"Provider supplied",
		"context_window":64000,
		"max_context_window":64000
	}]}`)}
	svc := &OpenAIGatewayService{}

	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(manifest, account))
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 1)
	require.Equal(t, "Synced Display", models[0]["display_name"])
	require.Equal(t, "Provider supplied", models[0]["description"])
	require.Equal(t, "ultra", models[0]["default_reasoning_level"])
	require.Equal(t, []string{"high", "ultra"}, effortsFromManifestModel(t, models[0]))
	require.Equal(t, []any{"text", "image"}, models[0]["input_modalities"])
	require.EqualValues(t, 64_000, models[0]["context_window"])
	require.EqualValues(t, 64_000, models[0]["max_context_window"])
}

// Scenario: 原生 manifest 的缺失字段在命中缓存后仍使用账号当前同步快照，而不是缓存中的本地默认值。
func TestCompleteAPIKeyCodexModelsManifestForClientUsesCurrentSnapshotForCachedNativeManifest(t *testing.T) {
	t.Parallel()

	reasoning := true
	account := newCodexModelsAPIKeyTestAccount("https://upstream.example/v1")
	setSnapshot := func(displayName string, contextWindow int64) {
		account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
			"deepseek-v4-pro": {
				ID: "deepseek-v4-pro", DisplayName: displayName,
				Reasoning: &reasoning, DefaultReasoningLevel: "ultra",
				SupportedReasoningLevels: []string{"high", "ultra"},
				InputModalities:          []string{"text", "image"},
				ContextWindow:            contextWindow,
			},
		}})
	}
	setSnapshot("Synced DeepSeek", 256_000)

	var calls atomic.Int32
	upstream := &codexModelsHTTPUpstreamStub{do: func(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{"models":[{
				"slug":"deepseek-v4-pro",
				"description":"Provider supplied"
			}]}`)),
		}, nil
	}}
	svc := newCodexModelsAPIKeyTestService(upstream)

	first, err := svc.FetchCodexModelsManifest(context.Background(), account, "0.150.0", "")
	require.NoError(t, err)
	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(first, account))
	firstModel := decodeCodexManifestModels(t, first.Body)[0]
	require.Equal(t, "Synced DeepSeek", firstModel["display_name"])
	require.Equal(t, "Provider supplied", firstModel["description"])
	require.Equal(t, "ultra", firstModel["default_reasoning_level"])
	require.Equal(t, []string{"high", "ultra"}, effortsFromManifestModel(t, firstModel))
	require.Equal(t, []any{"text", "image"}, firstModel["input_modalities"])
	require.EqualValues(t, 256_000, firstModel["context_window"])

	setSnapshot("Refreshed DeepSeek", 512_000)
	second, err := svc.FetchCodexModelsManifest(context.Background(), account, "0.150.0", "")
	require.NoError(t, err)
	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(second, account))
	secondModel := decodeCodexManifestModels(t, second.Body)[0]
	require.Equal(t, "Refreshed DeepSeek", secondModel["display_name"])
	require.Equal(t, "Provider supplied", secondModel["description"])
	require.EqualValues(t, 512_000, secondModel["context_window"])
	require.Equal(t, int32(1), calls.Load(), "second response should use the cached upstream source body")
}

func TestCompleteAPIKeyCodexModelsManifestForClientMarksOnlyOfficialVisionGPTImageInput(t *testing.T) {
	t.Parallel()

	svc := &OpenAIGatewayService{}
	manifest := &CodexModelsManifest{Body: []byte(`{"models":[{"slug":"gpt-6-astra"},{"slug":"gpt-5.6-sol"},{"slug":"gpt-4o"},{"slug":"gpt-3.5-turbo"},{"slug":"gpt-4"}]}`)}
	account := newCodexModelsAPIKeyTestAccount("")

	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(manifest, account))
	models := decodeCodexManifestModels(t, manifest.Body)
	require.Len(t, models, 5)

	bySlug := make(map[string]map[string]any, len(models))
	for _, model := range models {
		slug, ok := model["slug"].(string)
		require.True(t, ok)
		bySlug[slug] = model
	}
	for _, slug := range []string{"gpt-6-astra", "gpt-5.6-sol", "gpt-4o"} {
		require.Equal(t, []any{"text", "image"}, bySlug[slug]["input_modalities"])
		require.Equal(t, true, bySlug[slug]["supports_image_detail_original"])
	}
	for _, slug := range []string{"gpt-3.5-turbo", "gpt-4"} {
		require.Equal(t, []any{"text"}, bySlug[slug]["input_modalities"])
		require.Equal(t, false, bySlug[slug]["supports_image_detail_original"])
	}
	require.Equal(t, codexModelsManifestBodyETag(manifest.Body), manifest.ETag)
}

func TestCompleteAPIKeyCodexModelsManifestForClientFiltersOfficialNonAgentModels(t *testing.T) {
	t.Parallel()

	svc := &OpenAIGatewayService{}
	manifest := &CodexModelsManifest{Body: []byte(`{"models":[{"slug":"gpt-5.6-sol"},{"slug":"gpt-4o-realtime-preview"},{"slug":"gpt-4o-mini-tts"},{"slug":"text-embedding-3-large"},{"slug":"omni-moderation-latest"},{"slug":"o4-mini"},{"slug":"codex-mini-latest"}]}`)}
	account := newCodexModelsAPIKeyTestAccount("")

	require.NoError(t, svc.CompleteAPIKeyCodexModelsManifestForClient(manifest, account))
	require.Equal(t, []string{"gpt-5.6-sol", "o4-mini", "codex-mini-latest"}, codexManifestModelSlugs(t, manifest.Body))
	require.Equal(t, codexModelsManifestBodyETag(manifest.Body), manifest.ETag)
}

func TestAdjustAPIKeyCodexModelsManifest(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "affected models disable responses lite and preserve unknown fields",
			body: `{"models":[{"slug":"gpt-6-astra","use_responses_lite":true},{"slug":"gpt-5.6-sol","use_responses_lite":true,"unknown_model":{"enabled":true}},{"slug":"gpt-5.6-terra","use_responses_lite":true},{"slug":"gpt-5.6-luna","use_responses_lite":true}],"unknown_top":{"version":1}}`,
			want: `{"models":[{"slug":"gpt-6-astra","use_responses_lite":false},{"slug":"gpt-5.6-sol","unknown_model":{"enabled":true},"use_responses_lite":false},{"slug":"gpt-5.6-terra","use_responses_lite":false},{"slug":"gpt-5.6-luna","use_responses_lite":false}],"unknown_top":{"version":1}}`,
		},
		{
			name: "unaffected model unchanged",
			body: ` {"models":[{"slug":"gpt-5.6-codex","use_responses_lite":true}]} `,
			want: ` {"models":[{"slug":"gpt-5.6-codex","use_responses_lite":true}]} `,
		},
		{
			name: "false missing and alternate entries unchanged",
			body: `{"models":[{"slug":"gpt-5.6-sol","use_responses_lite":false},{"slug":"gpt-5.6-terra"},null,"gpt-5.6-luna",{"slug":17,"use_responses_lite":true}]}`,
			want: `{"models":[{"slug":"gpt-5.6-sol","use_responses_lite":false},{"slug":"gpt-5.6-terra"},null,"gpt-5.6-luna",{"slug":17,"use_responses_lite":true}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := adjustAPIKeyCodexModelsManifest([]byte(tt.body))
			require.NoError(t, err)
			require.Equal(t, tt.want, string(got))
		})
	}
}

func TestFetchCodexModelsManifestAPIKeyDisablesResponsesLiteForAffectedModels(t *testing.T) {
	const upstreamBody = `{"models":[{"slug":"gpt-5.6-sol","use_responses_lite":true},{"slug":"gpt-5.6-codex","use_responses_lite":true}],"metadata":{"version":1}}`
	upstream := &codexModelsHTTPUpstreamStub{do: func(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Etag": []string{`"upstream-strong"`}},
			Body:       io.NopCloser(strings.NewReader(upstreamBody)),
		}, nil
	}}

	s := newCodexModelsAPIKeyTestService(upstream)
	manifest, err := s.FetchCodexModelsManifest(context.Background(), newCodexModelsAPIKeyTestAccount("https://upstream.example"), "0.145.0", "")
	require.NoError(t, err)
	require.JSONEq(t, `{"models":[{"slug":"gpt-5.6-sol","use_responses_lite":false},{"slug":"gpt-5.6-codex","use_responses_lite":true}],"metadata":{"version":1}}`, string(manifest.Body))
	require.Equal(t, codexModelsManifestBodyETag(manifest.Body), manifest.ETag)
	require.Equal(t, `"upstream-strong"`, manifest.upstreamETag)

	notModified, err := s.FetchCodexModelsManifest(context.Background(), newCodexModelsAPIKeyTestAccount("https://upstream.example"), "0.145.0", manifest.ETag)
	require.NoError(t, err)
	require.True(t, notModified.NotModified)
	require.Equal(t, manifest.ETag, notModified.ETag)
}

func TestFetchCodexModelsManifestOAuthPreservesResponsesLite(t *testing.T) {
	const manifestBody = ` {"models":[{"slug":"gpt-5.6-sol","use_responses_lite":true}]} `
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(manifestBody))
	}))
	defer server.Close()
	original := chatgptCodexModelsURL
	chatgptCodexModelsURL = server.URL
	defer func() { chatgptCodexModelsURL = original }()

	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"access_token": "oauth-token"}}
	manifest, err := (&OpenAIGatewayService{}).FetchCodexModelsManifest(context.Background(), account, "0.144.1", "")
	require.NoError(t, err)
	require.Equal(t, manifestBody, string(manifest.Body))
	require.Equal(t, `W/"oauth"`, manifest.ETag)
}

func TestFetchCodexModelsManifestAPIKeyAdjustsOnlyTargetedModels(t *testing.T) {
	upstream := &codexManifestHTTPStub{body: `{"models":[{"slug":"gpt-5.6-sol","use_responses_lite":true},{"slug":"gpt-5.5","use_responses_lite":true}]}`, etag: `"upstream"`}
	service := &OpenAIGatewayService{
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
		httpUpstream: upstream,
	}
	account := &Account{ID: 2, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://provider.example/v1"}}
	manifest, err := service.FetchCodexModelsManifest(context.Background(), account, "0.144.1", "")
	require.NoError(t, err)
	require.Contains(t, string(manifest.Body), `"slug":"gpt-5.6-sol"`)
	require.Contains(t, string(manifest.Body), `"use_responses_lite":false`)
	require.Contains(t, string(manifest.Body), `"slug":"gpt-5.5","use_responses_lite":true`)
	require.NotEqual(t, `"upstream"`, manifest.ETag)
}

func TestFetchCodexModelsManifestAPIKeyConvertsOpenAIModelList(t *testing.T) {
	upstream := &codexManifestHTTPStub{body: `{"object":"list","data":[{"id":"gpt-5.6-sol"}]}`}
	service := &OpenAIGatewayService{
		cfg:          &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}},
		httpUpstream: upstream,
	}
	account := &Account{ID: 3, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://provider.example/v1"}}
	manifest, err := service.FetchCodexModelsManifest(context.Background(), account, "", "")
	require.NoError(t, err)
	require.JSONEq(t, `{"models":[{"slug":"gpt-5.6-sol"}]}`, string(manifest.Body))
}

func TestFetchCodexModelsManifestUsesConfiguredBodyLimit(t *testing.T) {
	upstream := &codexManifestHTTPStub{body: `{"models":[{"slug":"gpt-5.6"}]}`}
	service := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{ModelsListReadMaxBytes: 8}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}}, httpUpstream: upstream}
	account := &Account{ID: 5, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://provider.example/v1"}}
	_, err := service.FetchCodexModelsManifest(context.Background(), account, "0.144.0", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "response exceeds 8 bytes")
	require.True(t, IsRetryableCodexModelsManifestError(err))
}

func TestFetchCodexModelsManifestAcceptsConfiguredLimitAboveLegacyBoundary(t *testing.T) {
	body := `{"models":[{"slug":"gpt-5.6","display_name":"` + strings.Repeat("x", (8<<20)+1024) + `"}]}`
	upstream := &codexManifestHTTPStub{body: body}
	service := &OpenAIGatewayService{cfg: &config.Config{Gateway: config.GatewayConfig{ModelsListReadMaxBytes: 16 << 20}, Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{Enabled: false}}}, httpUpstream: upstream}
	account := &Account{ID: 6, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://provider.example/v1"}}
	manifest, err := service.FetchCodexModelsManifest(context.Background(), account, "0.144.0", "")
	require.NoError(t, err)
	require.Equal(t, body, string(manifest.Body))
}

func TestCodexModelsManifestETagMatchesWeakAndMultipleValues(t *testing.T) {
	require.True(t, codexModelsManifestETagMatches(`"other", W/"abc"`, `"abc"`))
	require.True(t, codexModelsManifestETagMatches("*", `"abc"`))
	require.False(t, codexModelsManifestETagMatches(`"other"`, `"abc"`))
}
