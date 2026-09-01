//go:build unit

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSafeDateFormat(t *testing.T) {
	tests := []struct {
		name        string
		granularity string
		expected    string
	}{
		// 合法值
		{"hour", "hour", "YYYY-MM-DD HH24:00"},
		{"day", "day", "YYYY-MM-DD"},
		{"week", "week", "IYYY-IW"},
		{"month", "month", "YYYY-MM"},

		// 非法值回退到默认
		{"空字符串", "", "YYYY-MM-DD"},
		{"未知粒度 year", "year", "YYYY-MM-DD"},
		{"未知粒度 minute", "minute", "YYYY-MM-DD"},

		// 恶意字符串
		{"SQL 注入尝试", "'; DROP TABLE users; --", "YYYY-MM-DD"},
		{"带引号", "day'", "YYYY-MM-DD"},
		{"带括号", "day)", "YYYY-MM-DD"},
		{"Unicode", "日", "YYYY-MM-DD"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := safeDateFormat(tc.granularity)
			require.Equal(t, tc.expected, got, "safeDateFormat(%q)", tc.granularity)
		})
	}
}

func TestBuildUsageLogBatchInsertQuery_UsesConflictDoNothing(t *testing.T) {
	log := &service.UsageLog{
		UserID:       1,
		APIKeyID:     2,
		AccountID:    3,
		RequestID:    "req-batch-no-update",
		Model:        "gpt-5",
		InputTokens:  10,
		OutputTokens: 5,
		TotalCost:    1.2,
		ActualCost:   1.2,
		CreatedAt:    time.Now().UTC(),
	}
	prepared := prepareUsageLogInsert(log)

	query, _ := buildUsageLogBatchInsertQuery([]string{usageLogBatchKey(log.RequestID, log.APIKeyID)}, map[string]usageLogInsertPrepared{
		usageLogBatchKey(log.RequestID, log.APIKeyID): prepared,
	})

	require.Contains(t, query, "ON CONFLICT (request_id, api_key_id) DO NOTHING")
	require.Contains(t, query, "ON CONFLICT (user_id, usage_date) DO UPDATE")
}

func TestUsageLogRepository_CreatePendingPayloadDuplicateAppliedLogDoesNotCreateOutbox(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	// The usage-log insert conflicts on the original request, then createSingle
	// resolves the durable row that has already completed settlement.
	mock.ExpectQuery(`(?s)WITH inserted AS \(\s*INSERT INTO usage_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}))
	mock.ExpectQuery(`SELECT id, created_at FROM usage_logs WHERE request_id = \$1 AND api_key_id = \$2`).
		WithArgs("req-already-applied", int64(19)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(9001), time.Now().UTC()))
	mock.ExpectExec(`(?s)UPDATE usage_logs\s+SET billing_status = 'pending', billing_error = NULL\s+WHERE id = \$1 AND billing_status <> 'applied'`).
		WithArgs(int64(9001)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	log := &service.UsageLog{
		UserID:        13,
		APIKeyID:      19,
		AccountID:     23,
		RequestID:     "req-already-applied",
		Model:         "gpt-5",
		BillingStatus: service.BillingSettlementApplied,
		BillingError:  billingStringPtr("previous failure must be cleared when applied"),
		CreatedAt:     time.Now().UTC(),
	}
	payload := &service.UsageBillingSettlementPayload{
		Version: 1,
		Primary: service.UsageBillingCommand{
			RequestID: "req-already-applied",
			APIKeyID:  19,
			UserID:    13,
		},
	}

	owned, settled, err := (&usageLogRepository{db: db}).CreatePendingPayloadWithOwnership(ctx, log, payload)
	require.NoError(t, err)
	require.False(t, owned)
	require.True(t, settled)
	require.Equal(t, int64(9001), log.ID)
	require.Equal(t, service.BillingSettlementApplied, log.BillingStatus)
	require.Nil(t, log.BillingError)
	require.Equal(t, int64(9001), payload.Primary.UsageLogID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageLogRepository_CreatePendingPayloadDuplicateFailedLogReclaimsOutboxLease(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)WITH inserted AS \(\s*INSERT INTO usage_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}))
	mock.ExpectQuery(`SELECT id, created_at FROM usage_logs WHERE request_id = \$1 AND api_key_id = \$2`).
		WithArgs("req-replay-failed", int64(29)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(9002), time.Now().UTC()))
	mock.ExpectExec(`(?s)UPDATE usage_logs\s+SET billing_status = 'pending', billing_error = NULL\s+WHERE id = \$1 AND billing_status <> 'applied'`).
		WithArgs(int64(9002)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)INSERT INTO usage_billing_settlement_outbox \(usage_log_id, request_id, api_key_id, payload, status, available_at, lease_until\).*VALUES .*'processing'.*ON CONFLICT \(usage_log_id\) DO UPDATE.*lease_until = NOW\(\) \+ \(30 \* INTERVAL '1 second'\).*WHERE usage_billing_settlement_outbox.status IN \('pending', 'failed'\)`).
		WithArgs(int64(9002), "req-replay-failed", int64(29), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	log := &service.UsageLog{UserID: 13, APIKeyID: 29, AccountID: 23, RequestID: "req-replay-failed", Model: "gpt-5", BillingStatus: service.BillingSettlementFailed, CreatedAt: time.Now().UTC()}
	payload := &service.UsageBillingSettlementPayload{Version: 1, Primary: service.UsageBillingCommand{RequestID: "req-replay-failed", APIKeyID: 29, UserID: 13}}

	owned, settled, err := (&usageLogRepository{db: db}).CreatePendingPayloadWithOwnership(ctx, log, payload)
	require.NoError(t, err)
	require.True(t, owned)
	require.False(t, settled)
	require.Equal(t, service.BillingSettlementPending, log.BillingStatus)
	require.Nil(t, log.BillingError)
	require.Equal(t, int64(9002), payload.Primary.UsageLogID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageLogRepository_ClaimPendingReclaimsExpiredProcessingLease(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)FROM usage_billing_settlement_outbox\s+WHERE \(status = 'pending' AND available_at <= NOW\(\)\)\s+OR \(status = 'processing' AND lease_until <= NOW\(\)\).*SET status = 'processing'.*RETURNING o.id, o.usage_log_id, o.payload, o.attempts`).
		WithArgs(1, int64(60)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "usage_log_id", "payload", "attempts"}).AddRow(
			int64(9010), int64(9011), []byte(`{"version":1,"primary":{"RequestID":"req-expired"}}`), 2,
		))
	mock.ExpectCommit()

	tasks, err := (&usageLogRepository{db: db}).ClaimPending(ctx, 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	require.Equal(t, int64(9011), tasks[0].UsageLogID)
	require.Equal(t, "req-expired", tasks[0].Command.RequestID)
	require.Equal(t, 2, tasks[0].Attempts)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageLogRepository_CreatePendingPayloadDuplicatePendingLogReclaimsOutboxLease(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)WITH inserted AS \(\s*INSERT INTO usage_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}))
	mock.ExpectQuery(`SELECT id, created_at FROM usage_logs WHERE request_id = \$1 AND api_key_id = \$2`).
		WithArgs("req-replay-pending", int64(31)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(9003), time.Now().UTC()))
	mock.ExpectExec(`(?s)UPDATE usage_logs\s+SET billing_status = 'pending', billing_error = NULL\s+WHERE id = \$1 AND billing_status <> 'applied'`).
		WithArgs(int64(9003)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)INSERT INTO usage_billing_settlement_outbox .*ON CONFLICT \(usage_log_id\) DO UPDATE.*status IN \('pending', 'failed'\)`).
		WithArgs(int64(9003), "req-replay-pending", int64(31), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	log := &service.UsageLog{UserID: 13, APIKeyID: 31, AccountID: 23, RequestID: "req-replay-pending", Model: "gpt-5", BillingStatus: service.BillingSettlementPending, CreatedAt: time.Now().UTC()}
	payload := &service.UsageBillingSettlementPayload{Version: 1, Primary: service.UsageBillingCommand{RequestID: "req-replay-pending", APIKeyID: 31, UserID: 13}}

	owned, settled, err := (&usageLogRepository{db: db}).CreatePendingPayloadWithOwnership(ctx, log, payload)
	require.NoError(t, err)
	require.True(t, owned)
	require.False(t, settled)
	require.Equal(t, service.BillingSettlementPending, log.BillingStatus)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageLogRepository_CreatePendingPayloadProcessingLeaseIsNotOwned(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)WITH inserted AS \(\s*INSERT INTO usage_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}))
	mock.ExpectQuery(`SELECT id, created_at FROM usage_logs WHERE request_id = \$1 AND api_key_id = \$2`).
		WithArgs("req-processing", int64(33)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(9004), time.Now().UTC()))
	mock.ExpectExec(`(?s)UPDATE usage_logs\s+SET billing_status = 'pending', billing_error = NULL\s+WHERE id = \$1 AND billing_status <> 'applied'`).
		WithArgs(int64(9004)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)INSERT INTO usage_billing_settlement_outbox .*ON CONFLICT \(usage_log_id\) DO UPDATE.*status IN \('pending', 'failed'\)`).
		WithArgs(int64(9004), "req-processing", int64(33), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	log := &service.UsageLog{UserID: 13, APIKeyID: 33, AccountID: 23, RequestID: "req-processing", Model: "gpt-5", BillingStatus: service.BillingSettlementPending, CreatedAt: time.Now().UTC()}
	payload := &service.UsageBillingSettlementPayload{Version: 1, Primary: service.UsageBillingCommand{RequestID: "req-processing", APIKeyID: 33, UserID: 13}}

	owned, settled, err := (&usageLogRepository{db: db}).CreatePendingPayloadWithOwnership(ctx, log, payload)
	require.NoError(t, err)
	require.False(t, owned)
	require.False(t, settled)
	require.Equal(t, service.BillingSettlementPending, log.BillingStatus)
	require.NoError(t, mock.ExpectationsWereMet())
}

func billingStringPtr(value string) *string {
	return &value
}
