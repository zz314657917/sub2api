package openai

import "testing"

func TestPricingModelSpellingGPT61Sol(t *testing.T) {
	for _, model := range []string{"gpt-6.1-sol", "gpt-6.1-sol-none", "gpt-6.1-sol-minimal",
		"gpt-6.1-sol-low", "gpt-6.1-sol-medium", "gpt-6.1-sol-high", "gpt-6.1-sol-xhigh",
		"gpt-6.1-sol-max", "openai/gpt-6.1-sol-openai-compact", "OpenAI/GPT_6.1_SOL_HIGH",
		" gpt-6.1 sol medium "} {
		if !IsGPT61SolModelSpelling(model) {
			t.Errorf("valid pricing spelling rejected: %q", model)
		}
	}
	for _, model := range []string{"gpt-6-sol", "gpt-6-luna", "gpt-6.1", "gpt-6.1-sol-invalid", "gpt-6.1-sol-high-extra", "claude-sol"} {
		if IsGPT61SolModelSpelling(model) {
			t.Errorf("unknown spelling accepted: %q", model)
		}
	}
}
