// Package openai provides helpers and types for OpenAI API integration.
package openai

import (
	_ "embed"
	"strings"
)

// Model represents an OpenAI model
type Model struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Created     int64  `json:"created"`
	OwnedBy     string `json:"owned_by"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
}

// DefaultModels OpenAI models list
var DefaultModels = []Model{
	{ID: "gpt-5.6-sol", Object: "model", Created: 1780876800, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.6 Sol"},
	{ID: "gpt-6-astra", Object: "model", Created: 1788480000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-6 Astra"},
	{ID: "gpt-6", Object: "model", Created: 1788480000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-6 (Astra)"},
	{ID: "gpt-6-sol", Object: "model", Created: 1788480000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-6 Sol"},
	{ID: "gpt-6-luna", Object: "model", Created: 1788480000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-6 Luna"},
	{ID: "gpt-5.6", Object: "model", Created: 1780876800, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.6 (Sol)"},
	{ID: "gpt-5.6-terra", Object: "model", Created: 1780876800, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.6 Terra"},
	{ID: "gpt-5.6-luna", Object: "model", Created: 1780876800, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.6 Luna"},
	{ID: "gpt-5.5", Object: "model", Created: 1776873600, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.5"},
	{ID: "gpt-5.4", Object: "model", Created: 1738368000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.4"},
	{ID: "gpt-5.4-mini", Object: "model", Created: 1738368000, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.4 Mini"},
	{ID: "gpt-5.3-codex", Object: "model", Created: 1735689600, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.3 Codex"},
	{ID: "gpt-5.3-codex-spark", Object: "model", Created: 1735689600, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.3 Codex Spark"},
	{ID: "codex-auto-review", Object: "model", Created: 1776902400, OwnedBy: "openai", Type: "model", DisplayName: "Codex Auto Review"},
	{ID: "gpt-5.2", Object: "model", Created: 1733875200, OwnedBy: "openai", Type: "model", DisplayName: "GPT-5.2"},
	{ID: "gpt-image-1", Object: "model", Created: 1733875200, OwnedBy: "openai", Type: "model", DisplayName: "GPT Image 1"},
	{ID: "gpt-image-1.5", Object: "model", Created: 1735689600, OwnedBy: "openai", Type: "model", DisplayName: "GPT Image 1.5"},
	{ID: "gpt-image-2", Object: "model", Created: 1738368000, OwnedBy: "openai", Type: "model", DisplayName: "GPT Image 2"},
	{ID: "gpt-image-2.5-flare", Object: "model", Created: 1788825600, OwnedBy: "openai", Type: "model", DisplayName: "GPT Image 2.5 Flare"},
	{ID: "gpt-image-2.5-sunburst", Object: "model", Created: 1788825600, OwnedBy: "openai", Type: "model", DisplayName: "GPT Image 2.5 Sunburst"},
	{ID: "gpt-image-2.5-flare-2026-09-08", Object: "model", Created: 1788825600, OwnedBy: "openai", Type: "model", DisplayName: "GPT Image 2.5 Flare (2026-09-08)"},
	{ID: "gpt-image-2.5-sunburst-2026-09-08", Object: "model", Created: 1788825600, OwnedBy: "openai", Type: "model", DisplayName: "GPT Image 2.5 Sunburst (2026-09-08)"},
}

// DefaultModelIDs returns the default model ID list
func DefaultModelIDs() []string {
	ids := make([]string, len(DefaultModels))
	for i, m := range DefaultModels {
		ids[i] = m.ID
	}
	return ids
}

// DefaultTestModel default model for testing OpenAI accounts
const DefaultTestModel = "gpt-5.4"

// CodexUsageProbeModel is the model used for OAuth Codex usage probes.
const CodexUsageProbeModel = "codex-auto-review"

// DefaultInstructions default instructions for non-Codex CLI requests
// Content loaded from instructions.txt at compile time
//
//go:embed instructions.txt
var DefaultInstructions string

// instructionsGPT51 / instructionsGPT52 / instructionsGPT55 are the Codex CLI
// base prompts for non-codex GPT-5.x models. GPT-5.5 is also the latest
// fallback for GPT-5.3/GPT-5.4 and unknown models.
//
//go:embed instructions_gpt5_1.txt
var instructionsGPT51 string

//go:embed instructions_gpt5_2.txt
var instructionsGPT52 string

//go:embed instructions_gpt5_5.txt
var instructionsGPT55 string

// Source: upstream b939fa9d4, openai/codex models.json at 121f91fd5d9d.
//
//go:embed instructions_gpt6_astra.txt
var instructionsGPT6Astra string

func latestCodexInstructions() string {
	if v := strings.TrimSpace(instructionsGPT55); v != "" {
		return instructionsGPT55
	}
	return DefaultInstructions
}

// CodexBaseInstructionsForModel returns the closest Codex base prompt for a
// model. Codex-family models use the GPT-5 Codex prompt, GPT-5.1/5.2/5.5 use
// their matching prompts, and unversioned or unknown models fall back to the
// latest known prompt.
func CodexBaseInstructionsForModel(model string) string {
	m := strings.ToLower(strings.TrimSpace(model))
	if slash := strings.LastIndexByte(m, '/'); slash >= 0 {
		m = strings.TrimSpace(m[slash+1:])
	}
	m = strings.ReplaceAll(m, "_", "-")
	m = strings.Join(strings.Fields(m), "-")
	for strings.Contains(m, "--") {
		m = strings.ReplaceAll(m, "--", "-")
	}
	switch {
	case m == "gpt-6" || m == "gpt-6-astra" || strings.HasPrefix(m, "gpt-6-astra-"):
		if strings.TrimSpace(instructionsGPT6Astra) != "" {
			return instructionsGPT6Astra
		}
	case strings.Contains(m, "codex"):
		return DefaultInstructions
	case strings.HasPrefix(m, "gpt-5.5"):
		return latestCodexInstructions()
	case strings.HasPrefix(m, "gpt-5.2"):
		if v := strings.TrimSpace(instructionsGPT52); v != "" {
			return instructionsGPT52
		}
	case strings.HasPrefix(m, "gpt-5.1"):
		if v := strings.TrimSpace(instructionsGPT51); v != "" {
			return instructionsGPT51
		}
	}
	return latestCodexInstructions()
}
