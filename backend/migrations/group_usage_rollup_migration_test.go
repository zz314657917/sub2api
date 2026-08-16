package migrations

import (
	"strings"
	"testing"
)

func TestMigration222CreatesRollupStateAndInvalidationTriggers(t *testing.T) {
	data, err := FS.ReadFile("222_group_usage_daily_rollups.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS usage_group_daily_rollups",
		"CREATE TABLE IF NOT EXISTS usage_group_rollup_state",
		"FOR UPDATE",
		"FOR KEY SHARE",
		"usage_logs_group_rollup_invalidate_insert",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration 222 is missing %q", required)
		}
	}
}

func TestMigration223UsesDatabaseSessionTimezone(t *testing.T) {
	data, err := FS.ReadFile("223_group_usage_rollup_timezone.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	for _, required := range []string{
		"timezone_name",
		"current_setting('TimeZone')",
		"invalidate_group_usage_rollup_state_after_insert",
	} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration 223 is missing %q", required)
		}
	}
}
