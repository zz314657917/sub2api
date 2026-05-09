package repository

import (
	"context"
	"testing"

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
