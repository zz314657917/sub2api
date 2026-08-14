package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestGetModelPricing_Grok45FallbackCard(t *testing.T) {
	svc := NewBillingService(&config.Config{}, nil)

	pricing, err := svc.GetModelPricing("grok-4.5")
	require.NoError(t, err)
	require.InDelta(t, 2e-6, pricing.InputPricePerToken, 1e-12)
	require.InDelta(t, 0.3e-6, pricing.CacheReadPricePerToken, 1e-12)
	require.InDelta(t, 6e-6, pricing.OutputPricePerToken, 1e-12)
	require.Equal(t, 200000, pricing.LongContextInputThreshold)
	require.InDelta(t, 2.0, pricing.LongContextInputMultiplier, 1e-12)
	require.InDelta(t, 2.0, pricing.LongContextOutputMultiplier, 1e-12)

	atBoundary, err := svc.CalculateCost("grok-4.5", UsageTokens{
		InputTokens:         100000,
		CacheCreationTokens: 50000,
		CacheReadTokens:     50000,
		OutputTokens:        100,
	}, 1)
	require.NoError(t, err)
	require.InDelta(t, float64(100000)*2e-6, atBoundary.InputCost, 1e-12)
	require.InDelta(t, float64(50000)*0.3e-6, atBoundary.CacheReadCost, 1e-12)
	require.InDelta(t, float64(100)*6e-6, atBoundary.OutputCost, 1e-12)

	aboveBoundary, err := svc.CalculateCost("grok-4.5", UsageTokens{
		InputTokens:         100001,
		CacheCreationTokens: 50000,
		CacheReadTokens:     50000,
		OutputTokens:        100,
	}, 1)
	require.NoError(t, err)
	require.InDelta(t, float64(100001)*2e-6*2, aboveBoundary.InputCost, 1e-12)
	require.InDelta(t, float64(50000)*0.3e-6*2, aboveBoundary.CacheReadCost, 1e-12)
	require.InDelta(t, float64(50000)*2e-6*2, aboveBoundary.CacheCreationCost, 1e-12)
	require.InDelta(t, float64(100)*6e-6*2, aboveBoundary.OutputCost, 1e-12)
}

func TestGetModelPricing_UnknownGrokTextFallback(t *testing.T) {
	svc := NewBillingService(&config.Config{}, nil)
	for _, tc := range []struct {
		model     string
		inputRate float64
	}{
		{"grok-5", 2e-6},
		{"grok-5-latest", 2e-6},
		{"grok-4.7-20260814", 2e-6},
		{"grok-build-next", 2e-6},
		{"grok-composer-3", 2e-6},
		{"composer-next", 2e-6},
		{"xai/grok-4.3", 1.25e-6},
		{"x-ai/grok-build-0.1", 1e-6},
		{"grok/grok-5", 2e-6},
		{"GROK/GROK-4.7", 2e-6},
		{"xai/grok/grok-5", 2e-6},
	} {
		t.Run(tc.model, func(t *testing.T) {
			pricing, err := svc.GetModelPricing(tc.model)
			require.NoError(t, err)
			require.InDelta(t, tc.inputRate, pricing.InputPricePerToken, 1e-12)
		})
	}

	for _, model := range []string{"grok", "grok-latest", "xai/grok", "x-ai/grok-latest", "grok/grok"} {
		t.Run(model, func(t *testing.T) {
			pricing, err := svc.GetModelPricing(model)
			require.NoError(t, err)
			require.InDelta(t, 1.25e-6, pricing.InputPricePerToken, 1e-12)
		})
	}

	for _, model := range []string{
		"grok-2-image-1212", "grok-5-video", "grok-2-voice", "grok-5-search", "grok-web-search", "grok-x-search",
		"grok-2-audio", "grok-speech-1", "grok-transcribe-1", "grok-realtime-1", "xai/grok-6-image",
	} {
		t.Run(model, func(t *testing.T) {
			pricing, err := svc.GetModelPricing(model)
			require.Nil(t, pricing)
			require.ErrorIs(t, err, ErrModelPricingUnavailable)
		})
	}

	dynamic := NewBillingService(&config.Config{}, &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"grok-5": {InputCostPerToken: 9e-6, OutputCostPerToken: 10e-6},
	}})
	pricing, err := dynamic.GetModelPricing("grok-5")
	require.NoError(t, err)
	require.InDelta(t, 9e-6, pricing.InputPricePerToken, 1e-12)
	require.InDelta(t, 10e-6, pricing.OutputPricePerToken, 1e-12)
}
