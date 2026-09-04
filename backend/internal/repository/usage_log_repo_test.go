package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type cafeLobbyRecorderStub struct {
	userID     int64
	occurredAt time.Time
	count      int
}

func (r *cafeLobbyRecorderStub) RecordPersistedUsage(userID int64, occurredAt time.Time) {
	r.userID = userID
	r.occurredAt = occurredAt
	r.count++
}

func TestUsageLogRepositoryRecordsCafeLobbyOnlyAfterInsert(t *testing.T) {
	recorder := &cafeLobbyRecorderStub{}
	repo := &usageLogRepository{lobbyUsageRecorder: recorder}
	createdAt := time.Date(2026, 8, 3, 4, 0, 0, 0, time.UTC)
	repo.recordPersistedUsage(17, createdAt)
	repo.recordPersistedUsage(0, createdAt)

	require.Equal(t, 1, recorder.count)
	require.Equal(t, int64(17), recorder.userID)
	require.Equal(t, createdAt, recorder.occurredAt)
	var _ service.CafeLobbyUsageRecorder = recorder
}

func TestUsageLogRepositoryCreateSingleRecordsCafeLobbyOnlyOnInsert(t *testing.T) {
	createdAt := time.Date(2026, 8, 3, 4, 0, 0, 0, time.UTC)
	t.Run("inserted", func(t *testing.T) {
		db, mock := newSQLMock(t)
		recorder := &cafeLobbyRecorderStub{}
		repo := newUsageLogRepositoryWithSQL(nil, db)
		repo.lobbyUsageRecorder = recorder
		log := &service.UsageLog{UserID: 17, APIKeyID: 3, AccountID: 5, RequestID: "s155-inserted", Model: "gpt-5", CreatedAt: createdAt}
		prepared := prepareUsageLogInsert(log)
		mock.ExpectQuery("INSERT INTO usage_logs").
			WithArgs(anySliceToDriverValues(appendUsageStatsTimezoneArg(prepared.args))...).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(8), createdAt))

		inserted, err := repo.createSingle(context.Background(), db, log)
		require.NoError(t, err)
		require.True(t, inserted)
		require.Equal(t, 1, recorder.count)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("idempotent conflict", func(t *testing.T) {
		db, mock := newSQLMock(t)
		recorder := &cafeLobbyRecorderStub{}
		repo := newUsageLogRepositoryWithSQL(nil, db)
		repo.lobbyUsageRecorder = recorder
		log := &service.UsageLog{UserID: 17, APIKeyID: 3, AccountID: 5, RequestID: "s155-conflict", Model: "gpt-5", CreatedAt: createdAt}
		prepared := prepareUsageLogInsert(log)
		mock.ExpectQuery("INSERT INTO usage_logs").
			WithArgs(anySliceToDriverValues(appendUsageStatsTimezoneArg(prepared.args))...).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}))
		mock.ExpectQuery("SELECT id, created_at FROM usage_logs").
			WithArgs(log.RequestID, log.APIKeyID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(8), createdAt))

		inserted, err := repo.createSingle(context.Background(), db, log)
		require.NoError(t, err)
		require.False(t, inserted)
		require.Zero(t, recorder.count)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBuildUsageLogBestEffortStateQueryReturnsPerInputState(t *testing.T) {
	prepared := prepareUsageLogInsert(&service.UsageLog{UserID: 17, APIKeyID: 3, RequestID: "s155-state", CreatedAt: time.Date(2026, 8, 3, 4, 0, 0, 0, time.UTC)})
	query, args := buildUsageLogBestEffortInsertStateQuery([]usageLogInsertPrepared{prepared})
	require.Contains(t, query, "input_idx")
	require.Contains(t, query, "'inserted', resolved.inserted")
	require.Len(t, args, len(prepared.args)+2)
	require.Equal(t, 0, args[0])
	require.Equal(t, prepared.args[0], args[1])
	require.Equal(t, usageStatsTimezoneName(), args[len(args)-1])
}

func TestDashboardAggregatedStatsReadsSettledCostFromAggregateTable(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := newUsageLogRepositoryWithSQL(nil, db)
	stats := &DashboardStats{}
	today := time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC)
	now := today.Add(10 * time.Minute)

	mock.ExpectQuery(`(?s)COALESCE\(SUM\(actual_cost\), 0\) as total_actual_cost,\s+COALESCE\(SUM\(account_cost\), 0\).*FROM usage_dashboard_daily`).
		WillReturnRows(sqlmock.NewRows([]string{
			"total_requests", "total_input_tokens", "total_output_tokens",
			"total_cache_creation_tokens", "total_cache_read_tokens", "total_cost",
			"total_actual_cost", "total_account_cost", "total_duration_ms",
		}).AddRow(2, 10, 20, 3, 4, 1.5, 1.2, 1.5, 100))
	mock.ExpectQuery(`(?s)FROM usage_dashboard_daily\s+WHERE bucket_date = \$1::date`).
		WithArgs(today).
		WillReturnRows(sqlmock.NewRows([]string{
			"today_requests", "today_input_tokens", "today_output_tokens",
			"today_cache_creation_tokens", "today_cache_read_tokens", "today_cost",
			"today_actual_cost", "today_account_cost", "active_users",
		}).AddRow(2, 10, 20, 3, 4, 1.5, 1.2, 1.5, 1))
	mock.ExpectQuery(`(?s)SELECT active_users\s+FROM usage_dashboard_hourly`).
		WithArgs(now.In(timezone.Location()).Truncate(time.Hour)).
		WillReturnRows(sqlmock.NewRows([]string{"active_users"}).AddRow(1))

	require.NoError(t, repo.fillDashboardUsageStatsAggregated(context.Background(), stats, today, now))
	require.Equal(t, 1.2, stats.TotalActualCost)
	require.Equal(t, 1.2, stats.TodayActualCost)
	require.NoError(t, mock.ExpectationsWereMet())
}
