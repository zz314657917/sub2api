package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAccountRepositoryListShareSummary(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &accountRepository{sql: db}

	rows := sqlmock.NewRows([]string{
		"frozen_amount",
		"available_amount",
		"transferred_amount",
		"total_amount",
		"count_frozen",
		"count_available",
		"count_transferred",
	}).AddRow(1.25, 2.5, 3.75, 7.5, int64(1), int64(2), int64(3))

	mock.ExpectQuery("FROM account_share_ledger").
		WithArgs(int64(42)).
		WillReturnRows(rows)

	summary, err := repo.ListShareSummary(context.Background(), 42)
	require.NoError(t, err)
	require.NotNil(t, summary)
	require.Equal(t, int64(42), summary.OwnerUserID)
	require.InDelta(t, 1.25, summary.FrozenAmount, 1e-9)
	require.InDelta(t, 2.5, summary.AvailableAmount, 1e-9)
	require.InDelta(t, 3.75, summary.TransferredAmount, 1e-9)
	require.InDelta(t, 7.5, summary.TotalAmount, 1e-9)
	require.Equal(t, int64(1), summary.CountFrozen)
	require.Equal(t, int64(2), summary.CountAvailable)
	require.Equal(t, int64(3), summary.CountTransferred)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountRepositoryGetUsageSummary(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &accountRepository{sql: db}
	startTime := time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"total_accounts",
		"private_accounts",
		"public_pending_accounts",
		"public_active_accounts",
		"public_suspended_accounts",
		"own_usage_cost",
		"own_usage_requests",
		"shared_usage_cost",
		"shared_usage_requests",
		"share_income",
		"platform_amount",
		"account_cost",
		"balance_deduction",
	}).AddRow(
		int64(4),
		int64(1),
		int64(1),
		int64(2),
		int64(0),
		1.25,
		int64(3),
		2.5,
		int64(4),
		0.75,
		0.25,
		4.5,
		0.5,
	)

	mock.ExpectQuery("WITH owned_accounts AS").
		WithArgs(int64(42), startTime, endTime).
		WillReturnRows(rows)

	summary, err := repo.GetUsageSummary(context.Background(), 42, startTime, endTime)
	require.NoError(t, err)
	require.NotNil(t, summary)
	require.Equal(t, int64(42), summary.OwnerUserID)
	require.Equal(t, int64(4), summary.TotalAccounts)
	require.Equal(t, int64(1), summary.PrivateAccounts)
	require.Equal(t, int64(1), summary.PublicPendingAccounts)
	require.Equal(t, int64(2), summary.PublicActiveAccounts)
	require.Equal(t, int64(0), summary.PublicSuspendedAccounts)
	require.InDelta(t, 1.25, summary.OwnUsageCost, 1e-9)
	require.Equal(t, int64(3), summary.OwnUsageRequests)
	require.InDelta(t, 2.5, summary.SharedUsageCost, 1e-9)
	require.Equal(t, int64(4), summary.SharedUsageRequests)
	require.InDelta(t, 0.75, summary.ShareIncome, 1e-9)
	require.InDelta(t, 0.25, summary.PlatformAmount, 1e-9)
	require.InDelta(t, 4.5, summary.AccountCost, 1e-9)
	require.InDelta(t, 0.5, summary.BalanceDeduction, 1e-9)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountRepositoryTransferAvailableShareToBalance(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &accountRepository{sql: db}

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE account_share_ledger").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"owner_amount"}).AddRow(1.25).AddRow(2.75))
	mock.ExpectQuery("UPDATE users").
		WithArgs(4.0, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(14.0))
	mock.ExpectCommit()

	transferred, balance, err := repo.TransferAvailableShareToBalance(context.Background(), 42)
	require.NoError(t, err)
	require.InDelta(t, 4.0, transferred, 1e-9)
	require.InDelta(t, 14.0, balance, 1e-9)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountRepositoryTransferAvailableShareToBalance_NoAvailableRows(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &accountRepository{sql: db}

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE account_share_ledger").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"owner_amount"}))
	mock.ExpectCommit()

	transferred, balance, err := repo.TransferAvailableShareToBalance(context.Background(), 42)
	require.NoError(t, err)
	require.Zero(t, transferred)
	require.Zero(t, balance)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountRepositoryTransferAvailableShareToBalance_InvalidUser(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &accountRepository{sql: db}

	transferred, balance, err := repo.TransferAvailableShareToBalance(context.Background(), 0)
	require.NoError(t, err)
	require.Zero(t, transferred)
	require.Zero(t, balance)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountRepositoryListShareSummary_InvalidUser(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &accountRepository{sql: db}

	summary, err := repo.ListShareSummary(context.Background(), 0)
	require.NoError(t, err)
	require.NotNil(t, summary)
	require.Equal(t, int64(0), summary.OwnerUserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountRepositoryGetUsageSummary_InvalidUser(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &accountRepository{sql: db}

	summary, err := repo.GetUsageSummary(context.Background(), 0, time.Time{}, time.Time{})
	require.NoError(t, err)
	require.NotNil(t, summary)
	require.Equal(t, int64(0), summary.OwnerUserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountRepositoryTransferAvailableShareToBalance_UserNotFound(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &accountRepository{sql: db}

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE account_share_ledger").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"owner_amount"}).AddRow(2.0))
	mock.ExpectQuery("UPDATE users").
		WithArgs(2.0, int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}))
	mock.ExpectRollback()

	transferred, balance, err := repo.TransferAvailableShareToBalance(context.Background(), 42)
	require.ErrorIs(t, err, service.ErrUserNotFound)
	require.Zero(t, transferred)
	require.Zero(t, balance)
	require.NoError(t, mock.ExpectationsWereMet())
}
