package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSanitizeAnthropicBodyForBetaTokens_StripsFallbackFieldsWithoutBetas(t *testing.T) {
	body := []byte(`{"model":"claude-sonnet-4-6","fallbacks":"default","fallback_credit_token":"tok","messages":[]}`)

	out, changed := sanitizeAnthropicBodyForBetaTokens(body, claude.BetaOAuth)

	require.True(t, changed)
	require.False(t, gjson.GetBytes(out, "fallbacks").Exists())
	require.False(t, gjson.GetBytes(out, "fallback_credit_token").Exists())
	require.True(t, gjson.GetBytes(out, "messages").Exists())
}

func TestSanitizeAnthropicBodyForBetaTokens_KeepsFallbackFieldsWithMatchingBetas(t *testing.T) {
	body := []byte(`{"fallbacks":["claude-opus-4-6"],"fallback_credit_token":"tok"}`)

	out, changed := sanitizeAnthropicBodyForBetaTokens(body, claude.BetaServerSideFallback)
	require.False(t, changed)
	require.True(t, gjson.GetBytes(out, "fallbacks").Exists())
	require.True(t, gjson.GetBytes(out, "fallback_credit_token").Exists())

	for _, betaHeader := range []string{claude.BetaFallbackCredit, claude.BetaFallbackCreditLegacy} {
		out, changed := sanitizeAnthropicBodyForBetaTokens(body, betaHeader)
		require.Truef(t, changed, "fallbacks should be removed with credit-only beta %q", betaHeader)
		require.Falsef(t, gjson.GetBytes(out, "fallbacks").Exists(), "credit-only beta %q", betaHeader)
		require.Truef(t, gjson.GetBytes(out, "fallback_credit_token").Exists(), "credit-only beta %q", betaHeader)
	}
}

func TestSanitizeAnthropicBodyForBetaTokens_PreservesContextManagementIndependently(t *testing.T) {
	body := []byte(`{"context_management":{"edits":[]},"fallbacks":"default"}`)

	out, changed := sanitizeAnthropicBodyForBetaTokens(body, claude.BetaContextManagement)

	require.True(t, changed)
	require.True(t, gjson.GetBytes(out, "context_management").Exists())
	require.False(t, gjson.GetBytes(out, "fallbacks").Exists())
}

func TestSanitizeBedrockFieldsForBetaTokens_StripsAnthropicFallbackFields(t *testing.T) {
	body := []byte(`{"context_management":{"edits":[]},"fallbacks":"default","fallback_credit_token":"tok"}`)

	out := sanitizeBedrockFieldsForBetaTokens(body, []string{bedrockContextManagementBetaToken, claude.BetaServerSideFallback})

	require.True(t, gjson.GetBytes(out, "context_management").Exists())
	require.False(t, gjson.GetBytes(out, "fallbacks").Exists())
	require.False(t, gjson.GetBytes(out, "fallback_credit_token").Exists())
}
