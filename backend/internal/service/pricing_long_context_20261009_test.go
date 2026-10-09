package service

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func pricing20261009(t *testing.T, body string) *PricingService {
	t.Helper()
	s := &PricingService{}
	data, err := s.parsePricingData([]byte(body))
	require.NoError(t, err)
	s.pricingData = data
	return s
}

func bundle20261009(t *testing.T) *PricingService {
	t.Helper()
	body, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	return pricing20261009(t, string(body))
}

func TestPricingLongContext20261009MetadataTruthTable(t *testing.T) {
	for _, tc := range []struct {
		name, fields  string
		present       bool
		threshold     int
		input, output float64
	}{
		{"absent", "", false, 272000, 2, 1.5},
		{"above", `,"input_cost_per_token_above_200k_tokens":4e-6,"output_cost_per_token_above_200k_tokens":15e-6`, true, 200000, 2, 1.5},
		{"lowest", `,"input_cost_per_token_above_272k_tokens":8e-6,"output_cost_per_token_above_100k_tokens":20e-6`, true, 100000, 1, 2},
		{"nonraised", `,"input_cost_per_token_above_200k_tokens":1e-6,"output_cost_per_token_above_200k_tokens":10e-6`, true, 200000, 1, 1},
		{"above-zero", `,"input_cost_per_token_above_200k_tokens":0`, true, 200000, 1, 1},
		{"ignored", `,"input_cost_per_token_above_200k_tokens_priority":9e-6,"output_cost_per_token_above_200k_tokens_flex":9e-6,"cache_read_input_token_cost_above_200k_tokens":9e-6`, false, 272000, 2, 1.5},
		{"disabled", `,"long_context_input_token_threshold":0,"input_cost_per_token_above_200k_tokens":4e-6`, true, 0, 1, 1},
		{"disabled-positive", `,"long_context_input_token_threshold":0,"long_context_input_cost_multiplier":2,"long_context_output_cost_multiplier":1.5,"input_cost_per_token_above_200k_tokens":4e-6`, true, 0, 2, 1.5},
		{"negative", `,"long_context_input_token_threshold":-1,"long_context_input_cost_multiplier":2`, true, -1, 2, 1},
		{"threshold-only", `,"long_context_input_token_threshold":200000`, true, 200000, 1, 1},
		{"zeros", `,"long_context_input_token_threshold":200000,"long_context_input_cost_multiplier":0,"long_context_output_cost_multiplier":0`, true, 200000, 1, 1},
		{"ones", `,"long_context_input_token_threshold":200000,"long_context_input_cost_multiplier":1,"long_context_output_cost_multiplier":1`, true, 200000, 1, 1},
		{"output-only", `,"long_context_input_token_threshold":200000,"long_context_input_cost_multiplier":0,"long_context_output_cost_multiplier":1.5`, true, 200000, 1, 1.5},
		{"negative-side", `,"long_context_input_token_threshold":200000,"long_context_input_cost_multiplier":-2,"long_context_output_cost_multiplier":1.5`, true, 200000, 1, 1.5},
		{"input-only", `,"long_context_input_token_threshold":200000,"long_context_input_cost_multiplier":2`, true, 200000, 2, 1},
		{"multiplier-no-threshold", `,"long_context_input_cost_multiplier":2,"input_cost_per_token_above_200k_tokens":4e-6`, true, 0, 2, 1},
		{"explicit-wins", `,"long_context_input_token_threshold":272000,"long_context_input_cost_multiplier":1,"output_cost_per_token_above_200k_tokens":30e-6`, true, 272000, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ps := pricing20261009(t, `{"gpt-6.1-sol":{"input_cost_per_token":2e-6,"output_cost_per_token":10e-6`+tc.fields+`}}`)
			raw := ps.GetModelPricing("gpt-6.1-sol")
			require.Equal(t, tc.present, raw.LongContextMetadataPresent)
			b := NewBillingService(&config.Config{}, ps)
			p, err := b.GetModelPricing("openai/gpt-6.1-sol-high")
			require.NoError(t, err)
			require.Equal(t, tc.present, p.LongContextMetadataPresent)
			require.Equal(t, tc.threshold, p.LongContextInputThreshold)
			require.Equal(t, tc.input, longContextMultiplierOrOne(p.LongContextInputMultiplier))
			require.Equal(t, tc.output, longContextMultiplierOrOne(p.LongContextOutputMultiplier))
			customInput := 2e-6
			resolver := NewModelPricingResolver(nil, b)
			resolved := resolver.Resolve(nil, PricingInput{Model: "gpt-6.1-sol", Group: &Group{
				LongContextPricingEnabled: true, ModelPricing: []ChannelModelPricing{{
					Models: []string{"gpt-6.1-sol"}, InputPrice: &customInput,
				}},
			}})
			require.Equal(t, tc.present, resolved.BasePricing.LongContextMetadataPresent)
			require.Equal(t, tc.threshold, resolved.BasePricing.LongContextInputThreshold)
			cost, err := b.CalculateCost("openai/gpt-6.1-sol-high", UsageTokens{InputTokens: 300000, OutputTokens: 1000}, 1)
			require.NoError(t, err)
			im, om := 1.0, 1.0
			if tc.threshold > 0 {
				im, om = tc.input, tc.output
			}
			require.InDelta(t, .6*im, cost.InputCost, 1e-12)
			require.InDelta(t, .01*om, cost.OutputCost, 1e-12)
		})
	}
}

func TestPricingLongContext20261009SourcesAliasesBoundaries(t *testing.T) {
	dynamic := pricing20261009(t, `{
		"gpt-6.1-sol":{"input_cost_per_token":2e-6,"output_cost_per_token":10e-6,"cache_read_input_token_cost":1e-7,"input_cost_per_token_above_272k_tokens":4e-6,"output_cost_per_token_above_272k_tokens":15e-6},
		"grok-4.7":{"input_cost_per_token":2e-6,"output_cost_per_token":6e-6,"cache_read_input_token_cost":5e-7,"litellm_provider":"xai","input_cost_per_token_above_200k_tokens":4e-6,"output_cost_per_token_above_200k_tokens":12e-6}}`)
	for source, ps := range map[string]*PricingService{"dynamic": dynamic, "bundle": bundle20261009(t), "static": {}, "hardcoded": nil} {
		t.Run(source, func(t *testing.T) {
			b := NewBillingService(&config.Config{}, ps)
			for _, model := range []string{"gpt-6.1-sol", "openai/gpt-6.1-sol-high", "gpt-6.1-sol-openai-compact", "OpenAI/GPT_6.1_SOL_MAX"} {
				p, err := b.GetModelPricing(model)
				require.NoError(t, err)
				require.InDelta(t, .1e-6, p.CacheReadPricePerToken, 1e-18)
				for _, n := range []int{271999, 272000, 272001} {
					c, err := b.CalculateCost(model, UsageTokens{InputTokens: n, OutputTokens: 1000}, 1)
					require.NoError(t, err)
					m, out := 1.0, .01
					if n > 272000 {
						m, out = 2, .015
					}
					require.InDelta(t, float64(n)*2e-6*m, c.InputCost, 1e-12)
					require.InDelta(t, out, c.OutputCost, 1e-12)
					require.Equal(t, n > 272000, c.LongContextBillingApplied)
				}
			}
			for _, model := range []string{"grok-4.7", "xai/grok-4.7", "x-ai/grok-4.7-latest", "grok/grok-4.7-latest"} {
				p, err := b.GetModelPricing(model)
				require.NoError(t, err)
				require.InDelta(t, .5e-6, p.CacheReadPricePerToken, 1e-18)
				for _, n := range []int{199999, 200000, 200001} {
					c, err := b.CalculateCost(model, UsageTokens{InputTokens: n - 1000, CacheReadTokens: 1000, OutputTokens: 1000}, 1)
					require.NoError(t, err)
					m := 1.0
					if n >= 200000 {
						m = 2
					}
					require.InDelta(t, float64(n-1000)*2e-6*m, c.InputCost, 1e-12)
					require.InDelta(t, .0005*m, c.CacheReadCost, 1e-12)
					require.InDelta(t, .006*m, c.OutputCost, 1e-12)
				}
			}
		})
	}
}

func TestPricingLongContext20261009BundleCoverage(t *testing.T) {
	ps := bundle20261009(t)
	fallback := NewBillingService(&config.Config{}, nil)
	for _, id := range []string{"gpt-6.1-sol", "claude-fable-5-1", "claude-opus-5", "claude-opus-5-5", "glm-5.3", "glm-5.3-flash", "grok-4.7", "kimi-k3"} {
		require.NotNil(t, ps.pricingData[id], id)
		got, err := NewBillingService(&config.Config{}, ps).GetModelPricing(id)
		require.NoError(t, err)
		want, err := fallback.GetModelPricing(id)
		require.NoError(t, err)
		require.InDelta(t, want.InputPricePerToken, got.InputPricePerToken, 1e-18, id)
		require.InDelta(t, want.OutputPricePerToken, got.OutputPricePerToken, 1e-18, id)
		require.InDelta(t, want.CacheReadPricePerToken, got.CacheReadPricePerToken, 1e-18, id)
	}
	for _, id := range []string{"deepseek-v3.2-exp", "gemini-3.1-flash"} {
		require.Nil(t, ps.pricingData[id])
		_, err := fallback.GetModelPricing(id)
		require.Error(t, err)
	}
	// The resource remains valid JSON; no unrelated card gets rewritten.
	raw, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	require.True(t, json.Valid(raw))
}

func TestPricingLongContext20261009LegacyDefaultsAndOriginalAliases(t *testing.T) {
	for _, model := range []string{"gpt-5.4", "gpt-5.5", "gpt-5.6-sol", "gpt-6-sol", "gpt-6-luna", "gpt-6-astra", "gpt-6.1-sol"} {
		body, err := json.Marshal(map[string]any{model: map[string]any{
			"input_cost_per_token": 2e-6, "output_cost_per_token": 10e-6,
		}})
		require.NoError(t, err)
		b := NewBillingService(&config.Config{}, pricing20261009(t, string(body)))
		card, err := b.GetModelPricing(model)
		require.NoError(t, err)
		require.False(t, card.LongContextMetadataPresent)
		require.Equal(t, 272000, card.LongContextInputThreshold)
		require.Equal(t, 2.0, card.LongContextInputMultiplier)
		require.Equal(t, 1.5, card.LongContextOutputMultiplier)
	}
	// Keep literal aliases ahead of normalized family cards.
	ps := pricing20261009(t, `{
		"openai/gpt-6.1-sol-high":{"input_cost_per_token":7e-6,"output_cost_per_token":11e-6,"long_context_input_token_threshold":0},
		"gpt-6.1-sol":{"input_cost_per_token":2e-6,"output_cost_per_token":10e-6,"input_cost_per_token_above_272k_tokens":4e-6},
		"gpt-6-astra":{"input_cost_per_token":10e-6,"output_cost_per_token":50e-6,"long_context_input_token_threshold":0}}`)
	b := NewBillingService(&config.Config{}, ps)
	c, err := b.CalculateCost("openai/gpt-6.1-sol-high", UsageTokens{InputTokens: 300000, OutputTokens: 1000}, 1)
	require.NoError(t, err)
	require.InDelta(t, 2.1, c.InputCost, 1e-12)
	require.False(t, c.LongContextBillingApplied)
	for _, alias := range []string{"gpt-6", "openai/gpt-6-astra-high", "gpt-6-astra-openai-compact"} {
		card, err := b.GetModelPricing(alias)
		require.NoError(t, err)
		require.True(t, card.LongContextMetadataPresent)
		require.Zero(t, card.LongContextInputThreshold)
		require.Equal(t, 6.0, card.UltrafastMultiplier)
	}
}

func TestPricingLongContext20261009GeminiCacheWriteDataContract(t *testing.T) {
	// Independent constants from upstream e2cfaa46e, not computed from the card
	// under test: cache-write is the corresponding standard/above input price.
	body, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	var cards map[string]map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(body, &cards))
	for _, tc := range []struct {
		model       string
		base, above float64
		hasPriority bool
	}{
		{"gemini-2.5-pro", 1.25e-6, 2.5e-6, false},
		{"gemini-3-pro-preview", 2e-6, 4e-6, true},
		{"gemini-3.1-pro-high", 2e-6, 4e-6, true},
		{"gemini-3.1-pro-low", 2e-6, 4e-6, true},
		{"gemini-3.1-pro-preview", 2e-6, 4e-6, true},
		{"gemini-3.1-pro-preview-customtools", 2e-6, 4e-6, false},
	} {
		t.Run(tc.model, func(t *testing.T) {
			card := cards[tc.model]
			require.NotNil(t, card)
			check := func(key string, expected float64) {
				t.Helper()
				raw, ok := card[key]
				require.True(t, ok, key)
				var value float64
				require.NoError(t, json.Unmarshal(raw, &value))
				require.InDelta(t, expected, value, 1e-18, key)
			}
			check("cache_creation_input_token_cost", tc.base)
			check("cache_creation_input_token_cost_above_200k_tokens", tc.above)
			check("input_cost_per_token", tc.base)
			check("input_cost_per_token_above_200k_tokens", tc.above)
			if tc.hasPriority {
				check("cache_creation_input_token_cost_priority", 3.6e-6)
				check("cache_creation_input_token_cost_above_200k_tokens_priority", 7.2e-6)
				check("input_cost_per_token_priority", 3.6e-6)
				check("input_cost_per_token_above_200k_tokens_priority", 7.2e-6)
			} else {
				require.NotContains(t, card, "cache_creation_input_token_cost_priority")
				require.NotContains(t, card, "cache_creation_input_token_cost_above_200k_tokens_priority")
			}
		})
	}
}
