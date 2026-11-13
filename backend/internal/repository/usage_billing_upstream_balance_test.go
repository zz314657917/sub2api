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

func TestDeductUsageBillingBalance_UpstreamOutcomes(t *testing.T) {
	for _, tc := range []struct {
		name        string
		firstErr    error
		fallbackErr error
		balance     float64
		wantErr     error
		sufficient  bool
	}{
		{name: "sufficient", balance: 90, sufficient: true},
		{name: "exact balance", balance: 0, sufficient: true},
		{name: "completed usage records debt", firstErr: sql.ErrNoRows, balance: -5},
		{name: "deleted user", firstErr: sql.ErrNoRows, fallbackErr: sql.ErrNoRows, wantErr: service.ErrUserNotFound},
		{name: "first update failure", firstErr: sql.ErrConnDone, wantErr: sql.ErrConnDone},
		{name: "fallback failure", firstErr: sql.ErrNoRows, fallbackErr: sql.ErrConnDone, wantErr: sql.ErrConnDone},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			mock.ExpectBegin()
			tx, err := db.BeginTx(context.Background(), nil)
			require.NoError(t, err)
			q := mock.ExpectQuery(usageBillingAtomicBalanceSQL).WithArgs(10.0, int64(42))
			if tc.firstErr != nil {
				q.WillReturnError(tc.firstErr)
			} else {
				q.WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(tc.balance))
			}
			if tc.firstErr == sql.ErrNoRows {
				q = mock.ExpectQuery(usageBillingAtomicOverdraftSQL).WithArgs(10.0, int64(42))
				if tc.fallbackErr != nil {
					q.WillReturnError(tc.fallbackErr)
				} else {
					q.WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(tc.balance))
				}
			}
			balance, sufficient, err := deductUsageBillingBalance(context.Background(), tx, 42, 10)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.balance, balance)
				require.Equal(t, tc.sufficient, sufficient)
			}
			mock.ExpectRollback()
			require.NoError(t, tx.Rollback())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUsageBillingRepositoryApply_StrictBalanceReservation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		updateErr error
		existsErr error
		wantErr   error
	}{
		{name: "success"},
		{name: "insufficient", updateErr: sql.ErrNoRows, wantErr: service.ErrInsufficientBalance},
		{name: "deleted user", updateErr: sql.ErrNoRows, existsErr: sql.ErrNoRows, wantErr: service.ErrUserNotFound},
		{name: "update error", updateErr: sql.ErrConnDone, wantErr: sql.ErrConnDone},
		{name: "existence error", updateErr: sql.ErrNoRows, existsErr: sql.ErrConnDone, wantErr: sql.ErrConnDone},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			cmd := &service.UsageBillingCommand{RequestID: "strict", UserID: 42, APIKeyID: 7, UsageLogID: 1001, BalanceCost: 10, RequireBalanceCheck: true}
			expectDeletedKeyBillingClaim(mock, cmd)
			mock.ExpectQuery(`(?s)INSERT INTO billing_usage_entries .*RETURNING id`).
				WithArgs(int64(1001), "strict", int64(42), int64(7), nil, int8(0), 10.0).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			q := mock.ExpectQuery(usageBillingAtomicBalanceSQL).WithArgs(10.0, int64(42))
			if tc.updateErr != nil {
				q.WillReturnError(tc.updateErr)
			} else {
				q.WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(0.0))
			}
			if tc.updateErr == sql.ErrNoRows {
				q = mock.ExpectQuery(`SELECT 1 FROM users WHERE id = \$1 AND deleted_at IS NULL`).WithArgs(int64(42))
				if tc.existsErr != nil {
					q.WillReturnError(tc.existsErr)
				} else {
					q.WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))
				}
			}
			if tc.wantErr != nil {
				mock.ExpectRollback()
			} else {
				mock.ExpectCommit()
			}
			result, err := (&usageBillingRepository{db: db}).Apply(context.Background(), cmd)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.True(t, result.Applied)
				require.Zero(t, *result.NewBalance)
				require.Equal(t, 10.0, result.BalanceCost)
				require.Zero(t, result.VoucherCost)
				require.False(t, result.BalanceOverdrafted)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
