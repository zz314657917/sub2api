package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeAccountCapabilities(t *testing.T) {
	require.Equal(t, []string{"chat", "image", "video", "embedding"}, NormalizeAccountCapabilities([]any{
		" llm ",
		"images",
		"videos",
		"embeddings",
		"chat",
		"unknown",
	}))

	require.Equal(t, []string{"chat", "video"}, NormalizeAccountCapabilities(map[string]any{
		"text":  true,
		"video": true,
		"image": false,
	}))
}

func TestAccountSupportsCapability_DefaultsToAllWhenUnsetOrEmpty(t *testing.T) {
	account := &Account{}

	require.True(t, account.SupportsCapability(AccountCapabilityChat))
	require.True(t, account.SupportsCapability(AccountCapabilityImage))
	require.True(t, account.SupportsCapability(AccountCapabilityVideo))
	require.True(t, account.SupportsCapability(AccountCapabilityEmbedding))

	account.Extra = map[string]any{accountSupportedCapabilitiesExtraKey: []any{}}
	require.True(t, account.SupportsCapability(AccountCapabilityImage))
}

func TestAccountSupportsCapability_ExplicitChatOnly(t *testing.T) {
	account := &Account{
		Extra: map[string]any{
			accountSupportedCapabilitiesExtraKey: []any{"chat"},
		},
	}

	require.True(t, account.HasExplicitSupportedCapabilities())
	require.True(t, account.SupportsCapability(AccountCapabilityChat))
	require.False(t, account.SupportsCapability(AccountCapabilityImage))
	require.False(t, account.SupportsCapability(AccountCapabilityVideo))
	require.False(t, account.SupportsCapability(AccountCapabilityEmbedding))
}

func TestApplyAccountSupportedCapabilities_NormalizesButKeepsEmptyOverride(t *testing.T) {
	extra := ApplyAccountSupportedCapabilities(map[string]any{
		accountSupportedCapabilitiesExtraKey: []any{"images", "text", "image"},
		"quota_limit":                        12,
	})

	require.Equal(t, []string{"chat", "image"}, extra[accountSupportedCapabilitiesExtraKey])
	require.Equal(t, 12, extra["quota_limit"])

	extra = ApplyAccountSupportedCapabilities(map[string]any{
		accountSupportedCapabilitiesExtraKey: []any{"unknown"},
	})
	require.Equal(t, []string{}, extra[accountSupportedCapabilitiesExtraKey])
}

func TestAccountSupportsOpenAICapabilities_UsesGenericCapabilities(t *testing.T) {
	account := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeAPIKey,
		Extra: map[string]any{
			accountSupportedCapabilitiesExtraKey: []any{"chat", "embedding"},
		},
	}

	require.True(t, accountSupportsOpenAICapabilities(account, OpenAIEndpointCapabilityChatCompletions, "", ""))
	require.True(t, accountSupportsOpenAICapabilities(account, OpenAIEndpointCapabilityEmbeddings, "", ""))
	require.False(t, accountSupportsOpenAICapabilities(account, "", OpenAIImagesCapabilityNative, ""))
	require.False(t, accountSupportsOpenAICapabilities(account, "", "", AccountCapabilityVideo))
}
