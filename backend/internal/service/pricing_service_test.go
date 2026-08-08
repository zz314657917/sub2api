package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestPricingSchedulerBlankRemoteURLDoesNotStart(t *testing.T) {
	svc := NewPricingService(&config.Config{Pricing: config.PricingConfig{RemoteURL: "  \t  "}}, nil)
	defer svc.Stop()

	svc.startUpdateScheduler()
	done := make(chan struct{})
	go func() {
		svc.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("blank remote URL must not start scheduler")
	}
}

func TestPricingNonEmptyInvalidRemoteURLStillReturnsValidationError(t *testing.T) {
	svc := NewPricingService(&config.Config{Pricing: config.PricingConfig{
		RemoteURL: "://invalid",
		DataDir:   t.TempDir(),
	}}, nil)

	err := svc.ForceUpdate()

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid pricing url")
}

func TestParsePricingData_ParsesPriorityAndServiceTierFields(t *testing.T) {
	svc := &PricingService{}
	body := []byte(`{
		"gpt-5.4": {
			"input_cost_per_token": 0.0000025,
			"input_cost_per_token_priority": 0.000005,
			"output_cost_per_token": 0.000015,
			"output_cost_per_token_priority": 0.00003,
			"cache_creation_input_token_cost": 0.0000025,
			"cache_read_input_token_cost": 0.00000025,
			"cache_read_input_token_cost_priority": 0.0000005,
			"supports_service_tier": true,
			"supports_prompt_caching": true,
			"litellm_provider": "openai",
			"mode": "chat"
		}
	}`)

	data, err := svc.parsePricingData(body)
	require.NoError(t, err)
	pricing := data["gpt-5.4"]
	require.NotNil(t, pricing)
	require.InDelta(t, 5e-6, pricing.InputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 3e-5, pricing.OutputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 5e-7, pricing.CacheReadInputTokenCostPriority, 1e-12)
	require.True(t, pricing.SupportsServiceTier)
}

func TestParsePricingData_ParsesPriorityCacheCreationField(t *testing.T) {
	svc := &PricingService{}
	data, err := svc.parsePricingData([]byte(`{
		"gpt-5.6-sol": {
			"input_cost_per_token": 0.000005,
			"output_cost_per_token": 0.00003,
			"cache_creation_input_token_cost_priority": 0.0000125
		},
		"gpt-5.6-zero": {
			"input_cost_per_token": 0.000005,
			"output_cost_per_token": 0.00003,
			"cache_creation_input_token_cost_priority": 0
		}
	}`))
	require.NoError(t, err)
	require.InDelta(t, 12.5e-6, data["gpt-5.6-sol"].CacheCreationInputTokenCostPriority, 1e-12)
	require.Zero(t, data["gpt-5.6-zero"].CacheCreationInputTokenCostPriority)
}

func TestParsePricingData_ParsesImageInputTokenPrice(t *testing.T) {
	svc := &PricingService{}
	data, err := svc.parsePricingData([]byte(`{
		"gpt-image-2": {
			"input_cost_per_token": 0.000005,
			"output_cost_per_token": 0.00001,
			"input_cost_per_image_token": 0.000008,
			"output_cost_per_image_token": 0.00003
		},
		"image-only": {
			"input_cost_per_image_token": 0.000008
		}
	}`))
	require.NoError(t, err)
	require.InDelta(t, 8e-6, data["gpt-image-2"].InputCostPerImageToken, 1e-12)
	require.False(t, data["gpt-image-2"].TokenPricingAbsent)
	require.InDelta(t, 8e-6, data["image-only"].InputCostPerImageToken, 1e-12)
	require.True(t, data["image-only"].TokenPricingAbsent)
}

func TestDefaultBuild_AvailableChannelPricingPreservesImageInputPrice(t *testing.T) {
	imageInputPrice := 8e-6
	require.False(t, pricingNeedsFallback(&ChannelModelPricing{ImageInputPrice: &imageInputPrice}))

	for _, mode := range []string{"chat", "image_generation"} {
		t.Run(mode, func(t *testing.T) {
			got := synthesizePricingFromLiteLLM(&LiteLLMModelPricing{
				Mode:                        mode,
				InputCostPerToken:           3e-6,
				InputCostPerImageToken:      imageInputPrice,
				OutputCostPerImageToken:     4e-5,
				CacheReadInputTokenCost:     3e-7,
				CacheCreationInputTokenCost: 3.75e-6,
			}, nil)
			require.NotNil(t, got)
			require.NotNil(t, got.ImageInputPrice)
			require.InDelta(t, imageInputPrice, *got.ImageInputPrice, 1e-12)
		})
	}
}

func TestDefaultBuild_GLM52FallbackUsesOfficialGLM51Price(t *testing.T) {
	svc := NewBillingService(nil, nil)

	got, err := svc.GetModelPricing("glm-5.2")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.InDelta(t, 1.4e-6, got.InputPricePerToken, 1e-12)
	require.InDelta(t, 4.4e-6, got.OutputPricePerToken, 1e-12)
	require.InDelta(t, 0.26e-6, got.CacheReadPricePerToken, 1e-12)
}

func TestDefaultBuild_LongContextPreservesImageInputCost(t *testing.T) {
	svc := NewBillingService(&config.Config{}, nil)
	svc.fallbackPrices["glm-5.2"] = &ModelPricing{
		InputPricePerToken:      3e-6,
		ImageInputPricePerToken: 9e-6,
		OutputPricePerToken:     15e-6,
		CacheReadPricePerToken:  0.3e-6,
	}
	tokens := UsageTokens{
		InputTokens:      150000,
		ImageInputTokens: 60000,
		OutputTokens:     1000,
		CacheReadTokens:  100000,
	}

	cost, err := svc.CalculateCostWithLongContext("glm-5.2", tokens, 1.0, 200000, 2.0)
	require.NoError(t, err)

	expectedTextInput := float64(60000+30000) * 3e-6
	expectedImageInput := float64(40000+20000) * 9e-6
	expectedTotal := expectedTextInput + expectedImageInput + float64(tokens.OutputTokens)*15e-6 + float64(tokens.CacheReadTokens)*0.3e-6
	expectedActual := float64(60000)*3e-6 + float64(40000)*9e-6 + float64(tokens.OutputTokens)*15e-6 + float64(tokens.CacheReadTokens)*0.3e-6 +
		(float64(30000)*3e-6+float64(20000)*9e-6)*2
	require.InDelta(t, expectedTextInput, cost.InputCost, 1e-12)
	require.InDelta(t, expectedImageInput, cost.ImageInputCost, 1e-12)
	require.InDelta(t, expectedTotal, cost.TotalCost, 1e-12)
	require.InDelta(t, expectedActual, cost.ActualCost, 1e-12)
}

func TestDefaultBuild_AccountStatsPricingUsesImageInputPrice(t *testing.T) {
	inputPrice := 0.001
	imageInputPrice := 0.004
	outputPrice := 0.002
	pricing := &ChannelModelPricing{
		BillingMode:     BillingModeToken,
		InputPrice:      &inputPrice,
		ImageInputPrice: &imageInputPrice,
		OutputPrice:     &outputPrice,
	}
	tokens := UsageTokens{InputTokens: 100, ImageInputTokens: 40, OutputTokens: 50}

	customCost := calculateStatsCost(pricing, tokens, 1)
	require.NotNil(t, customCost)
	require.InDelta(t, 0.32, *customCost, 1e-12)

	svc := NewBillingService(&config.Config{}, nil)
	svc.fallbackPrices["glm-5.2"] = &ModelPricing{
		InputPricePerToken:      0.001,
		ImageInputPricePerToken: 0.004,
		OutputPricePerToken:     0.002,
	}
	modelCost := tryModelFilePricing(svc, "glm-5.2", tokens)
	require.NotNil(t, modelCost)
	require.InDelta(t, 0.32, *modelCost, 1e-12)
}

func TestDefaultBuild_AccountStatsImageInputFallbackAndBounds(t *testing.T) {
	require.InDelta(t, 0.1, calculateInputTokenStatsCost(100, 40, 0.001, 0), 1e-12)
	require.InDelta(t, 0.4, calculateInputTokenStatsCost(100, 400, 0.001, 0.004), 1e-12)
	require.Zero(t, calculateInputTokenStatsCost(100, -1, 0, 0.004))
}

func TestDefaultBuild_ProportionalTokenCountAvoidsIntegerOverflow(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	require.Equal(t, maxInt-3, proportionalTokenCount(maxInt-1, maxInt-2, maxInt))
	require.Equal(t, 1, proportionalTokenCount(3, 2, 4))
	require.Zero(t, proportionalTokenCount(3, 2, 0))
}

func TestDefaultBuild_LongContextTotalAvoidsIntegerOverflow(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	require.Equal(t, maxInt, saturatingAddNonNegativeInts(maxInt-5, 10))

	svc := NewBillingService(&config.Config{}, nil)
	svc.fallbackPrices["glm-5.2"] = &ModelPricing{
		InputPricePerToken:      1e-6,
		ImageInputPricePerToken: 2e-6,
		CacheReadPricePerToken:  1e-7,
	}
	tokens := UsageTokens{
		InputTokens:      maxInt/2 + 100,
		ImageInputTokens: maxInt/2 + 100,
		CacheReadTokens:  maxInt/2 + 100,
	}

	cost, err := svc.CalculateCostWithLongContext("glm-5.2", tokens, 1, maxInt/2, 2)
	require.NoError(t, err)
	require.NotNil(t, cost)
	require.GreaterOrEqual(t, cost.ImageInputCost, 0.0)
	require.Greater(t, cost.ActualCost, cost.TotalCost)
}

func TestGetModelPricing_Gpt53CodexSparkUsesGpt51CodexPricing(t *testing.T) {
	sparkPricing := &LiteLLMModelPricing{InputCostPerToken: 1}
	gpt53Pricing := &LiteLLMModelPricing{InputCostPerToken: 9}

	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": sparkPricing,
			"gpt-5.3":       gpt53Pricing,
		},
	}

	got := svc.GetModelPricing("gpt-5.3-codex-spark")
	require.Same(t, sparkPricing, got)
}

func TestGetModelPricing_Gpt53CodexFallbackStillUsesGpt52Codex(t *testing.T) {
	gpt52CodexPricing := &LiteLLMModelPricing{InputCostPerToken: 2}

	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.2-codex": gpt52CodexPricing,
		},
	}

	got := svc.GetModelPricing("gpt-5.3-codex")
	require.Same(t, gpt52CodexPricing, got)
}

func TestGetModelPricing_OpenAIFallbackMatchedLoggedAsInfo(t *testing.T) {
	logSink, restore := captureStructuredLog(t)
	defer restore()

	gpt52CodexPricing := &LiteLLMModelPricing{InputCostPerToken: 2}
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.2-codex": gpt52CodexPricing,
		},
	}

	got := svc.GetModelPricing("gpt-5.3-codex")
	require.Same(t, gpt52CodexPricing, got)

	require.True(t, logSink.ContainsMessageAtLevel("[Pricing] OpenAI fallback matched gpt-5.3-codex -> gpt-5.2-codex", "info"))
	require.False(t, logSink.ContainsMessageAtLevel("[Pricing] OpenAI fallback matched gpt-5.3-codex -> gpt-5.2-codex", "warn"))
}

func TestGetModelPricing_Gpt54UsesStaticFallbackWhenRemoteMissing(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": &LiteLLMModelPricing{InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.GetModelPricing("gpt-5.4")
	require.NotNil(t, got)
	require.InDelta(t, 2.5e-6, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 1.5e-5, got.OutputCostPerToken, 1e-12)
	require.InDelta(t, 2.5e-7, got.CacheReadInputTokenCost, 1e-12)
	require.Equal(t, 272000, got.LongContextInputTokenThreshold)
	require.InDelta(t, 2.0, got.LongContextInputCostMultiplier, 1e-12)
	require.InDelta(t, 1.5, got.LongContextOutputCostMultiplier, 1e-12)
}

func TestGetModelPricing_OpenAICompactAliasUsesStaticFallback(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": {InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.GetModelPricing("openai/gpt5.5")
	require.NotNil(t, got)
	require.InDelta(t, 2.5e-6, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 1.5e-5, got.OutputCostPerToken, 1e-12)
}

func TestDefaultPricingIncludesCodexAutoReview(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)

	svc := &PricingService{}
	pricingData, err := svc.parsePricingData(data)
	require.NoError(t, err)
	svc.pricingData = pricingData

	got := svc.GetModelPricing("codex-auto-review")
	require.NotNil(t, got)
	require.InDelta(t, 2.5e-6, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 1.5e-5, got.OutputCostPerToken, 1e-12)
	require.InDelta(t, 2.5e-7, got.CacheReadInputTokenCost, 1e-12)
}

func TestDefaultPricingIncludesGpt56PreviewPrices(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)

	svc := &PricingService{}
	pricingData, err := svc.parsePricingData(data)
	require.NoError(t, err)

	cases := []struct {
		model                string
		input                float64
		output               float64
		inputPriority        float64
		outputPriority       float64
		cacheRead            float64
		cacheReadPriority    float64
		cacheMake            float64
		cacheMakeBatches     float64
		cacheMakeFlex        float64
		cacheMakePriority    float64
		longContextThreshold int
		longInputMultiplier  float64
		longOutputMultiplier float64
	}{
		{model: "gpt-5.6-sol", input: 5e-6, output: 30e-6, inputPriority: 10e-6, outputPriority: 60e-6, cacheRead: 0.5e-6, cacheReadPriority: 1e-6, cacheMake: 6.25e-6, cacheMakeBatches: 3.125e-6, cacheMakeFlex: 3.125e-6, cacheMakePriority: 12.5e-6, longContextThreshold: 272000, longInputMultiplier: 2, longOutputMultiplier: 1.5},
		{model: "gpt-5.6-terra", input: 2e-6, output: 12e-6, inputPriority: 4e-6, outputPriority: 24e-6, cacheRead: 0.2e-6, cacheReadPriority: 0.4e-6, cacheMake: 2.5e-6, cacheMakeBatches: 1.25e-6, cacheMakeFlex: 1.25e-6, cacheMakePriority: 5e-6, longContextThreshold: 272000, longInputMultiplier: 2, longOutputMultiplier: 1.5},
		{model: "gpt-5.6-luna", input: 0.2e-6, output: 1.2e-6, inputPriority: 0.4e-6, outputPriority: 2.4e-6, cacheRead: 0.02e-6, cacheReadPriority: 0.04e-6, cacheMake: 0.25e-6, cacheMakeBatches: 0.125e-6, cacheMakeFlex: 0.125e-6, cacheMakePriority: 0.5e-6, longContextThreshold: 272000, longInputMultiplier: 2, longOutputMultiplier: 1.5},
	}
	for _, tc := range cases {
		t.Run(tc.model, func(t *testing.T) {
			got := pricingData[tc.model]
			require.NotNil(t, got)
			require.InDelta(t, tc.input, got.InputCostPerToken, 1e-12)
			require.InDelta(t, tc.output, got.OutputCostPerToken, 1e-12)
			require.InDelta(t, tc.inputPriority, got.InputCostPerTokenPriority, 1e-12)
			require.InDelta(t, tc.outputPriority, got.OutputCostPerTokenPriority, 1e-12)
			require.InDelta(t, tc.cacheRead, got.CacheReadInputTokenCost, 1e-12)
			require.InDelta(t, tc.cacheReadPriority, got.CacheReadInputTokenCostPriority, 1e-12)
			require.InDelta(t, tc.cacheMake, got.CacheCreationInputTokenCost, 1e-12)
			require.InDelta(t, tc.cacheMakeBatches, got.CacheCreationInputTokenCostBatches, 1e-12)
			require.InDelta(t, tc.cacheMakeFlex, got.CacheCreationInputTokenCostFlex, 1e-12)
			require.InDelta(t, tc.cacheMakePriority, got.CacheCreationInputTokenCostPriority, 1e-12)
			require.Equal(t, tc.longContextThreshold, got.LongContextInputTokenThreshold)
			require.InDelta(t, tc.longInputMultiplier, got.LongContextInputCostMultiplier, 1e-12)
			require.InDelta(t, tc.longOutputMultiplier, got.LongContextOutputCostMultiplier, 1e-12)
			require.True(t, got.SupportsPromptCaching)
			require.True(t, got.SupportsServiceTier)
		})
	}
}

func TestGetModelPricing_Gpt56PreviewUsesDedicatedStaticFallbackWhenRemoteMissing(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": {InputCostPerToken: 1.25e-6},
		},
	}

	cases := []struct {
		model     string
		input     float64
		output    float64
		cacheRead float64
		cacheMake float64
	}{
		{model: "gpt-5.6-sol", input: 5e-6, output: 30e-6, cacheRead: 0.5e-6, cacheMake: 6.25e-6},
		{model: "gpt-5.6-terra", input: 2e-6, output: 12e-6, cacheRead: 0.2e-6, cacheMake: 2.5e-6},
		{model: "gpt-5.6-luna", input: 0.2e-6, output: 1.2e-6, cacheRead: 0.02e-6, cacheMake: 0.25e-6},
	}
	for _, tc := range cases {
		t.Run(tc.model, func(t *testing.T) {
			got := svc.GetModelPricing(tc.model)
			require.NotNil(t, got)
			require.InDelta(t, tc.input, got.InputCostPerToken, 1e-12)
			require.InDelta(t, tc.input*2, got.InputCostPerTokenPriority, 1e-12)
			require.InDelta(t, tc.output, got.OutputCostPerToken, 1e-12)
			require.InDelta(t, tc.output*2, got.OutputCostPerTokenPriority, 1e-12)
			require.InDelta(t, tc.cacheRead, got.CacheReadInputTokenCost, 1e-12)
			require.InDelta(t, tc.cacheRead*2, got.CacheReadInputTokenCostPriority, 1e-12)
			require.InDelta(t, tc.cacheMake, got.CacheCreationInputTokenCost, 1e-12)
			require.InDelta(t, tc.cacheMake*2, got.CacheCreationInputTokenCostPriority, 1e-12)
			require.Equal(t, 272000, got.LongContextInputTokenThreshold)
			require.True(t, got.SupportsPromptCaching)
			require.True(t, got.SupportsServiceTier)
		})
	}
}

func TestGetModelPricing_Gpt54MiniUsesDedicatedStaticFallbackWhenRemoteMissing(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": {InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.GetModelPricing("gpt-5.4-mini")
	require.NotNil(t, got)
	require.InDelta(t, 7.5e-7, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 4.5e-6, got.OutputCostPerToken, 1e-12)
	require.InDelta(t, 7.5e-8, got.CacheReadInputTokenCost, 1e-12)
	require.Zero(t, got.LongContextInputTokenThreshold)
}

func TestGetModelPricing_Gpt54NanoUsesDedicatedStaticFallbackWhenRemoteMissing(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": {InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.GetModelPricing("gpt-5.4-nano")
	require.NotNil(t, got)
	require.InDelta(t, 2e-7, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 1.25e-6, got.OutputCostPerToken, 1e-12)
	require.InDelta(t, 2e-8, got.CacheReadInputTokenCost, 1e-12)
	require.Zero(t, got.LongContextInputTokenThreshold)
}

func TestGetModelPricing_ClaudeOpus48FallsBackToOpus46Pricing(t *testing.T) {
	opus46Pricing := &LiteLLMModelPricing{
		InputCostPerToken:           5e-6,
		OutputCostPerToken:          25e-6,
		CacheCreationInputTokenCost: 6.25e-6,
		CacheReadInputTokenCost:     0.5e-6,
	}
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"claude-opus-4-6": opus46Pricing,
		},
	}

	got := svc.GetModelPricing("claude-opus-4-8")
	require.Same(t, opus46Pricing, got)
}

func TestGetModelPricing_ClaudeDecimalAliasMatchesHyphenatedPricing(t *testing.T) {
	sonnet45Pricing := &LiteLLMModelPricing{InputCostPerToken: 3e-6}
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"claude-sonnet-4-5": sonnet45Pricing,
		},
	}

	got := svc.GetModelPricing("claude-sonnet-4.5")
	require.Same(t, sonnet45Pricing, got)
}

func TestGetModelPricing_ImageModelDoesNotFallbackToTextModel(t *testing.T) {
	imagePricing := &LiteLLMModelPricing{InputCostPerToken: 3}
	textPricing := &LiteLLMModelPricing{InputCostPerToken: 9}

	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-image-2": imagePricing,
			"gpt-5.4":     textPricing,
		},
	}

	got := svc.GetModelPricing("gpt-image-3")
	require.Same(t, imagePricing, got)
}

func TestParsePricingData_PreservesPriorityAndServiceTierFields(t *testing.T) {
	raw := map[string]any{
		"gpt-5.4": map[string]any{
			"input_cost_per_token":                 2.5e-6,
			"input_cost_per_token_priority":        5e-6,
			"output_cost_per_token":                15e-6,
			"output_cost_per_token_priority":       30e-6,
			"cache_read_input_token_cost":          0.25e-6,
			"cache_read_input_token_cost_priority": 0.5e-6,
			"supports_service_tier":                true,
			"supports_prompt_caching":              true,
			"litellm_provider":                     "openai",
			"mode":                                 "chat",
		},
	}
	body, err := json.Marshal(raw)
	require.NoError(t, err)

	svc := &PricingService{}
	pricingMap, err := svc.parsePricingData(body)
	require.NoError(t, err)

	pricing := pricingMap["gpt-5.4"]
	require.NotNil(t, pricing)
	require.InDelta(t, 2.5e-6, pricing.InputCostPerToken, 1e-12)
	require.InDelta(t, 5e-6, pricing.InputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 15e-6, pricing.OutputCostPerToken, 1e-12)
	require.InDelta(t, 30e-6, pricing.OutputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 0.25e-6, pricing.CacheReadInputTokenCost, 1e-12)
	require.InDelta(t, 0.5e-6, pricing.CacheReadInputTokenCostPriority, 1e-12)
	require.True(t, pricing.SupportsServiceTier)
}

func TestParsePricingData_PreservesServiceTierPriorityFields(t *testing.T) {
	svc := &PricingService{}
	pricingData, err := svc.parsePricingData([]byte(`{
		"gpt-5.4": {
			"input_cost_per_token": 0.0000025,
			"input_cost_per_token_priority": 0.000005,
			"output_cost_per_token": 0.000015,
			"output_cost_per_token_priority": 0.00003,
			"cache_read_input_token_cost": 0.00000025,
			"cache_read_input_token_cost_priority": 0.0000005,
			"supports_service_tier": true,
			"litellm_provider": "openai",
			"mode": "chat"
		}
	}`))
	require.NoError(t, err)

	pricing := pricingData["gpt-5.4"]
	require.NotNil(t, pricing)
	require.InDelta(t, 0.0000025, pricing.InputCostPerToken, 1e-12)
	require.InDelta(t, 0.000005, pricing.InputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 0.000015, pricing.OutputCostPerToken, 1e-12)
	require.InDelta(t, 0.00003, pricing.OutputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 0.00000025, pricing.CacheReadInputTokenCost, 1e-12)
	require.InDelta(t, 0.0000005, pricing.CacheReadInputTokenCostPriority, 1e-12)
	require.True(t, pricing.SupportsServiceTier)
}
