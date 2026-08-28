package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGeminiLongContextBillingOptionsPassesRuleToUsageInput(t *testing.T) {
	threshold, multiplier := geminiLongContextBillingOptions(&service.LegacyLongContextRule{
		Threshold:  200000,
		Multiplier: 2,
	})

	require.Equal(t, 200000, threshold)
	require.InDelta(t, 2.0, multiplier, 1e-12)
}

func TestGeminiLongContextBillingOptionsLeavesDisabledRuleUnset(t *testing.T) {
	threshold, multiplier := geminiLongContextBillingOptions(nil)

	require.Zero(t, threshold)
	require.Zero(t, multiplier)
}
