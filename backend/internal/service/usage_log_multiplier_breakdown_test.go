package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUsageRateMultiplierBreakdown(t *testing.T) {
	t.Parallel()

	officialModel := apimartGPTImage2OfficialModel
	tests := []struct {
		name           string
		log            *UsageLog
		wantPricing    float64
		wantConversion float64
	}{
		{
			name:        "official image uses model trigger",
			log:         &UsageLog{ImageCount: 1, RateMultiplier: 8.4, RequestedModel: officialModel},
			wantPricing: 1, wantConversion: 8.4,
		},
		{
			name:        "official image preserves configured two times rate",
			log:         &UsageLog{ImageCount: 1, RateMultiplier: 16.8, Model: officialModel},
			wantPricing: 2, wantConversion: 8.4,
		},
		{
			name:        "mapped upstream official image uses model trigger",
			log:         &UsageLog{ImageCount: 1, RateMultiplier: 8.4, Model: "user-image", UpstreamModel: &officialModel},
			wantPricing: 1, wantConversion: 8.4,
		},
		{
			name: "apimart account applies to any image model",
			log: &UsageLog{ImageCount: 1, RateMultiplier: 8.4, Model: "gpt-image-2", Account: &Account{
				Platform:    PlatformOpenAI,
				Type:        AccountTypeAPIKey,
				Credentials: map[string]any{"base_url": "https://api.apimart.ai/v1"},
			}},
			wantPricing: 1, wantConversion: 8.4,
		},
		{
			name:        "ordinary OpenAI image remains composite pricing",
			log:         &UsageLog{ImageCount: 1, RateMultiplier: 2, Model: "gpt-image-2"},
			wantPricing: 2, wantConversion: 1,
		},
		{
			name:        "non image official model is not split",
			log:         &UsageLog{ImageCount: 0, RateMultiplier: 8.4, Model: officialModel},
			wantPricing: 8.4, wantConversion: 1,
		},
		{
			name:        "historical official image without account still uses immutable model",
			log:         &UsageLog{ImageCount: 1, RateMultiplier: 8.4, Model: officialModel},
			wantPricing: 1, wantConversion: 8.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pricing, conversion := UsageRateMultiplierBreakdown(tt.log)
			require.InDelta(t, tt.wantPricing, pricing, 1e-12)
			require.InDelta(t, tt.wantConversion, conversion, 1e-12)
			require.InDelta(t, tt.log.RateMultiplier, pricing*conversion, 1e-12)
		})
	}
}
