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
	mock.ExpectQuery(`WITH request_context AS`).
		WithArgs(int64(42), sqlmock.AnyArg(), service.DefaultAPIKeyName, "text", "").
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
	mock.ExpectQuery(`WITH request_context AS`).
		WithArgs(int64(42), sqlmock.AnyArg(), service.DefaultAPIKeyName, "image", "gpt-image-2").
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

func TestStudioBridgeRepositoryResolveChargeUsageRefsUsesOfficialImageModel(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	repo := &studioBridgeRepository{db: db}
	mock.ExpectQuery(`WITH request_context AS[\s\S]*model_mapping[\s\S]*api\.apimart\.ai`).
		WithArgs(int64(42), sqlmock.AnyArg(), service.DefaultAPIKeyName, "image", "gpt-image-2-official").
		WillReturnRows(sqlmock.NewRows([]string{"id", "group_id", "account_id"}).AddRow(int64(77), int64(12), int64(130)))

	refs, err := repo.resolveChargeUsageRefs(context.Background(), db, service.StudioBridgeChargeCommand{
		UserID: 42,
		Mode:   "generate",
		Model:  "gpt-image-2-official",
	})
	require.NoError(t, err)
	require.Equal(t, int64(77), refs.apiKeyID)
	require.Equal(t, int64(130), refs.accountID)
	require.True(t, refs.groupID.Valid)
	require.Equal(t, int64(12), refs.groupID.Int64)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStudioBridgeRepositoryGetUserSummaryUsesActualModelFields(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	repo := &studioBridgeRepository{db: db}
	createdAt := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`SELECT id, email, username, balance\s+FROM users`).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "username", "balance"}).AddRow(int64(42), "alice@example.test", "alice", 12.5))
	mock.ExpectQuery(`SELECT\s+request_id,\s+COALESCE\(NULLIF\(TRIM\(upstream_model\), ''\), NULLIF\(TRIM\(model\), ''\), NULLIF\(TRIM\(requested_model\), ''\), ''\),.*FROM usage_logs`).
		WithArgs(int64(42), 20).
		WillReturnRows(sqlmock.NewRows([]string{"request_id", "model", "requested_model", "upstream_model", "actual_model", "actual_cost", "created_at", "duration_ms", "billing_mode", "media_type", "inbound_endpoint"}).
			AddRow("req-1", "gpt-5.5", "auto", "gpt-5.5", "gpt-5.5", 0.25, createdAt, int64(4200), "token", "", "/studio-bridge/chat").
			AddRow("studio:task-image-1", "gpt-image-2", "auto", "", "gpt-image-2", 0.5, createdAt, int64(149001), "image", "image", "/studio-bridge/generate").
			AddRow("req-3", "auto", "auto", "", "auto", 0.0, createdAt, int64(0), "token", "", "/studio-bridge/chat"))
	mock.ExpectQuery(`SELECT id, amount, status, created_at, paid_at\s+FROM payment_orders`).
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "amount", "status", "created_at", "paid_at"}))

	summary, err := repo.GetUserSummary(context.Background(), 42, "https://example.test/recharge", 20)
	require.NoError(t, err)
	require.Len(t, summary.Usage, 3)
	require.Equal(t, "gpt-5.5", summary.Usage[0].Model)
	require.Equal(t, "auto", summary.Usage[0].RequestedModel)
	require.Equal(t, "gpt-5.5", summary.Usage[0].UpstreamModel)
	require.Equal(t, "gpt-5.5", summary.Usage[0].ActualModel)
	require.Equal(t, "Text", summary.Usage[0].Type)
	require.Equal(t, "req-1", summary.Usage[0].TaskID)
	require.Equal(t, int64(4200), summary.Usage[0].DurationMs)
	require.Equal(t, int64(5), summary.Usage[0].DurationSeconds)
	require.Equal(t, "success", summary.Usage[0].Status)
	require.Equal(t, "gpt-image-2", summary.Usage[1].Model)
	require.Equal(t, "auto", summary.Usage[1].RequestedModel)
	require.Empty(t, summary.Usage[1].UpstreamModel)
	require.Equal(t, "gpt-image-2", summary.Usage[1].ActualModel)
	require.Equal(t, "Image", summary.Usage[1].Type)
	require.Equal(t, "task-image-1", summary.Usage[1].TaskID)
	require.Equal(t, int64(150), summary.Usage[1].DurationSeconds)
	require.Equal(t, "gpt-5.5", summary.Usage[2].Model)
	require.Equal(t, "auto", summary.Usage[2].RequestedModel)
	require.Equal(t, "gpt-5.5", summary.Usage[2].ActualModel)
	require.Equal(t, "failed", summary.Usage[2].Status)
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
		Model:     "auto",
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
	mock.ExpectQuery(`WITH request_context AS`).
		WithArgs(cmd.UserID, sqlmock.AnyArg(), service.DefaultAPIKeyName, "image", "gpt-image-2").
		WillReturnRows(sqlmock.NewRows([]string{"id", "group_id", "account_id"}).AddRow(int64(77), int64(9), int64(88)))
	mock.ExpectQuery(`WITH inserted AS \([\s\S]*INSERT INTO usage_logs[\s\S]*duration_ms[\s\S]*\$17::timestamptz[\s\S]*INSERT INTO user_usage_daily_stats`).
		WithArgs(
			cmd.UserID,
			int64(77),
			int64(88),
			"studio:"+cmd.TaskID,
			"gpt-image-2",
			"auto",
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

func TestStudioBridgeResolvedUsageModelMapsAutoByRouteKind(t *testing.T) {
	require.Equal(t, "gpt-5.5", studioBridgeResolvedUsageModel("auto", "chat", "", "", ""))
	require.Equal(t, "gpt-image-2", studioBridgeResolvedUsageModel("auto", "generate", "", "", ""))
	require.Equal(t, "gpt-image-2", studioBridgeResolvedUsageModel("auto", "", string(service.BillingModeImage), "image", "/studio-bridge/generate"))
	require.Equal(t, "gpt-5.4", studioBridgeResolvedUsageModel("gpt-5.4", "chat", "", "", ""))
}

func TestStudioBridgeUsagePresentationFields(t *testing.T) {
	require.Equal(t, "Image", studioBridgeUsageTypeLabel("", string(service.BillingModeImage), "image", "/studio-bridge/generate", "auto"))
	require.Equal(t, "Video", studioBridgeUsageTypeLabel("video", "", "", "", "seedance"))
	require.Equal(t, "Text", studioBridgeUsageTypeLabel("chat", "", "", "", "auto"))
	require.Equal(t, "task-123", studioBridgeUsageTaskID("studio:task-123"))
	require.Equal(t, "req-123", studioBridgeUsageTaskID("req-123"))
	require.Equal(t, int64(0), studioBridgeUsageDurationSeconds(0))
	require.Equal(t, int64(1), studioBridgeUsageDurationSeconds(1))
	require.Equal(t, int64(2), studioBridgeUsageDurationSeconds(1001))
	require.Equal(t, "success", studioBridgeUsageStatus(0.01))
	require.Equal(t, "failed", studioBridgeUsageStatus(0))
}
