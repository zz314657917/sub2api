package openai

import "testing"

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
