package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestChannelMonitorUpstreamMigrationIsAdditiveAndPrivate(t *testing.T) {
	data, err := os.ReadFile("253_channel_monitor_upstream.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(data)
	for _, required := range []string{"DEFAULT 'chat_completions'", "DEFAULT 'probe'", "DEFAULT 0",
		"ON DELETE SET NULL", "quota JSONB", "'minimax'", "'antigravity'", "interval_seconds - jitter_seconds >= 15"} {
		if !strings.Contains(sql, required) {
			t.Errorf("migration missing %s", required)
		}
	}
	for _, denied := range []string{"opencode_go", "INSERT INTO settings", "DROP TABLE", "DELETE FROM", "UPDATE accounts"} {
		if strings.Contains(sql, denied) {
			t.Errorf("unexpected operation %s", denied)
		}
	}
}
