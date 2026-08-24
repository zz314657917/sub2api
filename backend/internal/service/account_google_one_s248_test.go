package service

import "testing"

func TestAccountGetModelMapping_GoogleOneUsesConservativeDefaults(t *testing.T) {
	account := &Account{
		Platform: PlatformGemini,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"oauth_type": "google_one",
		},
	}

	mapping := account.GetModelMapping()
	for _, model := range []string{"gemini-2.0-flash", "gemini-2.5-flash", "gemini-2.5-pro"} {
		if mapping[model] != model {
			t.Fatalf("expected Google One model %q to map to itself, got %q", model, mapping[model])
		}
	}
	for _, model := range []string{"gemini-2.5-flash-image", "gemini-3.1-flash-image", "gemini-3.5-flash"} {
		if _, ok := mapping[model]; ok {
			t.Fatalf("did not expect unsupported Google One model %q", model)
		}
	}
	if account.IsModelSupported("gemini-3.5-flash") {
		t.Fatal("Google One defaults must not treat unsupported models as eligible")
	}
}

func TestAccountGetModelMapping_GoogleOnePreservesExplicitMapping(t *testing.T) {
	account := &Account{
		Platform: PlatformGemini,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"oauth_type": "google_one",
			"model_mapping": map[string]any{
				"custom-model": "gemini-2.5-flash",
			},
		},
	}

	mapping := account.GetModelMapping()
	if mapping["custom-model"] != "gemini-2.5-flash" {
		t.Fatalf("expected explicit Google One mapping to be preserved, got %v", mapping)
	}
	if _, ok := mapping["gemini-2.5-flash"]; ok {
		t.Fatalf("did not expect defaults to overwrite an explicit mapping: %v", mapping)
	}
}
