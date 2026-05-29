package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestCanvasRepositoryCancelCanvasRunRestrictsUserAndActiveStatuses(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	now := time.Now()
	canceledAt := now.Add(time.Minute)
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "canvas_id", "status", "trigger_type", "api_key_id", "model",
		"input", "output", "error_message", "metadata", "started_at", "completed_at",
		"canceled_at", "created_at", "updated_at",
	}).AddRow(
		int64(30),
		int64(42),
		sql.NullInt64{Int64: 12, Valid: true},
		service.CanvasRunStatusCanceled,
		"manual",
		sql.NullInt64{Int64: 44, Valid: true},
		"gpt-image-2",
		[]byte(`{}`),
		[]byte(`{"ok":true}`),
		"",
		[]byte(`{}`),
		sql.NullTime{Time: now, Valid: true},
		sql.NullTime{Time: canceledAt, Valid: true},
		sql.NullTime{Time: canceledAt, Valid: true},
		now,
		canceledAt,
	)
	mock.ExpectQuery(`UPDATE canvas_runs\s+SET status = \$3,\s+canceled_at = COALESCE\(canceled_at, NOW\(\)\),\s+completed_at = COALESCE\(completed_at, NOW\(\)\),\s+updated_at = NOW\(\)\s+WHERE id = \$1 AND user_id = \$2 AND status IN \(\$4, \$5\)`).
		WithArgs(
			int64(30),
			int64(42),
			service.CanvasRunStatusCanceled,
			service.CanvasRunStatusPending,
			service.CanvasRunStatusRunning,
		).
		WillReturnRows(rows)

	repo := &canvasRepository{sql: db}
	run, err := repo.CancelCanvasRun(context.Background(), 42, 30)

	require.NoError(t, err)
	require.Equal(t, int64(30), run.ID)
	require.Equal(t, int64(42), run.UserID)
	require.Equal(t, service.CanvasRunStatusCanceled, run.Status)
	require.NotNil(t, run.CanceledAt)
	require.NotNil(t, run.CompletedAt)
	require.NoError(t, mock.ExpectationsWereMet())
}
