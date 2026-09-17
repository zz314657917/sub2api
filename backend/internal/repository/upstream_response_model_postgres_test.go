package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
)

const upstreamResponseModelPostgresDSNEnv = "UPSTREAM_RESPONSE_MODEL_POSTGRES_DSN"

func TestUpstreamResponseModelPostgresPersistenceAndFilters(t *testing.T) {
	ctx := context.Background()
	dsn := os.Getenv(upstreamResponseModelPostgresDSNEnv)
	if dsn == "" {
		t.Skipf("%s is not set", upstreamResponseModelPostgresDSNEnv)
	}
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.PingContext(ctx))
	prepareUpstreamResponseModelMigrationReplay(t, ctx, db)

	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	t.Cleanup(func() { _ = client.Close() })
	user, err := client.User.Create().SetEmail("response-model-audit-" + time.Now().UTC().Format("20060102150405.000000000") + "@example.test").SetPasswordHash("audit-test-password-hash").Save(ctx)
	require.NoError(t, err)
	key, err := client.APIKey.Create().SetUserID(user.ID).SetKey("sk-response-model-audit-" + time.Now().UTC().Format("150405.000000000")).SetName("audit").Save(ctx)
	require.NoError(t, err)
	account, err := client.Account.Create().
		SetName("response-model-audit-" + time.Now().UTC().Format("150405.000000000")).
		SetPlatform(service.PlatformOpenAI).
		SetType(service.AccountTypeOAuth).
		SetCredentials(map[string]any{}).
		SetExtra(map[string]any{}).
		SetConcurrency(1).
		SetPriority(1).
		SetStatus(service.StatusActive).
		SetSchedulable(true).
		Save(ctx)
	require.NoError(t, err)
	legacyRequestID := "audit-legacy-" + time.Now().UTC().Format("150405.000000000")
	_, err = db.ExecContext(ctx, `INSERT INTO usage_logs (user_id, api_key_id, account_id, request_id, model, created_at) VALUES ($1, $2, $3, $4, 'gpt-5', NOW())`, user.ID, key.ID, account.ID, legacyRequestID)
	require.NoError(t, err)

	// The production runner is intentionally called twice: the second pass
	// validates migration recording/checksum idempotency on the task database.
	require.NoError(t, ApplyMigrations(ctx, db))
	require.NoError(t, ApplyMigrations(ctx, db))

	var columns int
	require.NoError(t, db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'usage_logs'
		  AND column_name IN ('upstream_response_model', 'upstream_model_mismatch')
	`).Scan(&columns))
	require.Equal(t, 2, columns)
	var legacyResponseModel sql.NullString
	var legacyMismatch sql.NullBool
	require.NoError(t, db.QueryRowContext(ctx, "SELECT upstream_response_model, upstream_model_mismatch FROM usage_logs WHERE request_id = $1", legacyRequestID).Scan(&legacyResponseModel, &legacyMismatch))
	require.False(t, legacyResponseModel.Valid)
	require.False(t, legacyMismatch.Valid)

	repo := newUsageLogRepositoryWithSQL(client, db)
	createdAt := time.Now().UTC().Add(-time.Minute)
	match := false
	mismatch := true
	equalLog := &service.UsageLog{UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID, RequestID: "audit-equal-" + time.Now().UTC().Format("150405.000000000"), Model: "gpt-5", RequestedModel: "gpt-5", UpstreamResponseModel: stringPtrForResponseAudit("gpt-5"), UpstreamModelMismatch: &match, InputTokens: 3, OutputTokens: 2, TotalCost: 1, ActualCost: 1, CreatedAt: createdAt}
	differentLog := &service.UsageLog{UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID, RequestID: "audit-different-" + time.Now().UTC().Format("150405.000000000"), Model: "gpt-5", RequestedModel: "gpt-5", UpstreamModel: stringPtrForResponseAudit("gpt-5-upstream"), UpstreamResponseModel: stringPtrForResponseAudit("gpt-5-upstream"), UpstreamModelMismatch: &mismatch, InputTokens: 5, OutputTokens: 3, TotalCost: 2, ActualCost: 2, CreatedAt: createdAt.Add(time.Second)}
	unknownLog := &service.UsageLog{UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID, RequestID: "audit-unknown-" + time.Now().UTC().Format("150405.000000000"), Model: "gpt-5", RequestedModel: "gpt-5", InputTokens: 7, OutputTokens: 4, TotalCost: 3, ActualCost: 3, CreatedAt: createdAt.Add(2 * time.Second)}

	inserted, err := repo.Create(ctx, equalLog)
	require.NoError(t, err)
	require.True(t, inserted)
	inserted, err = repo.Create(ctx, differentLog)
	require.NoError(t, err)
	require.True(t, inserted)
	inserted, err = repo.Create(ctx, unknownLog)
	require.NoError(t, err)
	require.True(t, inserted)

	batchMismatch := true
	batchLog := &service.UsageLog{UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID, RequestID: "audit-batch-" + time.Now().UTC().Format("150405.000000000"), Model: "gpt-5", RequestedModel: "gpt-5", UpstreamResponseModel: stringPtrForResponseAudit("other-model"), UpstreamModelMismatch: &batchMismatch, CreatedAt: createdAt.Add(3 * time.Second)}
	batchResult := make(chan usageLogCreateResult, 1)
	repo.flushCreateBatch(db, []usageLogCreateRequest{{log: batchLog, prepared: prepareUsageLogInsert(batchLog), resultCh: batchResult}})
	batchInsert := <-batchResult
	require.NoError(t, batchInsert.err)
	require.True(t, batchInsert.inserted)

	bestEffortMismatch := true
	bestEffortLog := &service.UsageLog{UserID: user.ID, APIKeyID: key.ID, AccountID: account.ID, RequestID: "audit-best-effort-" + time.Now().UTC().Format("150405.000000000"), Model: "gpt-5", RequestedModel: "gpt-5", UpstreamResponseModel: stringPtrForResponseAudit("best-effort-other"), UpstreamModelMismatch: &bestEffortMismatch, CreatedAt: createdAt.Add(4 * time.Second)}
	require.NoError(t, repo.CreateBestEffort(ctx, bestEffortLog))

	readBack, err := repo.GetByID(ctx, differentLog.ID)
	require.NoError(t, err)
	require.NotNil(t, readBack.UpstreamResponseModel)
	require.Equal(t, "gpt-5-upstream", *readBack.UpstreamResponseModel)
	require.NotNil(t, readBack.UpstreamModelMismatch)
	require.True(t, *readBack.UpstreamModelMismatch)

	for _, tc := range []struct {
		name   string
		filter *bool
		want   int
	}{
		{name: "true", filter: &mismatch, want: 3},
		{name: "false", filter: &match, want: 1},
		{name: "unset includes null", filter: nil, want: 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			filters := usagestats.UsageLogFilters{UserID: user.ID, UpstreamModelMismatch: tc.filter, ExactTotal: true}
			rows, page, err := repo.ListWithFilters(ctx, pagination.PaginationParams{Page: 1, PageSize: 20}, filters)
			require.NoError(t, err)
			require.Len(t, rows, tc.want)
			require.Equal(t, int64(tc.want), page.Total)
			stats, err := repo.GetStatsWithFilters(ctx, filters)
			require.NoError(t, err)
			require.Equal(t, int64(tc.want), stats.TotalRequests)
		})
	}
	for pageNumber := 1; pageNumber <= 6; pageNumber++ {
		rows, page, err := repo.ListWithFilters(ctx, pagination.PaginationParams{Page: pageNumber, PageSize: 1}, usagestats.UsageLogFilters{UserID: user.ID, ExactTotal: true})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.Equal(t, int64(6), page.Total)
	}
	trend, err := repo.GetUsageTrendWithUsageFilters(ctx, createdAt.Add(-time.Second), createdAt.Add(time.Minute), "day", usagestats.UsageLogFilters{UserID: user.ID, UpstreamModelMismatch: &mismatch})
	require.NoError(t, err)
	require.Len(t, trend, 1)
	require.Equal(t, int64(3), trend[0].Requests)
	models, err := repo.GetModelStatsWithUsageFiltersBySource(ctx, createdAt.Add(-time.Second), createdAt.Add(time.Minute), usagestats.UsageLogFilters{UserID: user.ID, UpstreamModelMismatch: &mismatch}, usagestats.ModelSourceRequested)
	require.NoError(t, err)
	require.Len(t, models, 1)
	require.Equal(t, int64(3), models[0].Requests)
	groups, err := repo.GetGroupStatsWithUsageFilters(ctx, createdAt.Add(-time.Second), createdAt.Add(time.Minute), usagestats.UsageLogFilters{UserID: user.ID, UpstreamModelMismatch: &mismatch})
	require.NoError(t, err)
	require.Len(t, groups, 1)
	require.Equal(t, int64(3), groups[0].Requests)

	assertUpstreamResponseModelPartialIndexRecovery(t, ctx, db, unknownLog.ID)
}

func prepareUpstreamResponseModelMigrationReplay(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	_, err := db.ExecContext(ctx, "DROP INDEX CONCURRENTLY IF EXISTS idx_usage_logs_upstream_model_mismatch_created_at")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "ALTER TABLE usage_logs DROP COLUMN IF EXISTS upstream_response_model, DROP COLUMN IF EXISTS upstream_model_mismatch")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "DELETE FROM schema_migrations WHERE filename IN ('246_add_usage_log_upstream_response_model.sql', '247_add_usage_log_upstream_model_mismatch_index_notx.sql')")
	require.NoError(t, err)
}

func stringPtrForResponseAudit(value string) *string { return &value }

func assertUpstreamResponseModelPartialIndexRecovery(t *testing.T, ctx context.Context, db *sql.DB, auditLogID int64) {
	t.Helper()
	const migrationName = "247_add_usage_log_upstream_model_mismatch_index_notx.sql"
	const indexName = "idx_usage_logs_upstream_model_mismatch_created_at"
	_, err := db.ExecContext(ctx, "DELETE FROM schema_migrations WHERE filename = $1", migrationName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "DROP INDEX CONCURRENTLY IF EXISTS "+indexName)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `
		CREATE OR REPLACE FUNCTION response_model_audit_fail_index(value integer) RETURNS integer
		LANGUAGE plpgsql IMMUTABLE AS $$
		BEGIN
			IF value < 0 THEN RAISE EXCEPTION 'intentional index build failure'; END IF;
			RETURN value;
		END;
		$$
	`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "UPDATE usage_logs SET input_tokens = -1 WHERE id = $1", auditLogID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "CREATE INDEX CONCURRENTLY "+indexName+" ON usage_logs (response_model_audit_fail_index(input_tokens))")
	require.Error(t, err)
	_, err = db.ExecContext(ctx, "UPDATE usage_logs SET input_tokens = 7 WHERE id = $1", auditLogID)
	require.NoError(t, err)
	var invalid bool
	require.NoError(t, db.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM pg_class idx JOIN pg_index i ON i.indexrelid = idx.oid
		WHERE idx.relname = $1 AND NOT i.indisvalid
	)`, indexName).Scan(&invalid))
	require.True(t, invalid)
	require.NoError(t, ApplyMigrations(ctx, db))
	var valid bool
	var predicate string
	var definition string
	require.NoError(t, db.QueryRowContext(ctx, `
		SELECT i.indisvalid, COALESCE(pg_get_expr(i.indpred, i.indrelid), ''), pg_get_indexdef(i.indexrelid)
		FROM pg_class idx JOIN pg_index i ON i.indexrelid = idx.oid
		WHERE idx.relname = $1
	`, indexName).Scan(&valid, &predicate, &definition))
	require.True(t, valid)
	require.Contains(t, predicate, "upstream_model_mismatch IS TRUE")
	require.Contains(t, definition, "created_at DESC")
	require.Contains(t, definition, "id DESC")
	_, err = db.ExecContext(ctx, "DROP FUNCTION response_model_audit_fail_index(integer)")
	require.NoError(t, err)
}
