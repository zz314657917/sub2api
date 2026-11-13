package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeminiPricing20261111(t *testing.T) {
	cards := []struct {
		name                                                      string
		input, output, cache, imageInput, imageOutput, imageToken float64
	}{
		{"gemini-3-flash", 5e-7, 3e-6, 5e-8, 0, 0, 0},
		{"gemini-3-pro-image", 2e-6, 1.2e-5, 0, .0011, .134, .00012},
		{"gemini-3.1-flash-image", 5e-7, 3e-6, 0, .00056, .0672, .00006},
		{"gemini-3.1-flash-image-preview", 5e-7, 3e-6, 0, .00056, .0672, .00006},
		{"gemini-3.1-flash-lite-image", 3e-7, 2.5e-6, 0, .0003, .034, .00003},
		{"gemini-3.1-flash-lite", 2.5e-7, 1.5e-6, 2.5e-8, 0, 0, 0},
		{"gemini-3.1-flash-lite-preview", 2.5e-7, 1.5e-6, 2.5e-8, 0, 0, 0},
		{"gemini-3.5-flash", 1.5e-6, 9e-6, 1.5e-7, 0, 0, 0},
	}
	for _, tc := range cards {
		t.Run(tc.name, func(t *testing.T) {
			for _, prefix := range []string{"", "models/", "google/", "antigravity/", "publishers/google/models/", "projects/p/locations/l/publishers/google/models/"} {
				model := prefix + tc.name
				stale := &PricingService{pricingData: map[string]*LiteLLMModelPricing{}}
				p := stale.GetModelPricing(model)
				require.NotNil(t, p)
				require.Equal(t, tc.input, p.InputCostPerToken)
				require.Equal(t, tc.output, p.OutputCostPerToken)
				require.Equal(t, tc.cache, p.CacheReadInputTokenCost)
				require.Equal(t, tc.imageInput, p.InputCostPerImage)
				require.Equal(t, tc.imageOutput, p.OutputCostPerImage)
				require.Equal(t, tc.imageToken, p.OutputCostPerImageToken)
				require.Zero(t, p.CacheCreationInputTokenCost)
				for _, pricing := range []*PricingService{nil, stale} {
					billing := &BillingService{pricingService: pricing}
					got, err := billing.GetModelPricing(model)
					require.NoError(t, err)
					require.Equal(t, tc.input, got.InputPricePerToken)
					require.Equal(t, tc.output, got.OutputPricePerToken)
					require.Equal(t, tc.imageToken, got.ImageOutputPricePerToken)
				}
			}
			for _, input := range []float64{0, .123} {
				remote := &LiteLLMModelPricing{InputCostPerToken: input}
				for _, key := range []string{tc.name, "google/" + tc.name} {
					pricing := &PricingService{pricingData: map[string]*LiteLLMModelPricing{key: remote}}
					require.Same(t, remote, pricing.GetModelPricing("google/"+tc.name))
					billing := &BillingService{pricingService: pricing}
					got, err := billing.GetModelPricing("google/" + tc.name)
					require.NoError(t, err)
					require.Equal(t, input, got.InputPricePerToken)
				}
			}
			pricing := &PricingService{pricingData: map[string]*LiteLLMModelPricing{tc.name: {InputCostPerToken: tc.input}}}
			for _, suffix := range []string{"-unknown", "-20261111", "x"} {
				require.Nil(t, pricing.GetModelPricing(tc.name+suffix))
				_, err := (&BillingService{pricingService: pricing}).GetModelPricing(tc.name + suffix)
				require.Error(t, err)
			}
		})
	}
}

func geminiPricing20261111Bundle(t *testing.T) *PricingService {
	t.Helper()
	body, err := os.ReadFile("../../resources/model-pricing/model_prices_and_context_window.json")
	require.NoError(t, err)
	s := &PricingService{}
	s.pricingData, err = s.parsePricingData(body)
	require.NoError(t, err)
	return s
}

func TestGeminiPricing20261111BundledCards(t *testing.T) {
	bundle := geminiPricing20261111Bundle(t)
	cards := []struct {
		name                                                               string
		input, output, cache, priorityInput, priorityOutput, priorityCache float64
		imageInput, imageOutput, imageToken                                float64
	}{
		{"gemini-3-flash", 5e-7, 3e-6, 5e-8, 9e-7, 5.4e-6, 9e-8, 0, 0, 0},
		{"gemini-3-pro-image", 2e-6, 1.2e-5, 0, 0, 0, 0, .0011, .134, .00012},
		{"gemini-3.1-flash-image", 5e-7, 3e-6, 0, 0, 0, 0, .00056, .0672, .00006},
		{"gemini-3.1-flash-image-preview", 5e-7, 3e-6, 0, 0, 0, 0, .00056, .0672, .00006},
		{"gemini-3.1-flash-lite-image", 3e-7, 2.5e-6, 0, 0, 0, 0, .0003, .034, .00003},
		{"gemini-3.1-flash-lite", 2.5e-7, 1.5e-6, 2.5e-8, 4.5e-7, 2.7e-6, 4.5e-8, 0, 0, 0},
		{"gemini-3.1-flash-lite-preview", 2.5e-7, 1.5e-6, 2.5e-8, 0, 0, 0, 0, 0, 0},
		{"gemini-3.5-flash", 1.5e-6, 9e-6, 1.5e-7, 2.7e-6, 1.62e-5, 2.7e-7, 0, 0, 0},
	}
	for _, tc := range cards {
		t.Run(tc.name, func(t *testing.T) {
			// Read the parsed entry directly so fallback cannot hide a missing catalog card.
			p := bundle.pricingData[tc.name]
			require.NotNil(t, p)
			fallback, reserved := geminiUpstreamDefaultPricing(tc.name)
			require.True(t, reserved)
			require.NotNil(t, fallback)
			want := []float64{tc.input, tc.output, tc.cache, tc.priorityInput, tc.priorityOutput, tc.priorityCache, tc.imageInput, tc.imageOutput, tc.imageToken}
			for _, card := range []*LiteLLMModelPricing{p, fallback} {
				got := []float64{card.InputCostPerToken, card.OutputCostPerToken, card.CacheReadInputTokenCost, card.InputCostPerTokenPriority, card.OutputCostPerTokenPriority, card.CacheReadInputTokenCostPriority, card.InputCostPerImage, card.OutputCostPerImage, card.OutputCostPerImageToken}
				for i := range want {
					require.InDelta(t, want[i], got[i], 1e-15)
				}
				require.Zero(t, card.CacheCreationInputTokenCost)
				require.False(t, card.CacheCreationInputTokenCostExplicit)
				require.Equal(t, tc.priorityInput != 0, card.InputCostPerTokenPriorityExplicit)
				require.Equal(t, tc.priorityOutput != 0, card.OutputCostPerTokenPriorityExplicit)
				require.Equal(t, tc.priorityCache != 0, card.CacheReadInputTokenCostPriorityExplicit)
			}
			billing := &BillingService{pricingService: bundle}
			got, err := billing.GetModelPricing("google/" + tc.name)
			require.NoError(t, err)
			require.InDelta(t, tc.input, got.InputPricePerToken, 1e-15)
			require.InDelta(t, tc.output, got.OutputPricePerToken, 1e-15)
			require.InDelta(t, tc.priorityInput, got.InputPricePerTokenPriority, 1e-15)
			require.InDelta(t, tc.priorityOutput, got.OutputPricePerTokenPriority, 1e-15)
			require.InDelta(t, tc.priorityCache, got.CacheReadPricePerTokenPriority, 1e-15)
		})
	}
}

func TestGeminiPricing20261111ImageCosts(t *testing.T) {
	bundle := geminiPricing20261111Bundle(t)
	stale := &PricingService{pricingData: map[string]*LiteLLMModelPricing{}}
	cards := []struct {
		name  string
		price float64
	}{
		{"gemini-3-pro-image", .134},
		{"gemini-3.1-flash-image", .0672},
		{"gemini-3.1-flash-image-preview", .0672},
		{"gemini-3.1-flash-lite-image", .034},
	}
	for _, tc := range cards {
		t.Run(tc.name, func(t *testing.T) {
			for _, pricing := range []*PricingService{nil, stale, bundle} {
				billing := &BillingService{pricingService: pricing}
				for _, prefix := range []string{"", "models/", "google/", "antigravity/", "projects/p/locations/l/publishers/google/models/"} {
					for _, tier := range []struct {
						name   string
						factor float64
					}{{"1K", 1}, {"2K", 1.5}, {"4K", 2}} {
						got := billing.CalculateImageCost(prefix+tc.name, tier.name, 2, nil, 3)
						require.InDelta(t, tc.price*tier.factor*2, got.TotalCost, 1e-12)
						require.InDelta(t, tc.price*tier.factor*6, got.ActualCost, 1e-12)
						for _, override := range []float64{0, .321} {
							group := &ImagePriceConfig{Price1K: &override, Price2K: &override, Price4K: &override}
							got = billing.CalculateImageCost(prefix+tc.name, tier.name, 2, group, 3)
							require.InDelta(t, override*2, got.TotalCost, 1e-12)
							require.InDelta(t, override*6, got.ActualCost, 1e-12)
						}
					}
				}
			}
		})
	}
}

func TestGeminiPricing20261111RemoteImagePrecedence(t *testing.T) {
	for _, tc := range []struct {
		field string
		want  float64
	}{
		{`,"output_cost_per_image":0`, 0},
		{`,"output_cost_per_image":0.42`, .42},
		{"", .0672},
	} {
		for _, key := range []string{"gemini-3.1-flash-image", "google/gemini-3.1-flash-image"} {
			s := &PricingService{}
			var err error
			s.pricingData, err = s.parsePricingData([]byte(`{"` + key + `":{"input_cost_per_token":0,"output_cost_per_token":0,"input_cost_per_image":0.123` + tc.field + `}}`))
			require.NoError(t, err)
			p := s.GetModelPricing("google/gemini-3.1-flash-image")
			require.Same(t, s.pricingData[key], p)
			require.InDelta(t, .123, p.InputCostPerImage, 1e-15)
			require.Equal(t, tc.field != "", p.OutputCostPerImageExplicit)
			billing := &BillingService{pricingService: s}
			got, err := billing.GetModelPricing("google/gemini-3.1-flash-image")
			require.NoError(t, err)
			require.Zero(t, got.InputPricePerToken)
			require.Zero(t, got.OutputPricePerToken)
			cost := billing.CalculateImageCost("google/gemini-3.1-flash-image", "1K", 1, nil, 1)
			require.InDelta(t, tc.want, cost.TotalCost, 1e-15)
		}
	}
	// Literal provider pricing precedes normalized canonical pricing, even at zero.
	s := &PricingService{}
	var err error
	s.pricingData, err = s.parsePricingData([]byte(`{"gemini-3.1-flash-image":{"input_cost_per_token":0.2,"output_cost_per_image":0.8},"google/gemini-3.1-flash-image":{"input_cost_per_token":0,"output_cost_per_image":0}}`))
	require.NoError(t, err)
	require.Same(t, s.pricingData["google/gemini-3.1-flash-image"], s.GetModelPricing("google/gemini-3.1-flash-image"))
	require.Zero(t, (&BillingService{pricingService: s}).CalculateImageCost("google/gemini-3.1-flash-image", "1K", 1, nil, 1).TotalCost)
	// Unknown siblings have no static card, but explicit remote configuration works.
	unknown := "google/gemini-3.1-flash-image-custom"
	custom := &LiteLLMModelPricing{InputCostPerToken: .7}
	s.pricingData[unknown] = custom
	require.Same(t, custom, s.GetModelPricing(unknown))
	got, err := (&BillingService{pricingService: s}).GetModelPricing(unknown)
	require.NoError(t, err)
	require.Equal(t, .7, got.InputPricePerToken)
}

func TestGeminiPricing20261111PriorityTokenCosts(t *testing.T) {
	models := []string{
		"gemini-3-flash",
		"gemini-3-pro-image",
		"gemini-3.1-flash-image",
		"gemini-3.1-flash-image-preview",
		"gemini-3.1-flash-lite-image",
		"gemini-3.1-flash-lite",
		"gemini-3.1-flash-lite-preview",
		"gemini-3.5-flash",
	}
	prefixes := []string{"", "google/", "publishers/google/models/", "projects/p/locations/l/publishers/google/models/"}
	tokens := UsageTokens{InputTokens: 1000000, OutputTokens: 1000000, CacheReadTokens: 1000000}
	for _, model := range models {
		for _, prefix := range prefixes {
			for _, tc := range []struct {
				name   string
				fields string
				want   float64
			}{
				{"all_zero", `,"input_cost_per_token_priority":0,"output_cost_per_token_priority":0,"cache_read_input_token_cost_priority":0`, 0},
				{"output_only", `,"input_cost_per_token_priority":0,"output_cost_per_token_priority":0.0000054,"cache_read_input_token_cost_priority":0`, 5.4},
			} {
				t.Run(prefix+model+"/"+tc.name, func(t *testing.T) {
					ps := &PricingService{}
					var err error
					ps.pricingData, err = ps.parsePricingData([]byte(`{"` + model + `":{"input_cost_per_token":0.0000005,"output_cost_per_token":0.000003,"cache_read_input_token_cost":0.00000005` + tc.fields + `}}`))
					require.NoError(t, err)
					bs := &BillingService{pricingService: ps}
					p, err := bs.GetModelPricing(prefix + model)
					require.NoError(t, err)
					require.True(t, p.InputPricePriorityExplicit)
					require.True(t, p.OutputPricePriorityExplicit)
					require.True(t, p.CacheReadPricePriorityExplicit)
					require.False(t, p.StrictPriorityTier)
					cost := bs.computeTokenBreakdown(p, tokens, 1, "priority", false)
					if tc.want == 0 {
						require.Zero(t, cost.TotalCost)
					} else {
						require.InDelta(t, tc.want, cost.TotalCost, 1e-12)
					}
				})
			}
		}
	}
	bundle := geminiPricing20261111Bundle(t)
	for _, ps := range []*PricingService{nil, {pricingData: map[string]*LiteLLMModelPricing{}}, bundle} {
		bs := &BillingService{pricingService: ps}
		for _, prefix := range prefixes {
			p, err := bs.GetModelPricing(prefix + "gemini-3-flash")
			require.NoError(t, err)
			require.InDelta(t, 6.39, bs.computeTokenBreakdown(p, tokens, 1, "priority", false).TotalCost, 1e-12)
			preview, err := bs.GetModelPricing(prefix + "gemini-3.1-flash-lite-preview")
			require.NoError(t, err)
			require.False(t, preview.InputPricePriorityExplicit)
			require.False(t, preview.OutputPricePriorityExplicit)
			require.False(t, preview.CacheReadPricePriorityExplicit)
			normal := bs.computeTokenBreakdown(preview, tokens, 1, "", false)
			require.InDelta(t, 1.775, normal.TotalCost, 1e-12)
			require.InDelta(t, 2*normal.TotalCost, bs.computeTokenBreakdown(preview, tokens, 1, "priority", false).TotalCost, 1e-12)
			zero := 0.0
			channel := &ChannelModelPricing{InputPrice: &zero, OutputPrice: &zero}
			channelPricing, err := bs.GetModelPricingWithChannel(prefix+"gemini-3-flash", channel)
			require.NoError(t, err)
			for _, tier := range []string{"", "priority"} {
				cost := bs.computeTokenBreakdown(channelPricing, UsageTokens{InputTokens: 1000000, OutputTokens: 1000000}, 1, tier, false)
				require.Zero(t, cost.TotalCost, "channel zero must survive tier %q", tier)
			}
		}
	}
}

func TestGeminiPricing20261111UnknownPriorityZeroPolicy(t *testing.T) {
	ps := &PricingService{}
	var err error
	ps.pricingData, err = ps.parsePricingData([]byte(`{"gemini-3-flash-custom":{"input_cost_per_token":0.0000005,"output_cost_per_token":0.000003,"cache_read_input_token_cost":0.00000005,"input_cost_per_token_priority":0,"output_cost_per_token_priority":0,"cache_read_input_token_cost_priority":0}}`))
	require.NoError(t, err)
	bs := &BillingService{pricingService: ps}
	for _, prefix := range []string{"", "google/", "projects/p/locations/l/publishers/google/models/"} {
		model := prefix + "gemini-3-flash-custom"
		fallback, _ := geminiUpstreamDefaultPricing(model)
		require.Nil(t, fallback)
		p, err := bs.GetModelPricing(model)
		require.NoError(t, err)
		require.False(t, p.InputPricePriorityExplicit)
		require.False(t, p.OutputPricePriorityExplicit)
		require.False(t, p.CacheReadPricePriorityExplicit)
		cost := bs.computeTokenBreakdown(p, UsageTokens{InputTokens: 1000000, OutputTokens: 1000000, CacheReadTokens: 1000000}, 1, "priority", false)
		require.InDelta(t, 7.1, cost.TotalCost, 1e-12)
	}
}

func TestGeminiPricing20261111ChannelOverrides(t *testing.T) {
	bundle := geminiPricing20261111Bundle(t)
	for _, pricing := range []*PricingService{nil, bundle} {
		billing := &BillingService{pricingService: pricing}
		for _, value := range []float64{0, .123} {
			channel := &ChannelModelPricing{InputPrice: &value, OutputPrice: &value, ImageInputPrice: &value, ImageOutputPrice: &value}
			got, err := billing.GetModelPricingWithChannel("google/gemini-3.1-flash-image", channel)
			require.NoError(t, err)
			require.Equal(t, value, got.InputPricePerToken)
			require.Equal(t, value, got.OutputPricePerToken)
			require.Equal(t, value, got.ImageInputPricePerToken)
			require.Equal(t, value, got.ImageOutputPricePerToken)
		}
		got, err := billing.GetModelPricing("google/gemini-3.1-flash-image")
		require.NoError(t, err)
		require.InDelta(t, 5e-7, got.InputPricePerToken, 1e-15)
		require.InDelta(t, .00006, got.ImageOutputPricePerToken, 1e-15)
	}
}
