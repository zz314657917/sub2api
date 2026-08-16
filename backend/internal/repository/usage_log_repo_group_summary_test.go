package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGroupUsageSummary(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := newUsageLogRepositoryWithSQL(nil, db)
	todayStart := service.GroupUsageTodayStart(time.Date(2026, 8, 15, 16, 0, 0, 0, time.UTC))
	mock.ExpectQuery("WITH state_values AS").
		WillReturnRows(sqlmock.NewRows([]string{"group_id", "total_cost", "today_cost", "yesterday_cost"}).
			AddRow(int64(10), 12.5, 3.5, 2.5))

	result, err := repo.GetAllGroupUsageSummary(context.Background(), todayStart)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, int64(10), result[0].GroupID)
	require.InDelta(t, 12.5, result[0].TotalCost, 0.000001)
	require.InDelta(t, 3.5, result[0].TodayCost, 0.000001)
	require.InDelta(t, 2.5, result[0].YesterdayCost, 0.000001)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsageLogRepositoryGetAllGroupUsageSummaryUsesRollupTail(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := newUsageLogRepositoryWithSQL(nil, db)
	todayStart := service.GroupUsageTodayStart(time.Date(2026, 8, 15, 16, 0, 0, 0, time.UTC))
	mock.ExpectQuery("historical AS").
		WillReturnRows(sqlmock.NewRows([]string{"group_id", "total_cost", "today_cost", "yesterday_cost"}).
			AddRow(int64(11), 9.0, 4.0, 3.0))

	result, err := repo.GetAllGroupUsageSummary(context.Background(), todayStart)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.InDelta(t, 9, result[0].TotalCost, 0.000001)
	require.InDelta(t, 4, result[0].TodayCost, 0.000001)
	require.InDelta(t, 3, result[0].YesterdayCost, 0.000001)
	require.NoError(t, mock.ExpectationsWereMet())
}
