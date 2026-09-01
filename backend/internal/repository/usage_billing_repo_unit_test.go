//go:build unit

package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

const (
	usageBillingUserBalanceSQL     = `(?s)SELECT balance::double precision\s+FROM users\s+WHERE id = \$1 AND deleted_at IS NULL\s+FOR UPDATE`
	usageBillingVoucherSummarySQL  = `(?s)SELECT\s+u\.balance::double precision,\s+COALESCE\(voucher\.available_amount, 0\)::double precision,\s+voucher\.next_expires_at\s+FROM users u`
	usageBillingVoucherRowsSQL     = `(?s)SELECT id, remaining_amount::double precision\s+FROM welfare_vouchers\s+WHERE user_id = \$1`
	usageBillingOverdraftUpdateSQL = `(?s)UPDATE users\s+SET balance = balance - \$1,\s+updated_at = NOW\(\)\s+WHERE id = \$2 AND deleted_at IS NULL\s+RETURNING balance::double precision`
)

func TestDeductWelfareVoucherThenBalance_AllowsUsageBillingOverdraft(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	mock.ExpectQuery(usageBillingUserBalanceSQL).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(0.0007))
	mock.ExpectQuery(usageBillingVoucherRowsSQL).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}))
	mock.ExpectQuery(usageBillingOverdraftUpdateSQL).
		WithArgs(0.06, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(-0.0593))
	mock.ExpectCommit()

	result, err := deductWelfareVoucherThenBalance(ctx, tx, 42, 0.06, welfareVoucherOperationUsageBilling, "req:42", false)
	require.NoError(t, err)
	require.InDelta(t, 0.06, result.BalanceAmount, 0.00000001)
	require.InDelta(t, -0.0593, result.BalanceAfter, 0.00000001)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestApplyUsageBillingEffects_FlagsBalanceOverdraft(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	mock.ExpectQuery(usageBillingUserBalanceSQL).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(0.0007))
	mock.ExpectQuery(usageBillingVoucherRowsSQL).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "remaining_amount"}))
	mock.ExpectQuery(usageBillingOverdraftUpdateSQL).
		WithArgs(0.06, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(-0.0593))
	mock.ExpectCommit()

	result := &service.UsageBillingApplyResult{Applied: true}
	err = (&usageBillingRepository{}).applyUsageBillingEffects(ctx, tx, &service.UsageBillingCommand{
		RequestID:   "req-overdraft",
		APIKeyID:    42,
		UserID:      42,
		BalanceCost: 0.06,
	}, result)
	require.NoError(t, err)
	require.True(t, result.BalanceOverdrafted)
	require.NotNil(t, result.NewBalance)
	require.InDelta(t, -0.0593, *result.NewBalance, 0.00000001)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeductWelfareVoucherThenBalance_RejectsStrictReservation(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	mock.ExpectQuery(usageBillingUserBalanceSQL).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(0.0007))
	mock.ExpectQuery(usageBillingVoucherSummarySQL).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance", "available_amount", "next_expires_at"}).AddRow(0.0007, 0, nil))
	mock.ExpectRollback()

	_, err = deductWelfareVoucherThenBalance(ctx, tx, 42, 0.06, welfareVoucherOperationUsageBilling, "req:42", true)
	require.ErrorIs(t, err, service.ErrInsufficientBalance)
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeductWelfareVoucherThenBalance_ReturnsUserNotFound(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	mock.ExpectQuery(usageBillingUserBalanceSQL).
		WithArgs(int64(42)).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err = deductWelfareVoucherThenBalance(ctx, tx, 42, 0.06, welfareVoucherOperationUsageBilling, "req:42", false)
	require.ErrorIs(t, err, service.ErrUserNotFound)
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageBillingRepository_InsertUsageBillingEntryUsesUsageLogUniqueness(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	mock.ExpectQuery(`(?s)INSERT INTO billing_usage_entries \(\s*usage_log_id, user_id, api_key_id, subscription_id, billing_type, applied, delta_usd\s*\).*ON CONFLICT \(usage_log_id\) DO UPDATE.*delta_usd = billing_usage_entries\.delta_usd \+ EXCLUDED\.delta_usd.*RETURNING id`).
		WithArgs(int64(1001), int64(7), int64(8), nil, int8(2), 0.75).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(501)))
	mock.ExpectCommit()

	inserted, err := (&usageBillingRepository{}).insertUsageBillingEntry(ctx, tx, &service.UsageBillingCommand{
		UsageLogID:         1001,
		RequestID:          "req-trial-overage",
		UserID:             7,
		APIKeyID:           8,
		BillingType:        2,
		BalanceCost:        0.5,
		PrepaidBalanceCost: 0.1,
		SubscriptionCost:   0.15,
	})
	require.NoError(t, err)
	require.True(t, inserted)
	require.NoError(t, tx.Commit())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageBillingRepository_ExistingDedupRepairsMissingLedgerWithoutRededucting(t *testing.T) {
	ctx := context.Background()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	mock.ExpectBegin()
	mock.ExpectQuery(`(?s)INSERT INTO usage_billing_dedup \(request_id, api_key_id, request_fingerprint\).*RETURNING id`).
		WithArgs("req-ledger-repair", int64(8), "fingerprint").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`SELECT request_fingerprint\s+FROM usage_billing_dedup\s+WHERE request_id = \$1 AND api_key_id = \$2`).
		WithArgs("req-ledger-repair", int64(8)).
		WillReturnRows(sqlmock.NewRows([]string{"request_fingerprint"}).AddRow("fingerprint"))
	mock.ExpectExec(`(?s)INSERT INTO billing_usage_entries \(\s*usage_log_id, user_id, api_key_id, subscription_id, billing_type, applied, delta_usd\s*\).*ON CONFLICT \(usage_log_id\) DO NOTHING`).
		WithArgs(int64(1002), int64(7), int64(8), nil, int8(0), 0.75).
		WillReturnResult(sqlmock.NewResult(502, 1))
	mock.ExpectCommit()

	result, err := (&usageBillingRepository{db: db}).Apply(ctx, &service.UsageBillingCommand{
		UsageLogID:         1002,
		RequestID:          "req-ledger-repair",
		APIKeyID:           8,
		RequestFingerprint: "fingerprint",
		UserID:             7,
		BalanceCost:        0.75,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.Applied)
	require.NoError(t, mock.ExpectationsWereMet())
}
