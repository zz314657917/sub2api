//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAILongContextBillingMigration(t *testing.T) {
	ctx := context.Background()
	name := fmt.Sprintf("s220-openai-%d", time.Now().UnixNano())
	var accountID int64
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		INSERT INTO accounts (name, platform, type, credentials, extra, concurrency, priority, status, schedulable)
		VALUES ($1, 'openai', 'apikey', '{}'::jsonb, '{}'::jsonb, 1, 1, 'active', TRUE)
		RETURNING id`, name).Scan(&accountID))
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), "DELETE FROM accounts WHERE id = $1", accountID)
	})

	var enabled bool
	require.NoError(t, integrationDB.QueryRowContext(ctx,
		"SELECT (extra->>'openai_long_context_billing_enabled')::boolean FROM accounts WHERE id = $1", accountID,
	).Scan(&enabled))
	require.False(t, enabled)

	_, err := integrationDB.ExecContext(ctx,
		"UPDATE accounts SET extra = jsonb_build_object('openai_long_context_billing_enabled', 'true') WHERE id = $1", accountID,
	)
	require.Error(t, err)

	var nullable string
	require.NoError(t, integrationDB.QueryRowContext(ctx, `
		SELECT is_nullable
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'usage_logs' AND column_name = 'long_context_billing_applied'`,
	).Scan(&nullable))
	require.Equal(t, "NO", nullable)
}
