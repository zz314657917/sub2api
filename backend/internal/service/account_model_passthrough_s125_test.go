package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsModelSupported_OpenAIPassthroughIgnoresLeftoverMapping(t *testing.T) {
	for _, accountType := range []string{AccountTypeOAuth, AccountTypeAPIKey} {
		t.Run(accountType, func(t *testing.T) {
			account := &Account{
				Platform: PlatformOpenAI,
				Type:     accountType,
				Extra:    map[string]any{"openai_passthrough": true},
				Credentials: map[string]any{
					"model_mapping": map[string]any{"gpt-5.4": "gpt-5.4"},
				},
			}

			require.True(t, account.IsModelSupported("gpt-5.6-sol"))
			require.True(t, account.IsModelSupported("deepseek-v4"))
		})
	}

	nonPassthrough := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"model_mapping": map[string]any{"gpt-5.4": "gpt-5.4"},
		},
	}
	require.False(t, nonPassthrough.IsModelSupported("gpt-5.6-sol"))
}
