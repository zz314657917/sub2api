package migrations

import (
	"strings"
	"testing"
)

func TestGrokImage20TutorialMigration(t *testing.T) {
	content, err := FS.ReadFile("232_replace_grok_15_with_20_tutorial_pages.sql")
	if err != nil {
		t.Fatalf("read Grok 2.0 tutorial migration: %v", err)
	}
	sql := string(content)
	for _, required := range []string{
		"DELETE FROM tutorial_pages",
		"grok-imagine-2.0-ext",
		"grok-imagine-image-2.0",
		"aspect_ratio",
		"image_urls",
		"invalid_idempotency_key",
		"GET /v1/tasks/{task_id}",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing %q", required)
		}
	}
	for _, removed := range []string{"grok-imagine-1.5-apimart", "grok-imagine-1.5-edit-apimart"} {
		if strings.Contains(sql, removed) {
			t.Fatalf("migration still contains removed 1.5 model %q", removed)
		}
	}
}
