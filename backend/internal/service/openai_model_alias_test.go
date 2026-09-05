package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeKnownOpenAICodexModel_BareGPT56RoutesToSol(t *testing.T) {
	tests := map[string]string{
		"gpt-5.6":                    "gpt-5.6-sol",
		"openai/gpt-5.6":             "gpt-5.6-sol",
		"gpt5.6":                     "gpt-5.6-sol",
		"gpt-5.6-none":               "gpt-5.6-sol",
		"gpt-5.6-minimal":            "gpt-5.6-sol",
		"gpt-5.6-high":               "gpt-5.6-sol",
		"gpt-5.6-xhigh":              "gpt-5.6-sol",
		"gpt-5.6-max":                "gpt-5.6-sol",
		"gpt-5.6-2026-07-09":         "gpt-5.6-sol",
		"gpt-5.6-sol-max":            "gpt-5.6-sol",
		"openai/gpt-5.6-terra-xhigh": "gpt-5.6-terra",
		"gpt-5.6-luna-2026-07-09":    "gpt-5.6-luna",
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			require.Equal(t, expected, normalizeKnownOpenAICodexModel(input))
		})
	}
}

func TestNormalizeKnownOpenAICodexModel_GPT56RejectsUnknownSuffixes(t *testing.T) {
	for _, model := range []string{
		"gpt-5.6-ultra",
		"gpt-5.6-solstice",
		"gpt-5.6-terrain",
		"openai/gpt-5.6-ultra",
		"gpt-5.6-sol-ultra",
		"gpt-5.6-terra-ultra",
		"gpt-5.6-luna-ultra",
	} {
		t.Run(model, func(t *testing.T) {
			require.Empty(t, normalizeKnownOpenAICodexModel(model))
			require.Equal(t, model, normalizeCodexModel(model))
		})
	}
}

func TestIsOpenAIGPT56Model_BareAliasSuffixMatrix(t *testing.T) {
	for _, model := range []string{
		"gpt-5.6",
		"openai/gpt-5.6",
		"gpt5.6",
		"gpt-5.6-minimal",
		"gpt-5.6-xhigh",
		"gpt-5.6-max",
		"gpt-5.6-2026-07-09",
		"gpt-5.6-sol-max",
		"openai/gpt-5.6-terra-xhigh",
		"gpt-5.6-luna-2026-07-09",
	} {
		require.True(t, isOpenAIGPT56Model(model), model)
	}

	for _, model := range []string{
		"gpt-5.6-ultra",
		"gpt-5.6-solstice",
		"gpt-5.6-terrain",
		"gpt-5.6-sol-ultra",
		"gpt-5.6-terra-ultra",
		"gpt-5.6-luna-ultra",
		"gpt-5.4",
	} {
		require.False(t, isOpenAIGPT56Model(model), model)
	}
}

func TestUsageBillingModelCandidates_BareGPT56IncludesSol(t *testing.T) {
	require.Equal(t,
		[]string{"gpt-5.6", "gpt-5.6-sol"},
		usageBillingModelCandidates("gpt-5.6"),
	)
	require.Equal(t,
		[]string{"openai/gpt-5.6", "gpt-5.6", "gpt-5.6-sol"},
		usageBillingModelCandidates("openai/gpt-5.6"),
	)
	require.Equal(t,
		[]string{"gpt5.6-max", "gpt-5.6-max", "gpt-5.6-sol"},
		usageBillingModelCandidates("gpt5.6-max"),
	)
	require.Equal(t,
		[]string{"gpt-5.6-terra-max", "gpt-5.6-terra"},
		usageBillingModelCandidates("gpt-5.6-terra-max"),
	)
	require.Equal(t,
		[]string{"gpt-5.6-luna-xhigh", "gpt-5.6-luna"},
		usageBillingModelCandidates("gpt-5.6-luna-xhigh"),
	)

	for _, model := range []string{"gpt-5.6-ultra", "gpt-5.6-solstice", "gpt-5.6-terrain"} {
		candidates := usageBillingModelCandidates(model)
		require.Equal(t, []string{model}, candidates)
		require.NotContains(t, candidates, "gpt-5.6-sol")
		require.NotContains(t, candidates, "gpt-5.6-terra")
	}
}
