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
