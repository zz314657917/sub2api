package service

import "testing"

import "github.com/stretchr/testify/require"

func TestNormalizeOpenAIMessagesDispatchModelConfig(t *testing.T) {
	t.Parallel()

	cfg := normalizeOpenAIMessagesDispatchModelConfig(OpenAIMessagesDispatchModelConfig{
		OpusMappedModel:   " gpt-5.4-high ",
		SonnetMappedModel: "gpt-5.3-codex",
		HaikuMappedModel:  " gpt-5.4-mini-medium ",
		ExactModelMappings: map[string]string{
			" claude-sonnet-4-5-20250929 ": " gpt-5.2-high ",
			"":                             "gpt-5.4",
			"claude-opus-4-6":              " ",
		},
	})

	require.Equal(t, "gpt-5.4", cfg.OpusMappedModel)
	require.Equal(t, "gpt-5.3-codex", cfg.SonnetMappedModel)
	require.Equal(t, "gpt-5.4-mini", cfg.HaikuMappedModel)
	require.Equal(t, map[string]string{
		"claude-sonnet-4-5-20250929": "gpt-5.2",
	}, cfg.ExactModelMappings)
}

func TestResolveMessagesDispatchModel_CNProvidersBypassOpenAIDefaults(t *testing.T) {
	for _, platform := range []string{PlatformKimi, PlatformZhipu, PlatformDeepseek} {
		t.Run(platform, func(t *testing.T) {
			group := &Group{
				Platform: platform,
				MessagesDispatchModelConfig: OpenAIMessagesDispatchModelConfig{
					OpusMappedModel:    "gpt-5.4",
					ExactModelMappings: map[string]string{"claude-sonnet-4-5": "gpt-5.3-codex"},
				},
			}
			require.Empty(t, group.ResolveMessagesDispatchModel("claude-sonnet-4-5"))
		})
	}

	openAI := &Group{Platform: PlatformOpenAI}
	require.Equal(t, defaultOpenAIMessagesDispatchOpusMappedModel, openAI.ResolveMessagesDispatchModel("claude-opus-4-5"))
}
