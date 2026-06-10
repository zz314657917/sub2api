package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type studioBridgeRepository struct {
	db *sql.DB
}

func NewStudioBridgeRepository(db *sql.DB) service.StudioBridgeRepository {
	return &studioBridgeRepository{db: db}
}

func (r *studioBridgeRepository) GetUserSummary(ctx context.Context, userID int64, rechargeURL string, usageLimit int) (*service.StudioBridgeUserSummary, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("studio bridge repository db is nil")
	}
	var summary service.StudioBridgeUserSummary
	err := r.db.QueryRowContext(ctx, `
		SELECT id, email, username, balance
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(&summary.UserID, &summary.Email, &summary.Username, &summary.Balance)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	summary.RechargeURL = rechargeURL
	if usageLimit > 0 {
		usageRows, err := r.db.QueryContext(ctx, `
			SELECT request_id, COALESCE(NULLIF(requested_model, ''), model), actual_cost, created_at
			FROM usage_logs
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`, userID, usageLimit)
		if err != nil {
			return nil, err
		}
		defer usageRows.Close()
		for usageRows.Next() {
			var item service.StudioBridgeUsageSummary
			if err := usageRows.Scan(&item.RequestID, &item.Model, &item.ActualCost, &item.CreatedAt); err != nil {
				return nil, err
			}
			summary.Usage = append(summary.Usage, item)
		}
		if err := usageRows.Err(); err != nil {
			return nil, err
		}
		orderRows, err := r.db.QueryContext(ctx, `
			SELECT id, amount, status, created_at, paid_at
			FROM payment_orders
			WHERE user_id = $1 AND order_type = 'balance'
			ORDER BY created_at DESC
			LIMIT 10
		`, userID)
		if err != nil {
			return nil, err
		}
		defer orderRows.Close()
		for orderRows.Next() {
			var item service.StudioBridgeRechargeOrder
			if err := orderRows.Scan(&item.ID, &item.Amount, &item.Status, &item.CreatedAt, &item.PaidAt); err != nil {
				return nil, err
			}
			summary.Orders = append(summary.Orders, item)
		}
		if err := orderRows.Err(); err != nil {
			return nil, err
		}
	}
	if summary.Usage == nil {
		summary.Usage = []service.StudioBridgeUsageSummary{}
	}
	if summary.Orders == nil {
		summary.Orders = []service.StudioBridgeRechargeOrder{}
	}
	return &summary, nil
}

func (r *studioBridgeRepository) ReserveStudioBridgeCharge(ctx context.Context, cmd service.StudioBridgeChargeCommand) (_ *service.StudioBridgeChargeResult, err error) {
	if r == nil || r.db == nil {
		return nil, errors.New("studio bridge repository db is nil")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollbackSQLTx(&tx)

	chargeID, inserted, err := claimStudioBridgeReserveCharge(ctx, tx, cmd)
	if err != nil {
		return nil, err
	}
	if !inserted {
		existing, ok, err := lockStudioBridgeCharge(ctx, tx, cmd.AppID, cmd.ChargeKey)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, service.ErrStudioBridgeChargeKeyEmpty
		}
		if existing.Fingerprint != cmd.Fingerprint() {
			return nil, service.ErrStudioBridgeConflict
		}
		return finishStudioBridgeTx(tx, &service.StudioBridgeChargeResult{
			ChargeKey:    cmd.ChargeKey,
			Status:       existing.Status,
			Applied:      false,
			Amount:       existing.Amount,
			BalanceAfter: existing.BalanceAfter,
		})
	}

	balanceAfter, err := reserveStudioBridgeUserBalance(ctx, tx, cmd.UserID, cmd.Amount)
	if err != nil {
		return nil, err
	}
	if err := updateStudioBridgeChargeReservedBalance(ctx, tx, chargeID, balanceAfter); err != nil {
		return nil, err
	}
	return finishStudioBridgeTx(tx, &service.StudioBridgeChargeResult{
		ChargeKey:    cmd.ChargeKey,
		Status:       "reserved",
		Applied:      true,
		Amount:       cmd.Amount,
		BalanceAfter: balanceAfter,
	})
}

func (r *studioBridgeRepository) CommitStudioBridgeCharge(ctx context.Context, cmd service.StudioBridgeChargeCommand) (_ *service.StudioBridgeChargeResult, err error) {
	if r == nil || r.db == nil {
		return nil, errors.New("studio bridge repository db is nil")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollbackSQLTx(&tx)

	charge, ok, err := lockStudioBridgeCharge(ctx, tx, cmd.AppID, cmd.ChargeKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, service.ErrStudioBridgeChargeKeyEmpty
	}
	if charge.Fingerprint != cmd.Fingerprint() {
		return nil, service.ErrStudioBridgeConflict
	}
	if charge.Status == "committed" || charge.Status == "refunded" {
		return finishStudioBridgeTx(tx, &service.StudioBridgeChargeResult{
			ChargeKey:    cmd.ChargeKey,
			Status:       charge.Status,
			Applied:      false,
			Amount:       charge.Amount,
			BalanceAfter: charge.BalanceAfter,
		})
	}

	usageAmount := charge.Amount - charge.RefundedAmount
	if usageAmount > 0 && charge.UsageLoggedAt == nil {
		usageCmd := charge.command(cmd)
		if err := r.createChargeUsageLog(ctx, tx, usageCmd, usageAmount); err != nil {
			return nil, err
		}
	}
	if err := updateStudioBridgeChargeCommitted(ctx, tx, charge.ID); err != nil {
		return nil, err
	}
	return finishStudioBridgeTx(tx, &service.StudioBridgeChargeResult{
		ChargeKey:    cmd.ChargeKey,
		Status:       "committed",
		Applied:      true,
		Amount:       charge.Amount,
		BalanceAfter: charge.BalanceAfter,
	})
}

func (r *studioBridgeRepository) RefundStudioBridgeCharge(ctx context.Context, cmd service.StudioBridgeChargeCommand) (_ *service.StudioBridgeChargeResult, err error) {
	if r == nil || r.db == nil {
		return nil, errors.New("studio bridge repository db is nil")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer rollbackSQLTx(&tx)

	if strings.TrimSpace(cmd.RefundForChargeKey) == "" || strings.TrimSpace(cmd.RefundForChargeKey) == cmd.ChargeKey {
		return r.refundOriginalStudioBridgeCharge(ctx, tx, cmd)
	}
	return r.refundStudioBridgeChargeWithRefundKey(ctx, tx, cmd)
}

func (r *studioBridgeRepository) refundStudioBridgeChargeWithRefundKey(ctx context.Context, tx *sql.Tx, cmd service.StudioBridgeChargeCommand) (*service.StudioBridgeChargeResult, error) {
	original, ok, err := lockStudioBridgeCharge(ctx, tx, cmd.AppID, cmd.RefundForChargeKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, service.ErrStudioBridgeChargeKeyEmpty
	}

	refund, ok, err := lockStudioBridgeCharge(ctx, tx, cmd.AppID, cmd.ChargeKey)
	if err != nil {
		return nil, err
	}
	if ok {
		if refund.Fingerprint != cmd.Fingerprint() {
			return nil, service.ErrStudioBridgeConflict
		}
		return finishStudioBridgeTx(tx, &service.StudioBridgeChargeResult{
			ChargeKey:    cmd.ChargeKey,
			Status:       refund.Status,
			Applied:      false,
			Amount:       refund.Amount,
			BalanceAfter: refund.BalanceAfter,
		})
	}
	if original.UserID != cmd.UserID || original.Status != "reserved" || original.RefundedAmount+cmd.Amount > original.Amount+1e-9 {
		return nil, service.ErrStudioBridgeConflict
	}

	balanceAfter, err := refundStudioBridgeUserBalance(ctx, tx, cmd.UserID, cmd.Amount)
	if err != nil {
		return nil, err
	}
	if err := insertStudioBridgeCharge(ctx, tx, cmd, studioBridgeChargeInsert{
		Status:       "refunded",
		BalanceAfter: balanceAfter,
	}); err != nil {
		return nil, err
	}
	nextRefunded := original.RefundedAmount + cmd.Amount
	nextStatus := original.Status
	if nextRefunded+1e-9 >= original.Amount {
		nextStatus = "refunded"
	}
	if err := updateStudioBridgeOriginalRefund(ctx, tx, original.ID, nextRefunded, nextStatus, balanceAfter); err != nil {
		return nil, err
	}
	return finishStudioBridgeTx(tx, &service.StudioBridgeChargeResult{
		ChargeKey:    cmd.ChargeKey,
		Status:       "refunded",
		Applied:      true,
		Amount:       cmd.Amount,
		BalanceAfter: balanceAfter,
	})
}

func (r *studioBridgeRepository) refundOriginalStudioBridgeCharge(ctx context.Context, tx *sql.Tx, cmd service.StudioBridgeChargeCommand) (*service.StudioBridgeChargeResult, error) {
	charge, ok, err := lockStudioBridgeCharge(ctx, tx, cmd.AppID, cmd.ChargeKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, service.ErrStudioBridgeChargeKeyEmpty
	}
	if charge.UserID != cmd.UserID || charge.Fingerprint != cmd.Fingerprint() {
		return nil, service.ErrStudioBridgeConflict
	}
	if charge.Status == "refunded" {
		return finishStudioBridgeTx(tx, &service.StudioBridgeChargeResult{
			ChargeKey:    cmd.ChargeKey,
			Status:       charge.Status,
			Applied:      false,
			Amount:       charge.Amount,
			BalanceAfter: charge.BalanceAfter,
		})
	}
	if charge.Status != "reserved" {
		return nil, service.ErrStudioBridgeConflict
	}
	if cmd.Amount+1e-9 < charge.Amount-charge.RefundedAmount {
		return nil, service.ErrStudioBridgeConflict
	}
	refundAmount := charge.Amount - charge.RefundedAmount
	if refundAmount <= 0 {
		return finishStudioBridgeTx(tx, &service.StudioBridgeChargeResult{
			ChargeKey:    cmd.ChargeKey,
			Status:       charge.Status,
			Applied:      false,
			Amount:       charge.Amount,
			BalanceAfter: charge.BalanceAfter,
		})
	}
	balanceAfter, err := refundStudioBridgeUserBalance(ctx, tx, cmd.UserID, refundAmount)
	if err != nil {
		return nil, err
	}
	if err := updateStudioBridgeOriginalRefund(ctx, tx, charge.ID, charge.Amount, "refunded", balanceAfter); err != nil {
		return nil, err
	}
	return finishStudioBridgeTx(tx, &service.StudioBridgeChargeResult{
		ChargeKey:    cmd.ChargeKey,
		Status:       "refunded",
		Applied:      true,
		Amount:       refundAmount,
		BalanceAfter: balanceAfter,
	})
}

func rollbackSQLTx(tx **sql.Tx) {
	if tx != nil && *tx != nil {
		_ = (*tx).Rollback()
	}
}

func finishStudioBridgeTx[T any](tx *sql.Tx, result T) (T, error) {
	if err := tx.Commit(); err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}

type studioBridgeChargeRecord struct {
	ID                 int64
	AppID              string
	ChargeKey          string
	RefundForChargeKey string
	UserID             int64
	Amount             float64
	RefundedAmount     float64
	Status             string
	Fingerprint        string
	Reason             string
	TaskID             string
	Mode               string
	Model              string
	ActorUserID        string
	TeamID             string
	BalanceAfter       float64
	UsageLoggedAt      *time.Time
}

type studioBridgeChargeInsert struct {
	Status       string
	BalanceAfter float64
}

func claimStudioBridgeReserveCharge(ctx context.Context, tx *sql.Tx, cmd service.StudioBridgeChargeCommand) (int64, bool, error) {
	var id int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO studio_bridge_charges (
			app_id,
			charge_key,
			user_id,
			amount,
			status,
			fingerprint,
			reason,
			task_id,
			mode,
			model,
			actor_user_id,
			team_id,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, 'reserved', $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		ON CONFLICT (app_id, charge_key) DO NOTHING
		RETURNING id
	`,
		cmd.AppID,
		cmd.ChargeKey,
		cmd.UserID,
		cmd.Amount,
		cmd.Fingerprint(),
		studioBridgeNullableString(cmd.Reason),
		studioBridgeNullableString(cmd.TaskID),
		studioBridgeNullableString(cmd.Mode),
		studioBridgeNullableString(cmd.Model),
		studioBridgeNullableString(cmd.ActorUserID),
		studioBridgeNullableString(cmd.TeamID),
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

func insertStudioBridgeCharge(ctx context.Context, tx *sql.Tx, cmd service.StudioBridgeChargeCommand, values studioBridgeChargeInsert) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO studio_bridge_charges (
			app_id,
			charge_key,
			refund_for_charge_key,
			user_id,
			amount,
			status,
			fingerprint,
			reason,
			task_id,
			mode,
			model,
			actor_user_id,
			team_id,
			balance_after,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
	`,
		cmd.AppID,
		cmd.ChargeKey,
		studioBridgeNullableString(cmd.RefundForChargeKey),
		cmd.UserID,
		cmd.Amount,
		values.Status,
		cmd.Fingerprint(),
		studioBridgeNullableString(cmd.Reason),
		studioBridgeNullableString(cmd.TaskID),
		studioBridgeNullableString(cmd.Mode),
		studioBridgeNullableString(cmd.Model),
		studioBridgeNullableString(cmd.ActorUserID),
		studioBridgeNullableString(cmd.TeamID),
		values.BalanceAfter,
	)
	return err
}

func lockStudioBridgeCharge(ctx context.Context, tx *sql.Tx, appID, chargeKey string) (*studioBridgeChargeRecord, bool, error) {
	row := tx.QueryRowContext(ctx, `
		SELECT
			id,
			app_id,
			charge_key,
			COALESCE(refund_for_charge_key, ''),
			user_id,
			amount::double precision,
			refunded_amount::double precision,
			status,
			fingerprint,
			COALESCE(reason, ''),
			COALESCE(task_id, ''),
			COALESCE(mode, ''),
			COALESCE(model, ''),
			COALESCE(actor_user_id, ''),
			COALESCE(team_id, ''),
			COALESCE(balance_after, 0)::double precision,
			usage_logged_at
		FROM studio_bridge_charges
		WHERE app_id = $1 AND charge_key = $2
		FOR UPDATE
	`, appID, chargeKey)
	charge, err := scanStudioBridgeCharge(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return charge, true, nil
}

func scanStudioBridgeCharge(row *sql.Row) (*studioBridgeChargeRecord, error) {
	var charge studioBridgeChargeRecord
	var usageLoggedAt sql.NullTime
	err := row.Scan(
		&charge.ID,
		&charge.AppID,
		&charge.ChargeKey,
		&charge.RefundForChargeKey,
		&charge.UserID,
		&charge.Amount,
		&charge.RefundedAmount,
		&charge.Status,
		&charge.Fingerprint,
		&charge.Reason,
		&charge.TaskID,
		&charge.Mode,
		&charge.Model,
		&charge.ActorUserID,
		&charge.TeamID,
		&charge.BalanceAfter,
		&usageLoggedAt,
	)
	if err != nil {
		return nil, err
	}
	if usageLoggedAt.Valid {
		charge.UsageLoggedAt = &usageLoggedAt.Time
	}
	return &charge, nil
}

func reserveStudioBridgeUserBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64) (float64, error) {
	var balance float64
	err := tx.QueryRowContext(ctx, `
		UPDATE users
		SET balance = balance - $1,
			updated_at = NOW()
		WHERE id = $2
			AND deleted_at IS NULL
			AND balance >= $1
		RETURNING balance
	`, amount, userID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		var exists bool
		if e := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)`, userID).Scan(&exists); e != nil {
			return 0, e
		}
		if exists {
			return 0, service.ErrStudioBridgeInsufficient
		}
		return 0, service.ErrUserNotFound
	}
	return balance, err
}

func refundStudioBridgeUserBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64) (float64, error) {
	var balance float64
	err := tx.QueryRowContext(ctx, `
		UPDATE users
		SET balance = balance + $1,
			updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING balance
	`, amount, userID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, service.ErrUserNotFound
	}
	return balance, err
}

func updateStudioBridgeChargeReservedBalance(ctx context.Context, tx *sql.Tx, id int64, balanceAfter float64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE studio_bridge_charges
		SET balance_after = $2,
			updated_at = NOW()
		WHERE id = $1
	`, id, balanceAfter)
	return err
}

func updateStudioBridgeOriginalRefund(ctx context.Context, tx *sql.Tx, id int64, refundedAmount float64, status string, balanceAfter float64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE studio_bridge_charges
		SET refunded_amount = $2,
			status = $3,
			balance_after = $4,
			updated_at = NOW()
		WHERE id = $1
	`, id, refundedAmount, status, balanceAfter)
	return err
}

func updateStudioBridgeChargeCommitted(ctx context.Context, tx *sql.Tx, id int64) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE studio_bridge_charges
		SET status = 'committed',
			usage_logged_at = COALESCE(usage_logged_at, NOW()),
			updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func (c studioBridgeChargeRecord) command(fallback service.StudioBridgeChargeCommand) service.StudioBridgeChargeCommand {
	fallback.AppID = c.AppID
	fallback.UserID = c.UserID
	fallback.ChargeKey = c.ChargeKey
	fallback.RefundForChargeKey = c.RefundForChargeKey
	fallback.Amount = c.Amount
	fallback.Reason = firstNonEmptyString(c.Reason, fallback.Reason)
	fallback.TaskID = firstNonEmptyString(c.TaskID, fallback.TaskID)
	fallback.Mode = firstNonEmptyString(c.Mode, fallback.Mode)
	fallback.Model = firstNonEmptyString(c.Model, fallback.Model)
	fallback.ActorUserID = firstNonEmptyString(c.ActorUserID, fallback.ActorUserID)
	fallback.TeamID = firstNonEmptyString(c.TeamID, fallback.TeamID)
	return fallback
}

func studioBridgeNullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

type studioBridgeSQLExecutor interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (r *studioBridgeRepository) createChargeUsageLog(ctx context.Context, exec studioBridgeSQLExecutor, cmd service.StudioBridgeChargeCommand, amount float64) error {
	if cmd.UserID <= 0 || amount <= 0 {
		return nil
	}
	refs, err := r.resolveChargeUsageRefs(ctx, exec, cmd)
	if err != nil {
		return err
	}
	requestID := studioBridgeUsageRequestID(cmd)
	model := strings.TrimSpace(cmd.Model)
	if model == "" {
		model = "studio-bridge"
	}
	billingMode, mediaType, imageCount := studioBridgeUsageBillingFields(cmd.Mode)
	var groupID any
	if refs.groupID.Valid {
		groupID = refs.groupID.Int64
	}
	var mediaTypeArg any
	if mediaType != "" {
		mediaTypeArg = mediaType
	}
	var imageSizeArg any
	var imageSizeSourceArg any
	if imageCount > 0 {
		imageSizeArg = "1K"
		imageSizeSourceArg = "default"
	}
	var inserted bool
	err = exec.QueryRowContext(ctx, `
		WITH inserted AS (
			INSERT INTO usage_logs (
				user_id,
				api_key_id,
				account_id,
				request_id,
				model,
				requested_model,
				group_id,
				input_tokens,
				output_tokens,
				cache_creation_tokens,
				cache_read_tokens,
				cache_creation_5m_tokens,
				cache_creation_1h_tokens,
				image_output_tokens,
				image_output_cost,
				input_cost,
				output_cost,
				cache_creation_cost,
				cache_read_cost,
				total_cost,
				actual_cost,
				rate_multiplier,
				billing_type,
				request_type,
				stream,
				openai_ws_mode,
				image_count,
				image_size,
				image_size_source,
				billing_mode,
				media_type,
				inbound_endpoint,
				created_at
			) VALUES (
				$1, $2, $3, $4, $5, $5, $6,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, $7, $7, 1,
				$8, $9, FALSE, FALSE, $10, $11, $12, $13, $14, $15, NOW()
			)
			ON CONFLICT (request_id, api_key_id) DO NOTHING
			RETURNING user_id, created_at, actual_cost
		),
		daily_stats AS (
			INSERT INTO user_usage_daily_stats (
				user_id,
				usage_date,
				requests,
				tokens,
				actual_cost,
				night_requests,
				updated_at
			)
			SELECT
				user_id,
				(created_at AT TIME ZONE 'Asia/Shanghai')::date,
				1,
				0,
				actual_cost,
				CASE
					WHEN EXTRACT(HOUR FROM created_at AT TIME ZONE 'Asia/Shanghai') >= 0
						AND EXTRACT(HOUR FROM created_at AT TIME ZONE 'Asia/Shanghai') < 6 THEN 1
					ELSE 0
				END,
				NOW()
			FROM inserted
			ON CONFLICT (user_id, usage_date) DO UPDATE SET
				requests = user_usage_daily_stats.requests + EXCLUDED.requests,
				tokens = user_usage_daily_stats.tokens + EXCLUDED.tokens,
				actual_cost = user_usage_daily_stats.actual_cost + EXCLUDED.actual_cost,
				night_requests = user_usage_daily_stats.night_requests + EXCLUDED.night_requests,
				updated_at = NOW()
			RETURNING 1
		)
		SELECT EXISTS(SELECT 1 FROM inserted)
	`, cmd.UserID, refs.apiKeyID, refs.accountID, requestID, model, groupID, amount, service.BillingTypeBalance, int16(service.RequestTypeSync), imageCount, imageSizeArg, imageSizeSourceArg, billingMode, mediaTypeArg, studioBridgeUsageInboundEndpoint(cmd.Mode)).Scan(&inserted)
	return err
}

type studioBridgeChargeUsageRefs struct {
	apiKeyID  int64
	accountID int64
	groupID   sql.NullInt64
}

func (r *studioBridgeRepository) resolveChargeUsageRefs(ctx context.Context, exec studioBridgeSQLExecutor, cmd service.StudioBridgeChargeCommand) (studioBridgeChargeUsageRefs, error) {
	return r.createDefaultChargeUsageRefs(ctx, exec, cmd.UserID, studioBridgeUsageRouteKind(cmd))
}

func (r *studioBridgeRepository) createDefaultChargeUsageRefs(ctx context.Context, exec studioBridgeSQLExecutor, userID int64, routeKind string) (studioBridgeChargeUsageRefs, error) {
	var refs studioBridgeChargeUsageRefs
	err := exec.QueryRowContext(ctx, `
		WITH existing_key AS (
			SELECT id, group_id, COALESCE(multi_group_routes, '[]'::jsonb) AS multi_group_routes
			FROM api_keys
			WHERE user_id = $1
				AND deleted_at IS NULL
			ORDER BY id ASC
			LIMIT 1
		),
		key_exists AS (
			SELECT id
			FROM api_keys
			WHERE user_id = $1
				AND deleted_at IS NULL
			LIMIT 1
		),
		global_account_ref AS (
			SELECT id
			FROM accounts
			WHERE status = 'active' AND deleted_at IS NULL
			ORDER BY priority ASC, id ASC
			LIMIT 1
		),
		inserted_key AS (
			INSERT INTO api_keys (
				user_id,
				key,
				name,
				status,
				account_pool_strategy,
				multi_group_routes,
				created_at,
				updated_at
			)
			SELECT
				$1,
				$2,
				$3,
				'active',
				'shared_only',
				'[]'::jsonb,
				NOW(),
				NOW()
			WHERE NOT EXISTS (SELECT 1 FROM key_exists)
				AND EXISTS (SELECT 1 FROM global_account_ref)
			RETURNING id, group_id, '[]'::jsonb AS multi_group_routes
		),
		selected_key AS (
			SELECT id, group_id, multi_group_routes FROM existing_key
			UNION ALL
			SELECT id, group_id, multi_group_routes FROM inserted_key
			LIMIT 1
		),
		route_group AS (
			SELECT
				(route ->> 'group_id')::bigint AS group_id,
				CASE
					WHEN COALESCE(route ->> 'priority', '') ~ '^[0-9]+$' THEN (route ->> 'priority')::int
					ELSE 1
				END AS route_priority,
				route_ord
			FROM selected_key
			CROSS JOIN LATERAL jsonb_array_elements(selected_key.multi_group_routes) WITH ORDINALITY AS routes(route, route_ord)
			WHERE LOWER(COALESCE(route ->> 'enabled', 'false')) = 'true'
				AND COALESCE(route ->> 'group_id', '') ~ '^[0-9]+$'
				AND (
					($4 = 'image' AND LOWER(COALESCE(route ->> 'image_only', 'false')) = 'true')
					OR ($4 = 'text' AND LOWER(COALESCE(route ->> 'text_only', 'false')) = 'true')
					OR ($4 = 'video' AND jsonb_typeof(route -> 'model_patterns') = 'array' AND jsonb_array_length(route -> 'model_patterns') > 0)
				)
			ORDER BY route_priority ASC, route_ord ASC
			LIMIT 1
		),
		target_group AS (
			SELECT COALESCE((SELECT group_id FROM route_group), selected_key.group_id) AS group_id
			FROM selected_key
		),
		group_account_ref AS (
			SELECT accounts.id
			FROM target_group
			JOIN account_groups ON account_groups.group_id = target_group.group_id
			JOIN accounts ON accounts.id = account_groups.account_id
			WHERE target_group.group_id IS NOT NULL
				AND accounts.status = 'active'
				AND accounts.deleted_at IS NULL
			ORDER BY account_groups.priority ASC, accounts.priority ASC, accounts.id ASC
			LIMIT 1
		),
		account_ref AS (
			SELECT id FROM group_account_ref
			UNION ALL
			SELECT id FROM global_account_ref
			WHERE NOT EXISTS (SELECT 1 FROM group_account_ref)
			LIMIT 1
		)
		SELECT selected_key.id, target_group.group_id, account_ref.id
		FROM selected_key
		CROSS JOIN target_group
		CROSS JOIN account_ref
	`, userID, "sk-"+uuid.NewString(), service.DefaultAPIKeyName, routeKind).Scan(&refs.apiKeyID, &refs.groupID, &refs.accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return refs, errors.New("studio bridge usage log requires an active account")
	}
	return refs, err
}

func studioBridgeUsageRouteKind(cmd service.StudioBridgeChargeCommand) string {
	mode := strings.ToLower(strings.TrimSpace(cmd.Mode))
	switch mode {
	case "generate", "edit", "image":
		return "image"
	case "video":
		return "video"
	case "chat", "text":
		return "text"
	}
	model := strings.ToLower(strings.TrimSpace(cmd.Model))
	if strings.Contains(model, "image") || strings.HasPrefix(model, "gpt-image") {
		return "image"
	}
	if strings.Contains(model, "video") || strings.Contains(model, "seedance") {
		return "video"
	}
	return "text"
}

func studioBridgeUsageRequestID(cmd service.StudioBridgeChargeCommand) string {
	taskID := strings.TrimSpace(cmd.TaskID)
	if taskID != "" && len("studio:"+taskID) <= 64 {
		return "studio:" + taskID
	}
	source := strings.TrimSpace(cmd.ChargeKey)
	if source == "" {
		source = strings.TrimSpace(cmd.Reason)
	}
	sum := sha256.Sum256([]byte(source))
	return "studio:" + hex.EncodeToString(sum[:])[:24]
}

func studioBridgeUsageBillingFields(mode string) (billingMode string, mediaType string, imageCount int) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "generate", "edit", "image":
		return string(service.BillingModeImage), "image", 1
	case "video":
		return string(service.BillingModePerRequest), "video", 0
	default:
		return string(service.BillingModeToken), "", 0
	}
}

func studioBridgeUsageInboundEndpoint(mode string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = "unknown"
	}
	return "/studio-bridge/" + mode
}
