package migrations

import (
	"strings"
	"testing"
)

func TestSeedream5TutorialMigration(t *testing.T) {
	content, err := FS.ReadFile("233_add_seedream_5_tutorial_pages.sql")
	if err != nil {
		t.Fatalf("read Seedream 5 tutorial migration: %v", err)
	}
	sql := string(content)
	for _, required := range []string{"seedream-5-0-pro", "seedream-5-0-lite", "image_urls", "sequential_image_generation", "429 rate_limit_error"} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing %q", required)
		}
	}
	for _, forbidden := range []string{"ai.3zapi.top", "ai.3zapi.com", "api.apimart.ai"} {
		if strings.Contains(strings.ToLower(sql), forbidden) {
			t.Fatalf("migration contains forbidden host %q", forbidden)
		}
	}
}
