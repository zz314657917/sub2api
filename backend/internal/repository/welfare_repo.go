package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type welfareRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

func NewWelfareRepository(client *dbent.Client, sqlDB *sql.DB) service.WelfareRepository {
	return &welfareRepository{client: client, sql: sqlDB}
}

func (r *welfareRepository) executor(ctx context.Context) (sqlQueryExecutor, error) {
	exec := txAwareSQLExecutor(ctx, r.sql, r.client)
	if exec == nil {
		return nil, fmt.Errorf("welfare sql executor is not configured")
	}
	return exec, nil
}

func (r *welfareRepository) withTx(ctx context.Context, fn func(context.Context) error) error {
	if dbent.TxFromContext(ctx) != nil {
		return fn(ctx)
	}
	if r.client == nil {
		return fn(ctx)
	}

	tx, err := r.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := dbent.NewTxContext(ctx, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *welfareRepository) GetDailyCheckin(ctx context.Context, checkinDate string, userID int64) (*service.WelfareDailyCheckinRecord, error) {
	exec, err := r.executor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := exec.QueryContext(ctx, `
		SELECT id, checkin_date::text, reward_month, user_id, amount::double precision, redeem_code_id, created_at
		FROM welfare_daily_checkins
		WHERE checkin_date = $1 AND user_id = $2
	`, checkinDate, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrWelfareDailyCheckinNotFound
	}
	record, err := scanWelfareDailyCheckin(rows)
	if err != nil {
		return nil, err
	}
	return record, rows.Err()
}

func (r *welfareRepository) ListDailyCheckins(ctx context.Context, userID int64, month string) ([]service.WelfareDailyCheckinRecord, error) {
	exec, err := r.executor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := exec.QueryContext(ctx, `
		SELECT id, checkin_date::text, reward_month, user_id, amount::double precision, redeem_code_id, created_at
		FROM welfare_daily_checkins
		WHERE user_id = $1 AND reward_month = $2
		ORDER BY checkin_date ASC
	`, userID, month)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make([]service.WelfareDailyCheckinRecord, 0)
	for rows.Next() {
		record, err := scanWelfareDailyCheckin(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *record)
	}
	return result, rows.Err()
}

func (r *welfareRepository) CreateDailyCheckin(ctx context.Context, record *service.WelfareDailyCheckinRecord) error {
	if record == nil {
		return fmt.Errorf("nil welfare daily checkin record")
	}
	exec, err := r.executor(ctx)
	if err != nil {
		return err
	}
	rows, err := exec.QueryContext(ctx, `
		INSERT INTO welfare_daily_checkins (
			checkin_date, reward_month, user_id, amount, redeem_code_id
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, record.CheckinDate, record.RewardMonth, record.UserID, record.Amount, record.RedeemCodeID)
	if err != nil {
		if isUniqueConstraintViolation(err) {
			return service.ErrWelfareDailyCheckinAlreadyClaimed
		}
		return err
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return fmt.Errorf("create welfare daily checkin returned no row")
	}
	if err := rows.Scan(&record.ID, &record.CreatedAt); err != nil {
		return err
	}
	return rows.Err()
}

func (r *welfareRepository) AttachDailyCheckinRedeemCode(ctx context.Context, claimID, redeemCodeID int64) error {
	exec, err := r.executor(ctx)
	if err != nil {
		return err
	}
	result, err := exec.ExecContext(ctx, `
		UPDATE welfare_daily_checkins
		SET redeem_code_id = $2
		WHERE id = $1
	`, claimID, redeemCodeID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrWelfareDailyCheckinNotFound
	}
	return nil
}

func (r *welfareRepository) GetDailyCheckinMilestoneClaim(ctx context.Context, month string, milestoneDay int, userID int64) (*service.WelfareDailyCheckinMilestoneClaim, error) {
	exec, err := r.executor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := exec.QueryContext(ctx, `
		SELECT id, reward_month, milestone_day, user_id, amount::double precision, redeem_code_id, created_at
		FROM welfare_daily_checkin_milestone_claims
		WHERE reward_month = $1 AND milestone_day = $2 AND user_id = $3
	`, month, milestoneDay, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrWelfareCheckinMilestoneNotFound
	}
	claim, err := scanWelfareDailyCheckinMilestoneClaim(rows)
	if err != nil {
		return nil, err
	}
	return claim, rows.Err()
}

func (r *welfareRepository) ListDailyCheckinMilestoneClaims(ctx context.Context, month string, userID int64) ([]service.WelfareDailyCheckinMilestoneClaim, error) {
	exec, err := r.executor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := exec.QueryContext(ctx, `
		SELECT id, reward_month, milestone_day, user_id, amount::double precision, redeem_code_id, created_at
		FROM welfare_daily_checkin_milestone_claims
		WHERE reward_month = $1 AND user_id = $2
		ORDER BY milestone_day ASC
	`, month, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make([]service.WelfareDailyCheckinMilestoneClaim, 0)
	for rows.Next() {
		claim, err := scanWelfareDailyCheckinMilestoneClaim(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *claim)
	}
	return result, rows.Err()
}

func (r *welfareRepository) CreateDailyCheckinMilestoneClaim(ctx context.Context, claim *service.WelfareDailyCheckinMilestoneClaim) error {
	if claim == nil {
		return fmt.Errorf("nil welfare daily checkin milestone claim")
	}
	exec, err := r.executor(ctx)
	if err != nil {
		return err
	}
	rows, err := exec.QueryContext(ctx, `
		INSERT INTO welfare_daily_checkin_milestone_claims (
			reward_month, milestone_day, user_id, amount, redeem_code_id
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, claim.RewardMonth, claim.MilestoneDay, claim.UserID, claim.Amount, claim.RedeemCodeID)
	if err != nil {
		if isUniqueConstraintViolation(err) {
			return service.ErrWelfareCheckinMilestoneAlreadyClaimed
		}
		return err
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return fmt.Errorf("create welfare daily checkin milestone claim returned no row")
	}
	if err := rows.Scan(&claim.ID, &claim.CreatedAt); err != nil {
		return err
	}
	return rows.Err()
}

func (r *welfareRepository) AttachDailyCheckinMilestoneRedeemCode(ctx context.Context, claimID, redeemCodeID int64) error {
	exec, err := r.executor(ctx)
	if err != nil {
		return err
	}
	result, err := exec.ExecContext(ctx, `
		UPDATE welfare_daily_checkin_milestone_claims
		SET redeem_code_id = $2
		WHERE id = $1
	`, claimID, redeemCodeID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return service.ErrWelfareCheckinMilestoneNotFound
	}
	return nil
}

func (r *welfareRepository) GetNewUserTrial(ctx context.Context, userID int64) (*service.WelfareNewUserTrial, error) {
	exec, err := r.executor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := exec.QueryContext(ctx, `
		SELECT id, user_id, quota_amount::double precision, quota_used::double precision,
			status, activated_ip, first_started_at, first_success_at, last_request_id, created_at, updated_at
		FROM welfare_new_user_trials
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrWelfareNewUserTrialNotFound
	}
	trial, err := scanWelfareNewUserTrial(rows)
	if err != nil {
		return nil, err
	}
	return trial, rows.Err()
}

func (r *welfareRepository) BeginNewUserTrial(ctx context.Context, userID int64, quotaAmount float64, clientIP, requestID string, dayStart time.Time, ipActivationLimit int) (*service.WelfareNewUserTrial, error) {
	var trial *service.WelfareNewUserTrial
	err := r.withTx(ctx, func(txCtx context.Context) error {
		created, err := r.beginNewUserTrialTx(txCtx, userID, quotaAmount, clientIP, requestID, dayStart, ipActivationLimit)
		if err != nil {
			return err
		}
		trial = created
		return nil
	})
	return trial, err
}

func (r *welfareRepository) beginNewUserTrialTx(ctx context.Context, userID int64, quotaAmount float64, clientIP, requestID string, dayStart time.Time, ipActivationLimit int) (*service.WelfareNewUserTrial, error) {
	exec, err := r.executor(ctx)
	if err != nil {
		return nil, err
	}
	normalizedIP := strings.TrimSpace(clientIP)
	if ipActivationLimit > 0 && normalizedIP != "" {
		releaseLocks, err := lockRepositoryScopedKeys(ctx, r.client, exec, "welfare:new-user-trial:ip:"+normalizedIP)
		if err != nil {
			return nil, err
		}
		defer releaseLocks()

		current, err := r.GetNewUserTrial(ctx, userID)
		if err != nil && !errors.Is(err, service.ErrWelfareNewUserTrialNotFound) {
			return nil, err
		}
		if current == nil {
			count, err := r.CountNewUserTrialActivationsByIPSince(ctx, normalizedIP, dayStart)
			if err != nil {
				return nil, err
			}
			if count >= ipActivationLimit {
				return nil, service.ErrWelfareNewUserTrialDailyLimitExceeded
			}
		}
	}

	rows, err := exec.QueryContext(ctx, `
		WITH upserted AS (
			INSERT INTO welfare_new_user_trials (
				user_id, quota_amount, quota_used, status, activated_ip, first_started_at, last_request_id, updated_at
			)
			VALUES ($1, $2, 0, 'in_progress', NULLIF($3, ''), NOW(), $4, NOW())
			ON CONFLICT (user_id) DO UPDATE
			SET status = CASE
					WHEN welfare_new_user_trials.quota_used >= welfare_new_user_trials.quota_amount THEN 'exhausted'
					ELSE 'in_progress'
				END,
				quota_amount = CASE
					WHEN welfare_new_user_trials.quota_amount <= 0 THEN EXCLUDED.quota_amount
					ELSE welfare_new_user_trials.quota_amount
				END,
				activated_ip = COALESCE(welfare_new_user_trials.activated_ip, NULLIF($3, '')),
				first_started_at = COALESCE(welfare_new_user_trials.first_started_at, NOW()),
				last_request_id = $4,
				updated_at = NOW()
			WHERE welfare_new_user_trials.status IN ('available', 'active')
				AND welfare_new_user_trials.quota_used < welfare_new_user_trials.quota_amount
			RETURNING id, user_id, quota_amount::double precision, quota_used::double precision,
				status, activated_ip, first_started_at, first_success_at, last_request_id, created_at, updated_at
		)
		SELECT * FROM upserted
	`, userID, quotaAmount, normalizedIP, requestID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if rows.Next() {
		trial, err := scanWelfareNewUserTrial(rows)
		if err != nil {
			return nil, err
		}
		return trial, rows.Err()
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	current, getErr := r.GetNewUserTrial(ctx, userID)
	if getErr != nil {
		return nil, getErr
	}
	if current.Status == "in_progress" {
		return nil, service.ErrWelfareNewUserTrialAlreadyInProgress
	}
	return nil, service.ErrWelfareNewUserTrialExhausted
}

func (r *welfareRepository) CancelNewUserTrial(ctx context.Context, trialID int64, requestID string) error {
	exec, err := r.executor(ctx)
	if err != nil {
		return err
	}
	_, err = exec.ExecContext(ctx, `
		UPDATE welfare_new_user_trials t
		SET status = CASE
				WHEN t.first_success_at IS NULL THEN 'available'
				WHEN t.quota_used >= t.quota_amount THEN 'exhausted'
				ELSE 'active'
			END,
			updated_at = NOW()
		WHERE t.id = $1
			AND t.status = 'in_progress'
			AND t.last_request_id = $2
			AND NOT EXISTS (
				SELECT 1 FROM welfare_new_user_trial_usages u WHERE u.request_id = $2
			)
	`, trialID, requestID)
	return err
}

func (r *welfareRepository) ConsumeNewUserTrial(ctx context.Context, input service.WelfareNewUserTrialConsumeInput) (*service.WelfareNewUserTrial, bool, error) {
	if input.Amount <= 0 {
		trial, err := r.GetNewUserTrial(ctx, input.UserID)
		return trial, false, err
	}
	exec, err := r.executor(ctx)
	if err != nil {
		return nil, false, err
	}
	rows, err := exec.QueryContext(ctx, `
		WITH inserted AS (
			INSERT INTO welfare_new_user_trial_usages (
				trial_id, user_id, request_id, amount, model, api_key_id
			)
			VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, 0))
			ON CONFLICT (request_id) DO NOTHING
			RETURNING amount
		),
		updated AS (
			UPDATE welfare_new_user_trials t
			SET quota_used = LEAST(t.quota_amount, t.quota_used + (SELECT amount FROM inserted)),
				status = CASE
					WHEN t.quota_used + (SELECT amount FROM inserted) >= t.quota_amount THEN 'exhausted'
					ELSE 'active'
				END,
				first_success_at = COALESCE(t.first_success_at, NOW()),
				updated_at = NOW()
			WHERE t.id = $1
				AND t.user_id = $2
				AND t.last_request_id = $7
				AND EXISTS (SELECT 1 FROM inserted)
			RETURNING t.id, t.user_id, t.quota_amount::double precision, t.quota_used::double precision,
				t.status, t.activated_ip, t.first_started_at, t.first_success_at, t.last_request_id, t.created_at, t.updated_at
		)
		SELECT *, TRUE AS applied FROM updated
		UNION ALL
		SELECT t.id, t.user_id, t.quota_amount::double precision, t.quota_used::double precision,
			t.status, t.activated_ip, t.first_started_at, t.first_success_at, t.last_request_id, t.created_at, t.updated_at,
			FALSE AS applied
		FROM welfare_new_user_trials t
		WHERE t.id = $1 AND NOT EXISTS (SELECT 1 FROM updated)
		LIMIT 1
	`, input.TrialID, input.UserID, input.RequestID, input.Amount, input.Model, input.APIKeyID, input.TrialRequestID)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, false, err
		}
		return nil, false, service.ErrWelfareNewUserTrialNotFound
	}
	trial, applied, err := scanWelfareNewUserTrialWithApplied(rows)
	if err != nil {
		return nil, false, err
	}
	return trial, applied, rows.Err()
}

func (r *welfareRepository) SumNewUserTrialUsageSince(ctx context.Context, since time.Time) (float64, error) {
	exec, err := r.executor(ctx)
	if err != nil {
		return 0, err
	}
	rows, err := exec.QueryContext(ctx, `
		SELECT COALESCE(SUM(amount), 0)::double precision
		FROM welfare_new_user_trial_usages
		WHERE created_at >= $1
	`, since)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	var amount float64
	if rows.Next() {
		if err := rows.Scan(&amount); err != nil {
			return 0, err
		}
	}
	return amount, rows.Err()
}

func (r *welfareRepository) CountNewUserTrialActivationsByIPSince(ctx context.Context, clientIP string, since time.Time) (int, error) {
	exec, err := r.executor(ctx)
	if err != nil {
		return 0, err
	}
	rows, err := exec.QueryContext(ctx, `
		SELECT COUNT(*)::integer
		FROM welfare_new_user_trials
		WHERE activated_ip = $1 AND first_started_at >= $2
	`, clientIP, since)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()
	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
	}
	return count, rows.Err()
}

func (r *welfareRepository) FirstSuccessfulUsageAt(ctx context.Context, userID int64) (*time.Time, error) {
	exec, err := r.executor(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := exec.QueryContext(ctx, `
		SELECT created_at
		FROM usage_logs
		WHERE user_id = $1
		ORDER BY created_at ASC
		LIMIT 1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if rows.Next() {
		var firstSuccessAt time.Time
		if err := rows.Scan(&firstSuccessAt); err != nil {
			return nil, err
		}
		return &firstSuccessAt, rows.Err()
	}
	return nil, rows.Err()
}

type welfareDailyCheckinScanner interface {
	Scan(dest ...any) error
}

func scanWelfareDailyCheckin(scanner welfareDailyCheckinScanner) (*service.WelfareDailyCheckinRecord, error) {
	var record service.WelfareDailyCheckinRecord
	var redeemCodeID sql.NullInt64
	if err := scanner.Scan(
		&record.ID,
		&record.CheckinDate,
		&record.RewardMonth,
		&record.UserID,
		&record.Amount,
		&redeemCodeID,
		&record.CreatedAt,
	); err != nil {
		return nil, err
	}
	if redeemCodeID.Valid {
		record.RedeemCodeID = &redeemCodeID.Int64
	}
	return &record, nil
}

func scanWelfareDailyCheckinMilestoneClaim(scanner welfareDailyCheckinScanner) (*service.WelfareDailyCheckinMilestoneClaim, error) {
	var claim service.WelfareDailyCheckinMilestoneClaim
	var redeemCodeID sql.NullInt64
	if err := scanner.Scan(
		&claim.ID,
		&claim.RewardMonth,
		&claim.MilestoneDay,
		&claim.UserID,
		&claim.Amount,
		&redeemCodeID,
		&claim.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrWelfareCheckinMilestoneNotFound
		}
		return nil, err
	}
	if redeemCodeID.Valid {
		claim.RedeemCodeID = &redeemCodeID.Int64
	}
	return &claim, nil
}

func scanWelfareNewUserTrial(scanner welfareDailyCheckinScanner) (*service.WelfareNewUserTrial, error) {
	var trial service.WelfareNewUserTrial
	var activatedIP sql.NullString
	var firstStartedAt sql.NullTime
	var firstSuccessAt sql.NullTime
	var lastRequestID sql.NullString
	if err := scanner.Scan(
		&trial.ID,
		&trial.UserID,
		&trial.QuotaAmount,
		&trial.QuotaUsed,
		&trial.Status,
		&activatedIP,
		&firstStartedAt,
		&firstSuccessAt,
		&lastRequestID,
		&trial.CreatedAt,
		&trial.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrWelfareNewUserTrialNotFound
		}
		return nil, err
	}
	if activatedIP.Valid {
		trial.ActivatedIP = activatedIP.String
	}
	if firstStartedAt.Valid {
		trial.FirstStartedAt = &firstStartedAt.Time
	}
	if firstSuccessAt.Valid {
		trial.FirstSuccessAt = &firstSuccessAt.Time
	}
	if lastRequestID.Valid {
		trial.LastRequestID = lastRequestID.String
	}
	return &trial, nil
}

func scanWelfareNewUserTrialWithApplied(scanner welfareDailyCheckinScanner) (*service.WelfareNewUserTrial, bool, error) {
	var trial service.WelfareNewUserTrial
	var activatedIP sql.NullString
	var firstStartedAt sql.NullTime
	var firstSuccessAt sql.NullTime
	var lastRequestID sql.NullString
	var applied bool
	if err := scanner.Scan(
		&trial.ID,
		&trial.UserID,
		&trial.QuotaAmount,
		&trial.QuotaUsed,
		&trial.Status,
		&activatedIP,
		&firstStartedAt,
		&firstSuccessAt,
		&lastRequestID,
		&trial.CreatedAt,
		&trial.UpdatedAt,
		&applied,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, service.ErrWelfareNewUserTrialNotFound
		}
		return nil, false, err
	}
	if activatedIP.Valid {
		trial.ActivatedIP = activatedIP.String
	}
	if firstStartedAt.Valid {
		trial.FirstStartedAt = &firstStartedAt.Time
	}
	if firstSuccessAt.Valid {
		trial.FirstSuccessAt = &firstSuccessAt.Time
	}
	if lastRequestID.Valid {
		trial.LastRequestID = lastRequestID.String
	}
	return &trial, applied, nil
}
