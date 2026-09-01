package migrations

import (
	"strings"
	"testing"
)

func TestUsageBillingSettlementConsistency(t *testing.T) {
	data, err := FS.ReadFile("238_usage_billing_settlement_consistency.sql")
	if err != nil {
		t.Fatalf("read migration 238: %v", err)
	}
	sql := string(data)

	for _, want := range []string{
		"ADD COLUMN IF NOT EXISTS billing_status VARCHAR(16) NOT NULL DEFAULT 'applied'",
		"ADD COLUMN IF NOT EXISTS billing_error TEXT",
		"usage_logs_billing_status_check",
		"CHECK (billing_status IN ('pending', 'failed', 'applied'))",
		"CREATE TABLE IF NOT EXISTS usage_billing_settlement_outbox",
		"usage_log_id BIGINT NOT NULL REFERENCES usage_logs(id) ON DELETE CASCADE",
		"request_id VARCHAR(255) NOT NULL",
		"api_key_id BIGINT NOT NULL",
		"payload JSONB NOT NULL",
		"CHECK (status IN ('pending', 'processing', 'failed', 'applied'))",
		"UNIQUE (usage_log_id)",
		"CREATE INDEX IF NOT EXISTS idx_usage_billing_settlement_outbox_claim",
		"ON usage_billing_settlement_outbox (status, available_at, lease_until, id)",
		"CREATE UNIQUE INDEX IF NOT EXISTS usage_billing_settlement_outbox_request_api_key_unique",
		"ON usage_billing_settlement_outbox (request_id, api_key_id)",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration 238 missing outbox constraint or index %q", want)
		}
	}

	if strings.Contains(sql, "ledger_key") || strings.Contains(sql, "DROP INDEX IF EXISTS billing_usage_entries_usage_log_id_unique") {
		t.Fatal("migration 238 must preserve the existing usage_log_id ledger uniqueness")
	}
}
