package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildModelCatalog_GroupsChatImageAndReservedVideoModels(t *testing.T) {
	catalog := BuildModelCatalog(PlatformOpenAI, []string{
		"gpt-5.4",
		"gpt-image-2",
		"black-forest-labs/FLUX.2-klein-9B",
		"sora-2",
		"gpt-5.4",
	})

	require.Equal(t, "model_catalog", catalog.Object)
	require.Equal(t, []string{"gpt-5.4"}, catalog.ChatModels)
	require.Equal(t, []string{"black-forest-labs/FLUX.2-klein-9B", "gpt-image-2"}, catalog.ImageModels)
	require.Equal(t, []string{"sora-2"}, catalog.VideoModels)

	byID := modelCatalogItemsByID(catalog.Items)
	require.Equal(t, []string{ModelCapabilityChat}, byID["gpt-5.4"].Capabilities)
	require.Equal(t, []string{ModelCapabilityImage}, byID["gpt-image-2"].Capabilities)
	require.True(t, byID["gpt-image-2"].Enabled)
	require.Equal(t, []string{ModelCapabilityVideo}, byID["sora-2"].Capabilities)
	require.False(t, byID["sora-2"].Enabled)
}

func TestBuildModelCatalog_GeminiImageNamesStayChatUntilGatewaySupportsImages(t *testing.T) {
	catalog := BuildModelCatalog(PlatformGemini, []string{"gemini-2.5-flash-image"})

	require.Equal(t, []string{"gemini-2.5-flash-image"}, catalog.ChatModels)
	require.Empty(t, catalog.ImageModels)
	require.Empty(t, catalog.VideoModels)
	require.Equal(t, []string{ModelCapabilityChat}, catalog.Items[0].Capabilities)
	require.True(t, catalog.Items[0].Enabled)
}

func modelCatalogItemsByID(items []ModelCatalogItem) map[string]ModelCatalogItem {
	out := make(map[string]ModelCatalogItem, len(items))
	for _, item := range items {
		out[item.ID] = item
	}
	return out
}
