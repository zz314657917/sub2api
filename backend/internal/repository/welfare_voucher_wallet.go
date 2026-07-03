package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const (
	welfareVoucherOperationUsageBilling = "usage_billing"
	welfareVoucherOperationStudioBridge = "studio_bridge"
	welfareVoucherOperationOpenAIVideo  = "openai_video"

	welfareVoucherPrecision = 100000000
)

type welfareVoucherGrantInput struct {
	UserID       int64
	SourceType   string
	SourceID     int64
	Amount       float64
	ExpiresAt    *time.Time
	RedeemCodeID *int64
}

type welfareVoucherDeductResult struct {
	VoucherAmount float64
	BalanceAmount float64
	BalanceAfter  float64
}

type welfareVoucherBalanceSummary struct {
	Balance          float64
	VoucherAvailable float64
	NextExpiresAt    *time.Time
}

type welfareVoucherDeductRefundResult struct {
	VoucherAmount float64
	BalanceAmount float64
	BalanceAfter  float64
}

func normalizeWelfareVoucherAmount(amount float64) float64 {
	if amount <= 0 || math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0
	}
	return math.Round(amount*welfareVoucherPrecision) / welfareVoucherPrecision
}

func welfareVoucherNullableTimeArg(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return value.UTC()
}

func welfareVoucherNullableInt64Arg(value *int64) any {
	if value == nil || *value <= 0 {
		return nil
	}
	return *value
}

func welfareVoucherOperationKey(operationType, operationKey string) (string, string) {
	return strings.TrimSpace(operationType), strings.TrimSpace(operationKey)
}

func grantWelfareVoucher(ctx context.Context, exec sqlQueryExecutor, input welfareVoucherGrantInput) error {
	if exec == nil {
		return fmt.Errorf("welfare voucher executor is nil")
	}
	input.Amount = normalizeWelfareVoucherAmount(input.Amount)
	input.SourceType = strings.TrimSpace(input.SourceType)
	if input.UserID <= 0 || input.SourceType == "" || input.SourceID <= 0 || input.Amount <= 0 {
		return nil
	}

	rows, err := exec.QueryContext(ctx, `
		INSERT INTO welfare_vouchers (
			user_id, source_type, source_id, amount, remaining_amount, expires_at, status, redeem_code_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $4, $5, 'active', $6, NOW(), NOW())
		ON CONFLICT (source_type, source_id) DO NOTHING
		RETURNING id, remaining_amount::double precision
	`,
		input.UserID,
		input.SourceType,
		input.SourceID,
		input.Amount,
		welfareVoucherNullableTimeArg(input.ExpiresAt),
		welfareVoucherNullableInt64Arg(input.RedeemCodeID),
	)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return rows.Err()
	}
	var voucherID int64
	var remainingAfter float64
	if err := rows.Scan(&voucherID, &remainingAfter); err != nil {
		return err
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = exec.ExecContext(ctx, `
		INSERT INTO welfare_voucher_ledger (
			voucher_id, user_id, operation, amount, remaining_after, operation_type, operation_key, metadata, created_at
		)
		VALUES ($1, $2, 'grant', $3, $4, $5::text, $6::text, jsonb_build_object('source_type', $5::text, 'source_id', $6::text), NOW())
		ON CONFLICT DO NOTHING
	`, voucherID, input.UserID, input.Amount, remainingAfter, input.SourceType, fmt.Sprintf("%d", input.SourceID))
	return err
}

func getWelfareVoucherBalanceSummary(ctx context.Context, exec sqlQueryExecutor, userID int64) (*welfareVoucherBalanceSummary, error) {
	if exec == nil {
		return nil, fmt.Errorf("welfare voucher executor is nil")
	}
	var summary welfareVoucherBalanceSummary
	var nextExpiresAt sql.NullTime
	rows, err := exec.QueryContext(ctx, `
		SELECT
			u.balance::double precision,
			COALESCE(voucher.available_amount, 0)::double precision,
			voucher.next_expires_at
		FROM users u
		LEFT JOIN LATERAL (
			SELECT
				SUM(v.remaining_amount) AS available_amount,
				MIN(v.expires_at) AS next_expires_at
			FROM welfare_vouchers v
			WHERE v.user_id = u.id
				AND v.status = 'active'
				AND v.remaining_amount > 0
				AND (v.expires_at IS NULL OR v.expires_at > NOW())
		) voucher ON TRUE
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrUserNotFound
	}
	if err := rows.Scan(&summary.Balance, &summary.VoucherAvailable, &nextExpiresAt); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if nextExpiresAt.Valid {
		summary.NextExpiresAt = &nextExpiresAt.Time
	}
	return &summary, nil
}

func getWelfareVoucherAvailableAmount(ctx context.Context, exec sqlQueryExecutor, userID int64) (float64, error) {
	summary, err := getWelfareVoucherBalanceSummary(ctx, exec, userID)
	if err != nil {
		return 0, err
	}
	return normalizeWelfareVoucherAmount(summary.VoucherAvailable), nil
}

func deductWelfareVoucherThenBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64, operationType, operationKey string, requireSufficient bool) (*welfareVoucherDeductResult, error) {
	if tx == nil {
		return nil, fmt.Errorf("welfare voucher transaction is nil")
	}
	amount = normalizeWelfareVoucherAmount(amount)
	if amount <= 0 {
		return &welfareVoucherDeductResult{}, nil
	}
	operationType, operationKey = welfareVoucherOperationKey(operationType, operationKey)

	var balance float64
	err := tx.QueryRowContext(ctx, `
		SELECT balance::double precision
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, userID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	if requireSufficient {
		available, err := getWelfareVoucherAvailableAmount(ctx, tx, userID)
		if err != nil {
			return nil, err
		}
		if normalizeWelfareVoucherAmount(balance+available)+1e-9 < amount {
			return nil, service.ErrInsufficientBalance
		}
	}

	remaining := amount
	result := &welfareVoucherDeductResult{BalanceAfter: balance}
	rows, err := tx.QueryContext(ctx, `
		SELECT id, remaining_amount::double precision
		FROM welfare_vouchers
		WHERE user_id = $1
			AND status = 'active'
			AND remaining_amount > 0
			AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY expires_at ASC NULLS LAST, id ASC
		FOR UPDATE
	`, userID)
	if err != nil {
		return nil, err
	}
	type voucherRow struct {
		id        int64
		remaining float64
	}
	vouchers := make([]voucherRow, 0)
	for rows.Next() {
		var row voucherRow
		if err := rows.Scan(&row.id, &row.remaining); err != nil {
			_ = rows.Close()
			return nil, err
		}
		vouchers = append(vouchers, row)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	for _, voucher := range vouchers {
		if remaining <= 0 {
			break
		}
		use := normalizeWelfareVoucherAmount(math.Min(remaining, voucher.remaining))
		if use <= 0 {
			continue
		}
		var remainingAfter float64
		err := tx.QueryRowContext(ctx, `
			UPDATE welfare_vouchers
			SET remaining_amount = remaining_amount - $1,
				status = CASE WHEN remaining_amount - $1 <= 0 THEN 'depleted' ELSE status END,
				updated_at = NOW()
			WHERE id = $2
			RETURNING remaining_amount::double precision
		`, use, voucher.id).Scan(&remainingAfter)
		if err != nil {
			return nil, err
		}
		if operationType != "" && operationKey != "" {
			_, err = tx.ExecContext(ctx, `
				INSERT INTO welfare_voucher_deductions (
					user_id, voucher_id, operation_type, operation_key, amount, refunded_amount, created_at, updated_at
				)
				VALUES ($1, $2, $3, $4, $5, 0, NOW(), NOW())
				ON CONFLICT (operation_type, operation_key, voucher_id) DO UPDATE
				SET amount = welfare_voucher_deductions.amount + EXCLUDED.amount,
					updated_at = NOW()
			`, userID, voucher.id, operationType, operationKey, use)
			if err != nil {
				return nil, err
			}
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO welfare_voucher_ledger (
				voucher_id, user_id, operation, amount, remaining_after, operation_type, operation_key, created_at
			)
			VALUES ($1, $2, 'consume', $3, $4, NULLIF($5, ''), NULLIF($6, ''), NOW())
		`, voucher.id, userID, use, remainingAfter, operationType, operationKey)
		if err != nil {
			return nil, err
		}
		result.VoucherAmount = normalizeWelfareVoucherAmount(result.VoucherAmount + use)
		remaining = normalizeWelfareVoucherAmount(remaining - use)
	}

	if remaining > 0 {
		result.BalanceAmount = remaining
		err := tx.QueryRowContext(ctx, `
			UPDATE users
			SET balance = balance - $1,
				updated_at = NOW()
			WHERE id = $2 AND deleted_at IS NULL
			RETURNING balance::double precision
		`, remaining, userID).Scan(&result.BalanceAfter)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrUserNotFound
		}
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func refundWelfareVoucherDeductions(ctx context.Context, tx *sql.Tx, userID int64, amount float64, operationType, operationKey string) (*welfareVoucherDeductRefundResult, error) {
	if tx == nil {
		return nil, fmt.Errorf("welfare voucher transaction is nil")
	}
	amount = normalizeWelfareVoucherAmount(amount)
	result := &welfareVoucherDeductRefundResult{}
	if amount <= 0 {
		err := tx.QueryRowContext(ctx, `
			SELECT balance::double precision
			FROM users
			WHERE id = $1 AND deleted_at IS NULL
		`, userID).Scan(&result.BalanceAfter)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrUserNotFound
		}
		return result, err
	}
	operationType, operationKey = welfareVoucherOperationKey(operationType, operationKey)

	remaining := amount
	rows, err := tx.QueryContext(ctx, `
		SELECT id, voucher_id, (amount - refunded_amount)::double precision
		FROM welfare_voucher_deductions
		WHERE user_id = $1
			AND operation_type = $2
			AND operation_key = $3
			AND amount > refunded_amount
		ORDER BY id DESC
		FOR UPDATE
	`, userID, operationType, operationKey)
	if err != nil {
		return nil, err
	}
	type deductionRow struct {
		id         int64
		voucherID  int64
		refundable float64
	}
	deductions := make([]deductionRow, 0)
	for rows.Next() {
		var row deductionRow
		if err := rows.Scan(&row.id, &row.voucherID, &row.refundable); err != nil {
			_ = rows.Close()
			return nil, err
		}
		deductions = append(deductions, row)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	for _, deduction := range deductions {
		if remaining <= 0 {
			break
		}
		refund := normalizeWelfareVoucherAmount(math.Min(remaining, deduction.refundable))
		if refund <= 0 {
			continue
		}
		var remainingAfter float64
		err := tx.QueryRowContext(ctx, `
			UPDATE welfare_vouchers
			SET remaining_amount = LEAST(amount, remaining_amount + $1),
				status = CASE
					WHEN expires_at IS NOT NULL AND expires_at <= NOW() THEN 'expired'
					ELSE 'active'
				END,
				updated_at = NOW()
			WHERE id = $2
			RETURNING remaining_amount::double precision
		`, refund, deduction.voucherID).Scan(&remainingAfter)
		if err != nil {
			return nil, err
		}
		_, err = tx.ExecContext(ctx, `
			UPDATE welfare_voucher_deductions
			SET refunded_amount = refunded_amount + $1,
				updated_at = NOW()
			WHERE id = $2
		`, refund, deduction.id)
		if err != nil {
			return nil, err
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO welfare_voucher_ledger (
				voucher_id, user_id, operation, amount, remaining_after, operation_type, operation_key, created_at
			)
			VALUES ($1, $2, 'refund', $3, $4, $5, $6, NOW())
		`, deduction.voucherID, userID, refund, remainingAfter, operationType, operationKey)
		if err != nil {
			return nil, err
		}
		remaining = normalizeWelfareVoucherAmount(remaining - refund)
		result.VoucherAmount = normalizeWelfareVoucherAmount(result.VoucherAmount + refund)
	}

	if remaining > 0 {
		result.BalanceAmount = remaining
		err = tx.QueryRowContext(ctx, `
			UPDATE users
			SET balance = balance + $1,
				updated_at = NOW()
			WHERE id = $2 AND deleted_at IS NULL
			RETURNING balance::double precision
		`, remaining, userID).Scan(&result.BalanceAfter)
	} else {
		err = tx.QueryRowContext(ctx, `
			SELECT balance::double precision
			FROM users
			WHERE id = $1 AND deleted_at IS NULL
		`, userID).Scan(&result.BalanceAfter)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrUserNotFound
	}
	return result, err
}

func rebindWelfareVoucherDeductions(ctx context.Context, tx *sql.Tx, userID int64, operationType, oldOperationKey, newOperationKey string) error {
	if tx == nil {
		return fmt.Errorf("welfare voucher transaction is nil")
	}
	operationType, oldOperationKey = welfareVoucherOperationKey(operationType, oldOperationKey)
	_, newOperationKey = welfareVoucherOperationKey(operationType, newOperationKey)
	if userID <= 0 || operationType == "" || oldOperationKey == "" || newOperationKey == "" || oldOperationKey == newOperationKey {
		return nil
	}
	_, err := tx.ExecContext(ctx, `
		UPDATE welfare_voucher_deductions
		SET operation_key = $4,
			updated_at = NOW()
		WHERE user_id = $1
			AND operation_type = $2
			AND operation_key = $3
	`, userID, operationType, oldOperationKey, newOperationKey)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE welfare_voucher_ledger
		SET operation_key = $4
		WHERE user_id = $1
			AND operation_type = $2
			AND operation_key = $3
	`, userID, operationType, oldOperationKey, newOperationKey)
	return err
}
