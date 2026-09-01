package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAccountRepositoryResetQuotaUsedResetsShareDisplayWindowBaseline(t *testing.T) {
	exec := &recordingSQLExecutor{}
	repo := &accountRepository{sql: exec}

	require.NoError(t, repo.ResetQuotaUsed(context.Background(), 42))
	require.Len(t, exec.queries, 2)

	resetSQL := exec.queries[0]
	require.Contains(t, resetSQL, "share_display_5h_used")
	require.Contains(t, resetSQL, "share_display_5h_start")
	require.Contains(t, resetSQL, "share_display_7d_used")
	require.Contains(t, resetSQL, "share_display_7d_start")
	require.Contains(t, resetSQL, "rate_limited_at = NULL")
	require.Contains(t, resetSQL, "rate_limit_reset_at = NULL")
	require.NotContains(t, resetSQL, "codex_5h_used_percent")
	require.NotContains(t, resetSQL, "codex_7d_used_percent")

	require.Contains(t, exec.queries[1], "INSERT INTO scheduler_outbox")
	require.Equal(t, []any{int64(42)}, exec.args[0])
}

type recordingSQLExecutor struct {
	queries []string
	args    [][]any
	err     error
	result  sql.Result
}

func (e *recordingSQLExecutor) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	e.queries = append(e.queries, query)
	e.args = append(e.args, append([]any(nil), args...))
	if e.err != nil {
		return nil, e.err
	}
	if e.result != nil {
		return e.result, nil
	}
	return sqlmock.NewResult(1, 1), nil
}

func (e *recordingSQLExecutor) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	panic("unexpected QueryContext")
}

func TestAccountRepositoryResetQuotaUsedSucceeds(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &accountRepository{sql: db}

	mock.ExpectExec("UPDATE accounts SET extra =").
		WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO scheduler_outbox").
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, repo.ResetQuotaUsed(context.Background(), 42))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAccountRepositoryResetQuotaUsedReturnsNotFoundWithoutOutbox(t *testing.T) {
	exec := &recordingSQLExecutor{result: sqlmock.NewResult(0, 0)}
	repo := &accountRepository{sql: exec}

	require.ErrorIs(t, repo.ResetQuotaUsed(context.Background(), 42), service.ErrAccountNotFound)
	require.Len(t, exec.queries, 1)
	require.NotContains(t, exec.queries[0], "INSERT INTO scheduler_outbox")
}

func TestAccountRepositoryGetAccountUsageCostsSinceByWindowUsesPerWindowStarts(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &accountRepository{sql: db}

	mock.ExpectQuery("WITH windows").
		WithArgs(int64(1), "5h", sqlmock.AnyArg(), int64(1), "7d", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "suffix", "account_cost"}).
			AddRow(int64(1), "5h", 2.5).
			AddRow(int64(1), "7d", 4.0))

	got, err := repo.GetAccountUsageCostsSinceByWindow(context.Background(), []service.AccountUsageCostWindowRequest{
		{AccountID: 1, Suffix: "5h"},
		{AccountID: 1, Suffix: "7d"},
	})
	require.NoError(t, err)
	require.Equal(t, 2.5, got[service.AccountUsageCostWindowRequestKey{AccountID: 1, Suffix: "5h"}])
	require.Equal(t, 4.0, got[service.AccountUsageCostWindowRequestKey{AccountID: 1, Suffix: "7d"}])
	require.NoError(t, mock.ExpectationsWereMet())
}
