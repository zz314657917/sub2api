package repository

import (
	"context"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type exactTimeArg struct {
	want time.Time
}

func (a exactTimeArg) Match(value driver.Value) bool {
	got, ok := value.(time.Time)
	return ok && got.Equal(a.want)
}

func TestStudioBridgeRepositoryResolveChargeUsageRefsUsesDefaultKey(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	repo := &studioBridgeRepository{db: db}
	mock.ExpectQuery(`WITH existing_key AS`).
		WithArgs(int64(42), sqlmock.AnyArg(), service.DefaultAPIKeyName, "text").
		WillReturnRows(sqlmock.NewRows([]string{"id", "group_id", "account_id"}).AddRow(int64(77), nil, int64(88)))

	refs, err := repo.resolveChargeUsageRefs(context.Background(), db, service.StudioBridgeChargeCommand{
		UserID: 42,
		Mode:   "chat",
	})
	require.NoError(t, err)
	require.Equal(t, int64(77), refs.apiKeyID)
	require.Equal(t, int64(88), refs.accountID)
	require.False(t, refs.groupID.Valid)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStudioBridgeRepositoryResolveChargeUsageRefsPreservesDefaultKeyGroup(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	repo := &studioBridgeRepository{db: db}
	mock.ExpectQuery(`WITH existing_key AS`).
		WithArgs(int64(42), sqlmock.AnyArg(), service.DefaultAPIKeyName, "image").
		WillReturnRows(sqlmock.NewRows([]string{"id", "group_id", "account_id"}).AddRow(int64(77), int64(9), int64(88)))

	refs, err := repo.resolveChargeUsageRefs(context.Background(), db, service.StudioBridgeChargeCommand{
		UserID: 42,
		Mode:   "edit",
		Model:  "gpt-image-2",
	})
	require.NoError(t, err)
	require.Equal(t, int64(77), refs.apiKeyID)
	require.Equal(t, int64(88), refs.accountID)
	require.True(t, refs.groupID.Valid)
	require.Equal(t, int64(9), refs.groupID.Int64)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStudioBridgeRepositoryCommitWritesUsageDurationFromReserveTime(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	repo := &studioBridgeRepository{db: db}
	reserveCreatedAt := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	cmd := service.StudioBridgeChargeCommand{
		AppID:     service.StudioBridgeAppLuoyeAI,
		UserID:    42,
		ChargeKey: "task:image:precharge",
		Amount:    0.8,
		TaskID:    "image-task",
		Mode:      "generate",
		Model:     "gpt-image-2",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`FROM studio_bridge_charges[\s\S]*FOR UPDATE`).
		WithArgs(service.StudioBridgeAppLuoyeAI, cmd.ChargeKey).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"app_id",
			"charge_key",
			"refund_for_charge_key",
			"user_id",
			"amount",
			"refunded_amount",
			"status",
			"fingerprint",
			"reason",
			"task_id",
			"mode",
			"model",
			"actor_user_id",
			"team_id",
			"balance_after",
			"usage_logged_at",
			"created_at",
		}).AddRow(
			int64(1001),
			service.StudioBridgeAppLuoyeAI,
			cmd.ChargeKey,
			"",
			cmd.UserID,
			cmd.Amount,
			0.3,
			"reserved",
			cmd.Fingerprint(),
			"",
			cmd.TaskID,
			cmd.Mode,
			cmd.Model,
			"",
			"",
			9.2,
			nil,
			reserveCreatedAt,
		))
	mock.ExpectQuery(`WITH existing_key AS`).
		WithArgs(cmd.UserID, sqlmock.AnyArg(), service.DefaultAPIKeyName, "image").
		WillReturnRows(sqlmock.NewRows([]string{"id", "group_id", "account_id"}).AddRow(int64(77), int64(9), int64(88)))
	mock.ExpectQuery(`WITH inserted AS \([\s\S]*INSERT INTO usage_logs[\s\S]*duration_ms[\s\S]*\$16::timestamptz[\s\S]*INSERT INTO user_usage_daily_stats`).
		WithArgs(
			cmd.UserID,
			int64(77),
			int64(88),
			"studio:"+cmd.TaskID,
			"gpt-image-2",
			int64(9),
			0.5,
			service.BillingTypeBalance,
			int16(service.RequestTypeSync),
			1,
			"1K",
			"default",
			string(service.BillingModeImage),
			"image",
			"/studio-bridge/generate",
			exactTimeArg{want: reserveCreatedAt},
		).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectExec(`UPDATE studio_bridge_charges[\s\S]*SET status = 'committed'`).
		WithArgs(int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	result, err := repo.CommitStudioBridgeCharge(context.Background(), cmd)
	require.NoError(t, err)
	require.True(t, result.Applied)
	require.Equal(t, "committed", result.Status)
	require.NoError(t, mock.ExpectationsWereMet())
}
