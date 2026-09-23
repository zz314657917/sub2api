package openai

import "testing"

func TestDefaultModelsIncludeBareGPT56Alias(t *testing.T) {
	count := 0
	for _, model := range DefaultModels {
		if model.ID != "gpt-5.6" {
			continue
		}
		count++
		if model.DisplayName != "GPT-5.6 (Sol)" {
			t.Fatalf("gpt-5.6 display name = %q, want %q", model.DisplayName, "GPT-5.6 (Sol)")
		}
	}
	if count != 1 {
		t.Fatalf("DefaultModels contains %d bare gpt-5.6 entries, want exactly 1", count)
	}
}

func TestDefaultModelsIncludeGPT6Astra(t *testing.T) {
	if !containsModelID(DefaultModelIDs(), "gpt-6-astra") {
		t.Fatal("DefaultModels missing gpt-6-astra")
	}
	if !containsModelID(DefaultModelIDs(), "gpt-6") {
		t.Fatal("DefaultModels missing gpt-6")
	}
	if !containsModelID(DefaultModelIDs(), "gpt-6-sol") || !containsModelID(DefaultModelIDs(), "gpt-6-luna") {
		t.Fatal("DefaultModels missing GPT-6 Sol/Luna")
	}
}

func containsModelID(models []string, want string) bool {
	for _, model := range models {
		if model == want {
			return true
		}
	}
	return false
}

func TestDefaultModelsPreferConcreteGPT56SolForAccountTests(t *testing.T) {
	if len(DefaultModels) == 0 {
		t.Fatal("DefaultModels is empty")
	}
	if DefaultModels[0].ID != "gpt-5.6-sol" {
		t.Fatalf("DefaultModels[0].ID = %q, want %q", DefaultModels[0].ID, "gpt-5.6-sol")
	}
}

func TestDefaultModelsIncludeGPTImage25(t *testing.T) {
	ids := DefaultModelIDs()
	for _, want := range []string{"gpt-image-2.5-flare", "gpt-image-2.5-sunburst"} {
		found := false
		for _, id := range ids {
			found = found || id == want
		}
		if !found {
			t.Fatalf("DefaultModels does not contain %q", want)
		}
	}
}

func TestDefaultModelsContainsCodexAutoReview(t *testing.T) {
	for _, model := range DefaultModels {
		if model.ID == "codex-auto-review" {
			if model.DisplayName != "Codex Auto Review" {
				t.Fatalf("codex-auto-review display name = %q, want %q", model.DisplayName, "Codex Auto Review")
			}
			return
		}
	}

	t.Fatal("DefaultModels missing codex-auto-review")
}

func TestGPT6SolLunaModelIdentity(t *testing.T) {
	for _, model := range []string{"gpt-6-sol", "gpt-6-luna", "openai/gpt-6-sol-max",
		"gpt-6-luna-openai-compact", "GPT_6_SOL_XHIGH"} {
		if !IsGPT6SolOrLunaModelSpelling(model) {
			t.Errorf("%q should match GPT-6 Sol/Luna", model)
		}
	}
	for _, model := range []string{"gpt-6-astra", "gpt-6-solitude", "gpt-6-luna-preview"} {
		if IsGPT6SolOrLunaModelSpelling(model) {
			t.Errorf("%q should not match GPT-6 Sol/Luna", model)
		}
	}
}
