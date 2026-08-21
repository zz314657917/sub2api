package migrations

import (
	"strings"
	"testing"
)

func TestSeedream5MarkdownFixMigration(t *testing.T) {
	content, err := FS.ReadFile("234_fix_seedream_5_tutorial_markdown.sql")
	if err != nil {
		t.Fatalf("read Seedream 5 markdown fix migration: %v", err)
	}
	sql := string(content)
	for _, required := range []string{
		"seedream-5-0-pro",
		"seedream-5-0-lite",
		"replace(content_md",
		"E'\\\\`'",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing %q", required)
		}
	}
}
