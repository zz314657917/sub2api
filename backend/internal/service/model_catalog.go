package service

import (
	"context"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
)

const (
	ModelCapabilityChat  = "chat"
	ModelCapabilityImage = "image"
	ModelCapabilityVideo = "video"
)

type ModelCatalog struct {
	Object      string             `json:"object"`
	Items       []ModelCatalogItem `json:"items"`
	ChatModels  []string           `json:"chat_models"`
	ImageModels []string           `json:"image_models"`
	VideoModels []string           `json:"video_models"`
}

type ModelCatalogItem struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	DisplayName  string   `json:"display_name,omitempty"`
	Capabilities []string `json:"capabilities"`
	Enabled      bool     `json:"enabled"`
}

func (s *GatewayService) GetAvailableModelCatalog(ctx context.Context, groupID *int64, platform string) ModelCatalog {
	modelIDs := s.GetAvailableModels(ctx, groupID, platform)
	if len(modelIDs) > 0 {
		catalog := BuildModelCatalog(platform, modelIDs)
		if len(catalog.Items) > 0 {
			return catalog
		}
	}
	return BuildModelCatalog(platform, defaultCatalogModelIDs(platform))
}

func BuildModelCatalog(platform string, modelIDs []string) ModelCatalog {
	items := make([]ModelCatalogItem, 0, len(modelIDs))
	seen := map[string]struct{}{}
	for _, modelID := range modelIDs {
		id := strings.TrimSpace(modelID)
		if id == "" || strings.Contains(id, "*") {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		capabilities := ModelCapabilitiesForPlatform(platform, id)
		items = append(items, ModelCatalogItem{
			ID:           id,
			Name:         id,
			DisplayName:  id,
			Capabilities: capabilities,
			Enabled:      !modelCapabilitiesOnlyVideo(capabilities),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if rank := modelCatalogCapabilityRank(items[i].Capabilities) - modelCatalogCapabilityRank(items[j].Capabilities); rank != 0 {
			return rank < 0
		}
		return items[i].ID < items[j].ID
	})
	catalog := ModelCatalog{Object: "model_catalog", Items: items}
	for _, item := range items {
		if modelHasCapability(item.Capabilities, ModelCapabilityChat) {
			catalog.ChatModels = append(catalog.ChatModels, item.ID)
		}
		if modelHasCapability(item.Capabilities, ModelCapabilityImage) {
			catalog.ImageModels = append(catalog.ImageModels, item.ID)
		}
		if modelHasCapability(item.Capabilities, ModelCapabilityVideo) {
			catalog.VideoModels = append(catalog.VideoModels, item.ID)
		}
	}
	return catalog
}

func ModelCapabilitiesForPlatform(platform, modelID string) []string {
	id := strings.TrimSpace(modelID)
	if id == "" {
		return nil
	}
	if isVideoCatalogModel(id) || IsVideoGenerationIntent("", id, nil) {
		return []string{ModelCapabilityVideo}
	}
	if isImageCatalogModel(id) && supportsGatewayImageAPI(platform) {
		return []string{ModelCapabilityImage}
	}
	return []string{ModelCapabilityChat}
}

func defaultCatalogModelIDs(platform string) []string {
	switch strings.TrimSpace(platform) {
	case PlatformOpenAI:
		out := make([]string, 0, len(openai.DefaultModels))
		for _, model := range openai.DefaultModels {
			out = append(out, model.ID)
		}
		return out
	case PlatformGemini:
		out := make([]string, 0, len(geminicli.DefaultModels))
		for _, model := range geminicli.DefaultModels {
			out = append(out, model.ID)
		}
		return out
	case PlatformAntigravity:
		models := antigravity.DefaultModels()
		out := make([]string, 0, len(models))
		for _, model := range models {
			out = append(out, model.ID)
		}
		return out
	default:
		out := make([]string, 0, len(claude.DefaultModels))
		for _, model := range claude.DefaultModels {
			out = append(out, model.ID)
		}
		return out
	}
}

func supportsGatewayImageAPI(platform string) bool {
	platform = strings.TrimSpace(platform)
	return platform == "" || platform == PlatformOpenAI
}

func isImageCatalogModel(modelID string) bool {
	lower := strings.ToLower(strings.TrimSpace(modelID))
	if lower == "" {
		return false
	}
	imageHints := []string{
		"image",
		"imagen",
		"flux",
		"stable-diffusion",
		"sdxl",
		"dall-e",
		"midjourney",
		"kolors",
		"ideogram",
		"recraft",
	}
	for _, hint := range imageHints {
		if strings.Contains(lower, hint) {
			return true
		}
	}
	return false
}

func isVideoCatalogModel(modelID string) bool {
	lower := strings.ToLower(strings.TrimSpace(modelID))
	if lower == "" {
		return false
	}
	videoHints := []string{
		"video",
		"text-to-video",
		"image-to-video",
		"t2v",
		"i2v",
		"sora",
		"veo",
		"kling",
		"hailuo",
		"runway",
		"luma",
		"seedance",
		"wan2",
		"wanx",
	}
	for _, hint := range videoHints {
		if strings.Contains(lower, hint) {
			return true
		}
	}
	return false
}

func modelCapabilitiesOnlyVideo(capabilities []string) bool {
	return len(capabilities) == 1 && capabilities[0] == ModelCapabilityVideo
}

func modelHasCapability(capabilities []string, capability string) bool {
	for _, item := range capabilities {
		if item == capability {
			return true
		}
	}
	return false
}

func modelCatalogCapabilityRank(capabilities []string) int {
	switch {
	case modelHasCapability(capabilities, ModelCapabilityChat):
		return 0
	case modelHasCapability(capabilities, ModelCapabilityImage):
		return 1
	case modelHasCapability(capabilities, ModelCapabilityVideo):
		return 2
	default:
		return 3
	}
}
