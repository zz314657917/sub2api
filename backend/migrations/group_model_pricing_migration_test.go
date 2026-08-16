package migrations

import (
	"strings"
	"testing"
)

func TestMigration221IsAdditiveAndBackfillsLongContextPricing(t *testing.T) {
	data, err := FS.ReadFile("221_group_model_pricing.sql")
	if err != nil {
		t.Fatalf("read migration 221: %v", err)
	}
	sql := string(data)
	for _, want := range []string{
		"ADD COLUMN IF NOT EXISTS long_context_pricing_enabled BOOLEAN NOT NULL DEFAULT TRUE",
		"ADD COLUMN IF NOT EXISTS model_pricing JSONB",
		"WHERE long_context_pricing_enabled IS DISTINCT FROM TRUE",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration 221 missing %q", want)
		}
	}
}
