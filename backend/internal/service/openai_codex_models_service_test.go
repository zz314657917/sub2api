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
		require.Equal(t, "Bearer oauth-token", r.Header.Get("Authorization"))
		require.Equal(t, openai.CodexDefaultOriginator, r.Header.Get("Originator"))
		require.Equal(t, codexCLIUserAgent, r.Header.Get("User-Agent"))
		require.Equal(t, "0.144.1", r.URL.Query().Get("client_version"))
		w.Header().Set("ETag", `W/"oauth"`)
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
