package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/groupbuyplan"
	_ "github.com/Wei-Shaw/sub2api/ent/runtime"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

func TestLockCafeRoomPlanLocksAndRejectsPlanWithoutRoomFulfillment(t *testing.T) {
	var capturedSQL string
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(_ string, actual string) error {
		capturedSQL = actual
		return nil
	})))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	driver := entsql.OpenDB(dialect.Postgres, db)
	client := dbent.NewClient(dbent.Driver(driver))
	t.Cleanup(func() { _ = client.Close() })
	now := time.Date(2026, 8, 21, 18, 30, 0, 0, time.UTC)

	mock.ExpectQuery("locked cafe room plan").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows(groupbuyplan.Columns).AddRow(
			int64(7), now, now, "ordinary plan", nil, "token_pinpinpin", 2, 2, 1.0, 1.0,
			"", "", "", 1, int64(3), []byte("{}"), []byte("[]"), 30, 60, "manual",
			"aggregate_tier", 0.0, 0.0, 0.0, 0.0, false, "balance_credit", nil, "active", 0, nil, nil,
		))

	err = lockCafeRoomPlan(context.Background(), client, 7)
	require.ErrorIs(t, err, service.ErrCafePlanInvalid)
	require.Contains(t, strings.ToUpper(capturedSQL), "FOR UPDATE")
	require.NoError(t, mock.ExpectationsWereMet())
}
