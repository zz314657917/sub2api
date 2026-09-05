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
