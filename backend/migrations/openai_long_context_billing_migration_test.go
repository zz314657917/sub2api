package migrations

import (
	"strings"
	"testing"
)

func TestMigration220OpenAILongContextBillingIsAdditiveAndStrict(t *testing.T) {
	data, err := FS.ReadFile("220_openai_long_context_billing.sql")
	if err != nil {
		t.Fatalf("read migration 220: %v", err)
	}
	sql := string(data)
	for _, want := range []string{
		"ADD COLUMN IF NOT EXISTS long_context_billing_applied BOOLEAN NOT NULL DEFAULT FALSE",
		"CREATE OR REPLACE FUNCTION public.enforce_openai_long_context_billing_extra()",
		"openai_long_context_billing_enabled must be a boolean",
		"DROP TRIGGER IF EXISTS accounts_enforce_openai_long_context_billing_extra ON accounts",
		"CREATE TRIGGER accounts_enforce_openai_long_context_billing_extra",
	} {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration 220 missing %q", want)
		}
	}
	backfillIndex := strings.Index(sql, "UPDATE accounts")
	triggerIndex := strings.Index(sql, "CREATE TRIGGER accounts_enforce_openai_long_context_billing_extra")
	if backfillIndex == -1 || triggerIndex == -1 || backfillIndex > triggerIndex {
		t.Fatal("migration 220 must normalize legacy account values before enabling the strict trigger")
	}
	requireNoShadowColumns(t, sql)
}

func requireNoShadowColumns(t *testing.T, sql string) {
	t.Helper()
	for _, forbidden := range []string{"parent_account_id", "quota_dimension"} {
		if strings.Contains(sql, forbidden) {
			t.Fatalf("migration 220 must not reference unavailable %q", forbidden)
		}
	}
}
