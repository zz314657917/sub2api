package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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
		SELECT
			u.id,
			u.email,
			u.username,
			u.balance::double precision,
			COALESCE(voucher.available_amount, 0)::double precision
		FROM users u
		LEFT JOIN LATERAL (
			SELECT SUM(v.remaining_amount) AS available_amount
			FROM welfare_vouchers v
			WHERE v.user_id = u.id
				AND v.status = 'active'
				AND v.remaining_amount > 0
				AND (v.expires_at IS NULL OR v.expires_at > NOW())
		) voucher ON TRUE
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`, userID).Scan(&summary.UserID, &summary.Email, &summary.Username, &summary.Balance, &summary.VoucherBalance)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	summary.TotalAvailable = summary.Balance + summary.VoucherBalance
	summary.RechargeURL = rechargeURL
	if usageLimit > 0 {
		usageRows, err := r.db.QueryContext(ctx, `
			SELECT
				request_id,
				COALESCE(NULLIF(TRIM(upstream_model), ''), NULLIF(TRIM(model), ''), NULLIF(TRIM(requested_model), ''), ''),
				COALESCE(NULLIF(TRIM(requested_model), ''), ''),
				COALESCE(NULLIF(TRIM(upstream_model), ''), ''),
				COALESCE(NULLIF(TRIM(upstream_model), ''), NULLIF(TRIM(model), ''), NULLIF(TRIM(requested_model), ''), ''),
				actual_cost,
				created_at,
				COALESCE(duration_ms, 0)::bigint,
				COALESCE(NULLIF(TRIM(billing_mode), ''), ''),
				COALESCE(NULLIF(TRIM(media_type), ''), ''),
				COALESCE(NULLIF(TRIM(inbound_endpoint), ''), '')
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
			var billingMode, mediaType, inboundEndpoint string
			var durationMs int64
			if err := usageRows.Scan(&item.RequestID, &item.Model, &item.RequestedModel, &item.UpstreamModel, &item.ActualModel, &item.ActualCost, &item.CreatedAt, &durationMs, &billingMode, &mediaType, &inboundEndpoint); err != nil {
				return nil, err
			}
			item.Model = studioBridgeResolvedUsageModel(item.Model, "", billingMode, mediaType, inboundEndpoint)
			item.ActualModel = studioBridgeResolvedUsageModel(item.ActualModel, "", billingMode, mediaType, inboundEndpoint)
			item.Type = studioBridgeUsageTypeLabel("", billingMode, mediaType, inboundEndpoint, item.Model)
			item.TaskID = studioBridgeUsageTaskID(item.RequestID)
			item.DurationMs = durationMs
			item.DurationSeconds = studioBridgeUsageDurationSeconds(durationMs)
			item.Status = studioBridgeUsageStatus(item.ActualCost)
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
		if !studioBridgeChargeFingerprintMatches(existing, cmd) {
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

	balanceAfter, err := reserveStudioBridgeUserBalance(ctx, tx, cmd.UserID, cmd.Amount, cmd.ChargeKey)
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
	if !studioBridgeChargeFingerprintMatches(charge, cmd) {
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
		if err := r.createChargeUsageLog(ctx, tx, usageCmd, usageAmount, charge.CreatedAt); err != nil {
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
		if !studioBridgeChargeFingerprintMatches(refund, cmd) {
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
	refundAmount := studioBridgeRefundAmount(original, cmd)
	if original.UserID != cmd.UserID || original.Status != "reserved" || original.RefundedAmount+refundAmount > original.Amount+1e-9 {
		return nil, service.ErrStudioBridgeConflict
	}

	balanceAfter, err := refundStudioBridgeUserBalance(ctx, tx, cmd.UserID, refundAmount, original.ChargeKey)
	if err != nil {
		return nil, err
	}
	cmd.Amount = refundAmount
	if err := insertStudioBridgeCharge(ctx, tx, cmd, studioBridgeChargeInsert{
		Status:       "refunded",
		BalanceAfter: balanceAfter,
	}); err != nil {
		return nil, err
	}
	nextRefunded := original.RefundedAmount + refundAmount
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
		Amount:       refundAmount,
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
	if charge.UserID != cmd.UserID || !studioBridgeChargeFingerprintMatches(charge, cmd) {
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
	balanceAfter, err := refundStudioBridgeUserBalance(ctx, tx, cmd.UserID, refundAmount, charge.ChargeKey)
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
	ImageCount         int
	ImageSize          string
	ImageSizeSource    string
	ImageSizeBreakdown map[string]int
	BalanceAfter       float64
	UsageLoggedAt      *time.Time
	CreatedAt          time.Time
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
			image_count,
			image_size,
			image_size_source,
			image_size_breakdown,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, 'reserved', $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW(), NOW())
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
		cmd.ImageCount,
		studioBridgeUsageImageSize(cmd, cmd.ImageCount),
		studioBridgeUsageImageSizeSource(cmd, cmd.ImageCount),
		studioBridgeUsageImageSizeBreakdown(cmd, cmd.ImageCount),
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
			image_count,
			image_size,
			image_size_source,
			image_size_breakdown,
			balance_after,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, NOW(), NOW())
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
		cmd.ImageCount,
		studioBridgeUsageImageSize(cmd, cmd.ImageCount),
		studioBridgeUsageImageSizeSource(cmd, cmd.ImageCount),
		studioBridgeUsageImageSizeBreakdown(cmd, cmd.ImageCount),
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
			COALESCE(image_count, 0),
			COALESCE(image_size, ''),
			COALESCE(image_size_source, ''),
			image_size_breakdown::text,
			COALESCE(balance_after, 0)::double precision,
			usage_logged_at,
			created_at
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
	var imageSizeBreakdown sql.NullString
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
		&charge.ImageCount,
		&charge.ImageSize,
		&charge.ImageSizeSource,
		&imageSizeBreakdown,
		&charge.BalanceAfter,
		&usageLoggedAt,
		&charge.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if usageLoggedAt.Valid {
		charge.UsageLoggedAt = &usageLoggedAt.Time
	}
	charge.ImageSizeBreakdown = stringIntMapFromNullJSON(imageSizeBreakdown)
	return &charge, nil
}

func reserveStudioBridgeUserBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64, chargeKey string) (float64, error) {
	result, err := deductWelfareVoucherThenBalance(ctx, tx, userID, amount, welfareVoucherOperationStudioBridge, chargeKey, true)
	if errors.Is(err, service.ErrInsufficientBalance) {
		return 0, service.ErrStudioBridgeInsufficient
	}
	if err != nil {
		return 0, err
	}
	return result.BalanceAfter, nil
}

func refundStudioBridgeUserBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64, chargeKey string) (float64, error) {
	result, err := refundWelfareVoucherDeductions(ctx, tx, userID, amount, welfareVoucherOperationStudioBridge, chargeKey)
	if err != nil {
		return 0, err
	}
	return result.BalanceAfter, nil
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
	fallback.AmountUnit = firstNonEmptyString(service.StudioBridgeAmountUnitFromFingerprint(c.Fingerprint), fallback.AmountUnit)
	fallback.Reason = firstNonEmptyString(c.Reason, fallback.Reason)
	fallback.TaskID = firstNonEmptyString(c.TaskID, fallback.TaskID)
	fallback.Mode = firstNonEmptyString(c.Mode, fallback.Mode)
	fallback.Model = firstNonEmptyString(c.Model, fallback.Model)
	fallback.ActorUserID = firstNonEmptyString(c.ActorUserID, fallback.ActorUserID)
	fallback.TeamID = firstNonEmptyString(c.TeamID, fallback.TeamID)
	if c.ImageCount > 0 {
		fallback.ImageCount = c.ImageCount
	}
	fallback.ImageSize = firstNonEmptyString(c.ImageSize, fallback.ImageSize)
	fallback.ImageSizeSource = firstNonEmptyString(c.ImageSizeSource, fallback.ImageSizeSource)
	if len(c.ImageSizeBreakdown) > 0 {
		fallback.ImageSizeBreakdown = copyStudioBridgeImageSizeBreakdown(c.ImageSizeBreakdown)
	}
	return fallback
}

func studioBridgeChargeFingerprintMatches(charge *studioBridgeChargeRecord, cmd service.StudioBridgeChargeCommand) bool {
	if charge == nil {
		return false
	}
	if charge.Fingerprint == cmd.Fingerprint() {
		return true
	}
	if cmd.RawAmount() == cmd.Amount {
		return false
	}
	legacy := cmd
	legacy.Amount = cmd.RawAmount()
	legacy.AmountUnit = ""
	return charge.Fingerprint == legacy.Fingerprint()
}

func studioBridgeRefundAmount(original *studioBridgeChargeRecord, cmd service.StudioBridgeChargeCommand) float64 {
	if original == nil {
		return cmd.Amount
	}
	refundCmd := original.command(cmd)
	refundCmd.ChargeKey = cmd.ChargeKey
	refundCmd.RefundForChargeKey = cmd.RefundForChargeKey
	return service.NormalizeStudioBridgeChargeAmount(refundCmd, cmd.RawAmount())
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

func (r *studioBridgeRepository) createChargeUsageLog(ctx context.Context, exec studioBridgeSQLExecutor, cmd service.StudioBridgeChargeCommand, amount float64, chargeCreatedAt time.Time) error {
	if cmd.UserID <= 0 || amount <= 0 {
		return nil
	}
	refs, err := r.resolveChargeUsageRefs(ctx, exec, cmd)
	if err != nil {
		return err
	}
	requestID := studioBridgeUsageRequestID(cmd)
	requestedModel := strings.TrimSpace(cmd.Model)
	if requestedModel == "" {
		requestedModel = "studio-bridge"
	}
	billingMode, mediaType, imageCount := studioBridgeUsageBillingFields(cmd.Mode)
	if cmd.ImageCount > 0 {
		imageCount = cmd.ImageCount
	}
	imageSize := studioBridgeUsageImageSize(cmd, imageCount)
	imageSizeSource := studioBridgeUsageImageSizeSource(cmd, imageCount)
	imageSizeBreakdown := studioBridgeUsageImageSizeBreakdown(cmd, imageCount)
	inboundEndpoint := studioBridgeUsageInboundEndpoint(cmd.Mode)
	model := studioBridgeResolvedUsageModel(requestedModel, cmd.Mode, billingMode, mediaType, inboundEndpoint)
	var groupID any
	if refs.groupID.Valid {
		groupID = refs.groupID.Int64
	}
	var mediaTypeArg any
	if mediaType != "" {
		mediaTypeArg = mediaType
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
				duration_ms,
				image_count,
				image_size,
				image_size_source,
				image_size_breakdown,
				billing_mode,
				media_type,
				inbound_endpoint,
				created_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7,
				0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, $8, $8, 1,
				$9, $10, FALSE, FALSE,
				CASE
					WHEN $18::timestamptz IS NULL THEN NULL
					ELSE LEAST(
						2147483647,
						GREATEST(0, FLOOR(EXTRACT(EPOCH FROM (NOW() - $18::timestamptz)) * 1000))
					)::int
				END,
				$11, $12, $13, $14, $15, $16, $17, NOW()
			)
			ON CONFLICT (request_id, api_key_id) DO NOTHING
			RETURNING user_id, created_at, actual_cost
		),
		daily_stats AS (
			INSERT INTO user_usage_daily_stats (
				user_id,
				usage_date,
				requests,
				input_tokens,
				output_tokens,
				cache_creation_tokens,
				cache_read_tokens,
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
				0,
				0,
				0,
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
				input_tokens = user_usage_daily_stats.input_tokens + EXCLUDED.input_tokens,
				output_tokens = user_usage_daily_stats.output_tokens + EXCLUDED.output_tokens,
				cache_creation_tokens = user_usage_daily_stats.cache_creation_tokens + EXCLUDED.cache_creation_tokens,
				cache_read_tokens = user_usage_daily_stats.cache_read_tokens + EXCLUDED.cache_read_tokens,
				tokens = user_usage_daily_stats.tokens + EXCLUDED.tokens,
				actual_cost = user_usage_daily_stats.actual_cost + EXCLUDED.actual_cost,
				night_requests = user_usage_daily_stats.night_requests + EXCLUDED.night_requests,
				updated_at = NOW()
			RETURNING 1
		)
		SELECT EXISTS(SELECT 1 FROM inserted)
	`, cmd.UserID, refs.apiKeyID, refs.accountID, requestID, model, requestedModel, groupID, amount, service.BillingTypeBalance, int16(service.RequestTypeSync), imageCount, imageSize, imageSizeSource, imageSizeBreakdown, billingMode, mediaTypeArg, inboundEndpoint, studioBridgeDurationStartArg(chargeCreatedAt)).Scan(&inserted)
	return err
}

func studioBridgeUsageImageSize(cmd service.StudioBridgeChargeCommand, imageCount int) any {
	if imageCount <= 0 || strings.TrimSpace(cmd.ImageSize) == "" {
		return nil
	}
	return studioBridgeNormalizeUsageImageSize(cmd.ImageSize)
}

func studioBridgeUsageImageSizeSource(cmd service.StudioBridgeChargeCommand, imageCount int) any {
	if imageCount <= 0 || strings.TrimSpace(cmd.ImageSize) == "" {
		return nil
	}
	switch strings.TrimSpace(cmd.ImageSizeSource) {
	case service.ImageSizeSourceOutput:
		return service.ImageSizeSourceOutput
	case service.ImageSizeSourceInput:
		return service.ImageSizeSourceInput
	case service.ImageSizeSourceDefault:
		return service.ImageSizeSourceDefault
	case service.ImageSizeSourceLegacy:
		return service.ImageSizeSourceLegacy
	default:
		return service.ImageSizeSourceDefault
	}
}

func studioBridgeUsageImageSizeBreakdown(cmd service.StudioBridgeChargeCommand, imageCount int) any {
	if imageCount <= 0 || strings.TrimSpace(cmd.ImageSize) == "" {
		return nil
	}
	breakdown := cmd.ImageSizeBreakdown
	if len(breakdown) == 0 {
		if tier, ok := studioBridgeUsageImageSizeTier(cmd.ImageSize); ok {
			breakdown = map[string]int{tier: imageCount}
		}
	}
	if len(breakdown) == 0 {
		return nil
	}
	data, err := json.Marshal(breakdown)
	if err != nil || len(data) == 0 {
		return nil
	}
	return string(data)
}

func studioBridgeNormalizeUsageImageSize(size string) string {
	trimmed := strings.TrimSpace(size)
	if strings.EqualFold(trimmed, service.ImageBillingSizeMixed) {
		return service.ImageBillingSizeMixed
	}
	return service.NormalizeImageBillingTierOrDefault(trimmed)
}

func studioBridgeUsageImageSizeTier(size string) (string, bool) {
	if strings.EqualFold(strings.TrimSpace(size), service.ImageBillingSizeMixed) {
		return "", false
	}
	return service.ClassifyImageBillingTier(size)
}

func copyStudioBridgeImageSizeBreakdown(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func studioBridgeDurationStartArg(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

type studioBridgeChargeUsageRefs struct {
	apiKeyID  int64
	accountID int64
	groupID   sql.NullInt64
}

func (r *studioBridgeRepository) resolveChargeUsageRefs(ctx context.Context, exec studioBridgeSQLExecutor, cmd service.StudioBridgeChargeCommand) (studioBridgeChargeUsageRefs, error) {
	routeKind := studioBridgeUsageRouteKind(cmd)
	requestedModel := strings.TrimSpace(cmd.Model)
	if strings.EqualFold(requestedModel, "auto") {
		requestedModel = studioBridgeResolvedUsageModel(requestedModel, cmd.Mode, "", "", studioBridgeUsageInboundEndpoint(cmd.Mode))
	}
	return r.createDefaultChargeUsageRefs(ctx, exec, cmd.UserID, routeKind, requestedModel)
}

func (r *studioBridgeRepository) createDefaultChargeUsageRefs(ctx context.Context, exec studioBridgeSQLExecutor, userID int64, routeKind, requestedModel string) (studioBridgeChargeUsageRefs, error) {
	var refs studioBridgeChargeUsageRefs
	routeModelMatchRankSQL := studioBridgeRouteModelMatchRankSQL("route", "request_context.requested_model", "request_context.mapped_requested_model")
	accountModelMatchRankSQL := studioBridgeAccountModelMatchRankSQL("accounts", "request_context.requested_model", "request_context.mapped_requested_model")
	query := `
		WITH request_context AS (
			SELECT
				LOWER(TRIM($5::text)) AS requested_model,
				CASE
					WHEN LOWER(TRIM($5::text)) = 'gpt-image-2-official' THEN 'gpt-image-2'
					ELSE LOWER(TRIM($5::text))
				END AS mapped_requested_model
		),
		existing_key AS (
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
		global_account_candidates AS (
			SELECT
				accounts.id,
				global_group.group_id,
				accounts.priority,
				` + accountModelMatchRankSQL + ` AS model_match_rank
			FROM accounts
			CROSS JOIN request_context
			LEFT JOIN LATERAL (
				SELECT account_groups.group_id
				FROM account_groups
				WHERE account_groups.account_id = accounts.id
				ORDER BY account_groups.priority ASC, account_groups.group_id ASC
				LIMIT 1
			) global_group ON TRUE
			WHERE status = 'active' AND deleted_at IS NULL
		),
		global_account_ref AS (
			SELECT id, group_id
			FROM global_account_candidates
			ORDER BY
				model_match_rank DESC,
				priority ASC,
				id ASC
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
		route_candidates AS (
			SELECT
				(route ->> 'group_id')::bigint AS group_id,
				CASE
					WHEN COALESCE(route ->> 'priority', '') ~ '^[0-9]+$' THEN (route ->> 'priority')::int
					ELSE 1
				END AS route_priority,
				route_ord,
				` + routeModelMatchRankSQL + ` AS model_match_rank,
				(
					($4 = 'image' AND LOWER(COALESCE(route ->> 'image_only', 'false')) = 'true')
					OR ($4 = 'text' AND LOWER(COALESCE(route ->> 'text_only', 'false')) = 'true')
					OR ($4 = 'video' AND jsonb_typeof(route -> 'model_patterns') = 'array' AND jsonb_array_length(route -> 'model_patterns') > 0)
				) AS route_kind_match
			FROM selected_key
			CROSS JOIN request_context
			CROSS JOIN LATERAL jsonb_array_elements(selected_key.multi_group_routes) WITH ORDINALITY AS routes(route, route_ord)
			WHERE LOWER(COALESCE(route ->> 'enabled', 'false')) = 'true'
				AND COALESCE(route ->> 'group_id', '') ~ '^[0-9]+$'
		),
		route_group AS (
			SELECT group_id
			FROM route_candidates
			WHERE route_kind_match OR model_match_rank > 0
			ORDER BY model_match_rank DESC, route_priority ASC, route_ord ASC
			LIMIT 1
		),
		target_group AS (
			SELECT COALESCE((SELECT group_id FROM route_group), selected_key.group_id) AS group_id
			FROM selected_key
		),
		group_account_candidates AS (
			SELECT
				accounts.id,
				target_group.group_id,
				account_groups.priority AS account_group_priority,
				accounts.priority AS account_priority,
				` + accountModelMatchRankSQL + ` AS model_match_rank
			FROM target_group
			CROSS JOIN request_context
			JOIN account_groups ON account_groups.group_id = target_group.group_id
			JOIN accounts ON accounts.id = account_groups.account_id
			WHERE target_group.group_id IS NOT NULL
				AND accounts.status = 'active'
				AND accounts.deleted_at IS NULL
		),
		group_account_ref AS (
			SELECT id, group_id
			FROM group_account_candidates
			ORDER BY model_match_rank DESC, account_group_priority ASC, account_priority ASC, id ASC
			LIMIT 1
		),
		group_supported_account_ref AS (
			SELECT id, group_id
			FROM group_account_candidates
			WHERE model_match_rank > 0
			ORDER BY model_match_rank DESC, account_group_priority ASC, account_priority ASC, id ASC
			LIMIT 1
		),
		global_supported_account_ref AS (
			SELECT id, group_id
			FROM global_account_candidates
			WHERE model_match_rank > 0
			ORDER BY model_match_rank DESC, priority ASC, id ASC
			LIMIT 1
		),
		account_ref AS (
			SELECT id, group_id FROM group_supported_account_ref
			UNION ALL
			SELECT id, group_id FROM global_supported_account_ref
			WHERE NOT EXISTS (SELECT 1 FROM group_supported_account_ref)
			UNION ALL
			SELECT id, group_id FROM group_account_ref
			WHERE NOT EXISTS (SELECT 1 FROM group_supported_account_ref)
				AND NOT EXISTS (SELECT 1 FROM global_supported_account_ref)
			UNION ALL
			SELECT id, group_id FROM global_account_ref
			WHERE NOT EXISTS (SELECT 1 FROM group_supported_account_ref)
				AND NOT EXISTS (SELECT 1 FROM global_supported_account_ref)
				AND NOT EXISTS (SELECT 1 FROM group_account_ref)
			LIMIT 1
		)
		SELECT selected_key.id, COALESCE(account_ref.group_id, target_group.group_id), account_ref.id
		FROM selected_key
		CROSS JOIN target_group
		CROSS JOIN account_ref
	`
	err := exec.QueryRowContext(ctx, query, userID, "sk-"+uuid.NewString(), service.DefaultAPIKeyName, routeKind, requestedModel).Scan(&refs.apiKeyID, &refs.groupID, &refs.accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return refs, errors.New("studio bridge usage log requires an active account")
	}
	return refs, err
}

func studioBridgeRouteModelMatchRankSQL(routeExpr, requestedModelExpr, mappedRequestedModelExpr string) string {
	patternMatch := studioBridgePatternMatchSQL("pattern", requestedModelExpr, mappedRequestedModelExpr)
	return `
		CASE
			WHEN ` + requestedModelExpr + ` = '' THEN 0
			WHEN jsonb_typeof(` + routeExpr + ` -> 'model_patterns') = 'array'
				AND EXISTS (
					SELECT 1
					FROM jsonb_array_elements_text(` + routeExpr + ` -> 'model_patterns') AS patterns(pattern)
					WHERE ` + patternMatch + `
				) THEN 2
			WHEN jsonb_typeof(` + routeExpr + ` -> 'model_patterns') = 'array'
				AND jsonb_array_length(` + routeExpr + ` -> 'model_patterns') > 0 THEN -1
			ELSE 0
		END`
}

func studioBridgeAccountModelMatchRankSQL(accountExpr, requestedModelExpr, mappedRequestedModelExpr string) string {
	modelMappingExpr := `(CASE WHEN jsonb_typeof(` + accountExpr + `.credentials -> 'model_mapping') = 'object' THEN ` + accountExpr + `.credentials -> 'model_mapping' ELSE '{}'::jsonb END)`
	requestedModelMatch := studioBridgeModelMappingMatchSQL(modelMappingExpr, requestedModelExpr, requestedModelExpr)
	mappedModelMatch := studioBridgeModelMappingMatchSQL(modelMappingExpr, mappedRequestedModelExpr, mappedRequestedModelExpr)
	baseURLHostExpr := `LOWER(split_part(regexp_replace(regexp_replace(COALESCE(` + accountExpr + `.credentials ->> 'base_url', ''), '^[a-z][a-z0-9+.-]*://', ''), '/.*$', ''), ':', 1))`
	return `
		CASE
			WHEN ` + requestedModelExpr + ` = '' THEN 0
			WHEN ` + requestedModelMatch + ` THEN 3
			WHEN ` + mappedModelMatch + `
				AND ` + requestedModelExpr + ` = 'gpt-image-2-official'
				AND LOWER(` + accountExpr + `.platform) = 'openai'
				AND LOWER(` + accountExpr + `.type) = 'apikey'
				AND ` + baseURLHostExpr + ` = 'api.apimart.ai' THEN 2
			WHEN jsonb_typeof(` + modelMappingExpr + `) = 'object'
				AND ` + modelMappingExpr + ` <> '{}'::jsonb THEN -1
			ELSE 0
		END`
}

func studioBridgeModelMappingMatchSQL(modelMappingExpr, requestedModelExpr, mappedRequestedModelExpr string) string {
	patternMatch := studioBridgePatternMatchSQL("model", requestedModelExpr, mappedRequestedModelExpr)
	return `EXISTS (
		SELECT 1
		FROM jsonb_object_keys(` + modelMappingExpr + `) AS mapping_models(model)
		WHERE ` + patternMatch + `
	)`
}

func studioBridgePatternMatchSQL(patternExpr, requestedModelExpr, mappedRequestedModelExpr string) string {
	normalizedPattern := `LOWER(TRIM(` + patternExpr + `))`
	wildcardPrefix := `LEFT(` + normalizedPattern + `, GREATEST(LENGTH(` + normalizedPattern + `) - 1, 0))`
	return `(
		` + normalizedPattern + ` = ` + requestedModelExpr + `
		OR ` + normalizedPattern + ` = ` + mappedRequestedModelExpr + `
		OR (RIGHT(` + normalizedPattern + `, 1) = '*' AND LEFT(` + requestedModelExpr + `, LENGTH(` + wildcardPrefix + `)) = ` + wildcardPrefix + `)
		OR (RIGHT(` + normalizedPattern + `, 1) = '*' AND LEFT(` + mappedRequestedModelExpr + `, LENGTH(` + wildcardPrefix + `)) = ` + wildcardPrefix + `)
	)`
}

func studioBridgeUsageRouteKind(cmd service.StudioBridgeChargeCommand) string {
	return studioBridgeUsageRouteKindFromHints(cmd.Mode, "", "", "", cmd.Model)
}

func studioBridgeUsageRouteKindFromHints(mode, billingMode, mediaType, inboundEndpoint, model string) string {
	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case "generate", "edit", "image":
		return "image"
	case "video":
		return "video"
	case "chat", "text":
		return "text"
	}
	billingMode = strings.ToLower(strings.TrimSpace(billingMode))
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	inboundEndpoint = strings.ToLower(strings.TrimSpace(inboundEndpoint))
	if billingMode == string(service.BillingModeImage) || mediaType == "image" || strings.Contains(inboundEndpoint, "image") || strings.Contains(inboundEndpoint, "generate") || strings.Contains(inboundEndpoint, "edit") {
		return "image"
	}
	if mediaType == "video" || strings.Contains(inboundEndpoint, "video") {
		return "video"
	}
	if strings.Contains(inboundEndpoint, "chat") || strings.Contains(inboundEndpoint, "text") {
		return "text"
	}
	model = strings.ToLower(strings.TrimSpace(model))
	if strings.Contains(model, "image") || strings.HasPrefix(model, "gpt-image") {
		return "image"
	}
	if strings.Contains(model, "video") || strings.Contains(model, "seedance") {
		return "video"
	}
	return "text"
}

func studioBridgeUsageTypeLabel(mode, billingMode, mediaType, inboundEndpoint, model string) string {
	switch studioBridgeUsageRouteKindFromHints(mode, billingMode, mediaType, inboundEndpoint, model) {
	case "image":
		return "Image"
	case "video":
		return "Video"
	default:
		return "Text"
	}
}

func studioBridgeResolvedUsageModel(model, mode, billingMode, mediaType, inboundEndpoint string) string {
	model = strings.TrimSpace(model)
	if !strings.EqualFold(model, "auto") {
		return model
	}
	switch studioBridgeUsageRouteKindFromHints(mode, billingMode, mediaType, inboundEndpoint, model) {
	case "image":
		return "gpt-image-2"
	case "text":
		return "gpt-5.5"
	default:
		return model
	}
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

func studioBridgeUsageTaskID(requestID string) string {
	requestID = strings.TrimSpace(requestID)
	if taskID := strings.TrimPrefix(requestID, "studio:"); taskID != requestID {
		return taskID
	}
	return requestID
}

func studioBridgeUsageDurationSeconds(durationMs int64) int64 {
	if durationMs <= 0 {
		return 0
	}
	return (durationMs + 999) / 1000
}

func studioBridgeUsageStatus(actualCost float64) string {
	if actualCost > 0 {
		return "success"
	}
	return "failed"
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
