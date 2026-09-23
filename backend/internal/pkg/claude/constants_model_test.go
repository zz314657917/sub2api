package claude

import "testing"

func TestDefaultModelsContainsFable51(t *testing.T) {
	for _, model := range DefaultModels {
		if model.ID == "claude-fable-5-1" {
			if model.DisplayName != "Claude Fable 5.1" {
				t.Fatalf("unexpected Fable 5.1 display name: %q", model.DisplayName)
			}
			return
		}
	}
	t.Fatal("claude-fable-5-1 is missing from DefaultModels")
}

func TestDefaultModelsContainsOpus55(t *testing.T) {
	for _, model := range DefaultModels {
		if model.ID == "claude-opus-5-5" {
			if model.DisplayName != "Claude Opus 5.5" || model.CreatedAt != "2026-09-22T00:00:00Z" {
				t.Fatalf("unexpected Opus 5.5 descriptor: %+v", model)
			}
			return
		}
	}
	t.Fatal("claude-opus-5-5 missing from DefaultModels")
}

func TestIsOpus55StrictSpelling(t *testing.T) {
	for _, id := range []string{"claude-opus-5-5", "models/claude-opus-5-5-thinking",
		"anthropic.claude-opus-5-5-20260922", "anthropic/claude-opus-5-5-20260922-thinking"} {
		if !IsOpus55(id) {
			t.Errorf("%q should match Opus 5.5", id)
		}
	}
	for _, id := range []string{"claude-opus-5", "claude-opus-5-5-preview", "claude-opus-5-50", "claude-opus-5-5-20260922-extra"} {
		if IsOpus55(id) {
			t.Errorf("%q should not match Opus 5.5", id)
		}
	}
}
