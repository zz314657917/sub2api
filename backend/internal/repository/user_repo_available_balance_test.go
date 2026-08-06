package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func newAvailableBalanceRepo(t *testing.T, capturedSQL *string) (*userRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(_ string, actual string) error {
		*capturedSQL = actual
		return nil
	})))
	require.NoError(t, err)

	driver := entsql.OpenDB(dialect.Postgres, db)
	client := dbent.NewClient(dbent.Driver(driver))
	t.Cleanup(func() {
		_ = client.Close()
		_ = db.Close()
	})
	return &userRepository{client: client}, mock
}

func TestUserRepositoryDeductAvailableBalance(t *testing.T) {
	t.Run("locks the user and returns only available balance", func(t *testing.T) {
		var capturedSQL string
		repo, mock := newAvailableBalanceRepo(t, &capturedSQL)
		mock.ExpectQuery("deduct available balance").
			WithArgs(10.0, int64(7)).
			WillReturnRows(sqlmock.NewRows([]string{"deducted"}).AddRow(4.0))

		deducted, err := repo.DeductAvailableBalance(context.Background(), 7, 10)
		require.NoError(t, err)
		require.Equal(t, 4.0, deducted)
		require.Contains(t, strings.ToUpper(capturedSQL), "FOR UPDATE")
		require.Contains(t, capturedSQL, "LEAST($1, GREATEST(target.balance, 0))")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns not found when no current user row exists", func(t *testing.T) {
		var capturedSQL string
		repo, mock := newAvailableBalanceRepo(t, &capturedSQL)
		mock.ExpectQuery("deduct available balance").
			WithArgs(10.0, int64(7)).
			WillReturnRows(sqlmock.NewRows([]string{"deducted"}))

		deducted, err := repo.DeductAvailableBalance(context.Background(), 7, 10)
		require.Zero(t, deducted)
		require.ErrorIs(t, err, service.ErrUserNotFound)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rejects negative deductions without querying", func(t *testing.T) {
		var capturedSQL string
		repo, mock := newAvailableBalanceRepo(t, &capturedSQL)

		deducted, err := repo.DeductAvailableBalance(context.Background(), 7, -1)
		require.Zero(t, deducted)
		require.ErrorContains(t, err, "deduction amount must be nonnegative")
		require.Empty(t, capturedSQL)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
