package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type membershipRepository struct {
	db *sql.DB
}

func NewMembershipRepository(db *sql.DB) service.MembershipRepository {
	return &membershipRepository{db: db}
}

func (r *membershipRepository) GetMonthlyNetPaid(ctx context.Context, userID int64, start, end time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(GREATEST(pay_amount - CASE WHEN refund_at IS NOT NULL THEN COALESCE(refund_amount, 0) ELSE 0 END, 0)), 0)
		FROM payment_orders
		WHERE user_id = $1
		  AND completed_at IS NOT NULL
		  AND completed_at >= $2
		  AND completed_at < $3
	`
	var amount float64
	if err := r.db.QueryRowContext(ctx, query, userID, start, end).Scan(&amount); err != nil {
		return 0, err
	}
	return amount, nil
}

func (r *membershipRepository) UpsertAutoGrant(ctx context.Context, grant *service.MembershipGrant) (*service.MembershipGrant, bool, error) {
	if grant == nil {
		return nil, false, nil
	}
	query := `
		INSERT INTO user_membership_grants (
			user_id, tier, source, period_key, period_start, period_end,
			qualified_amount, starts_at, expires_at, status,
			subscription_group_id, source_order_id
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12
		)
		ON CONFLICT (user_id, source, period_key, tier)
		DO UPDATE SET
			qualified_amount = EXCLUDED.qualified_amount,
			source_order_id = COALESCE(user_membership_grants.source_order_id, EXCLUDED.source_order_id),
			subscription_group_id = CASE
				WHEN user_membership_grants.status = $13 THEN COALESCE(user_membership_grants.subscription_group_id, EXCLUDED.subscription_group_id)
				ELSE EXCLUDED.subscription_group_id
			END,
			subscription_id = CASE
				WHEN user_membership_grants.status = $13 THEN user_membership_grants.subscription_id
				ELSE NULL
			END,
			starts_at = CASE
				WHEN user_membership_grants.status = $13 THEN user_membership_grants.starts_at
				ELSE EXCLUDED.starts_at
			END,
			expires_at = CASE
				WHEN user_membership_grants.status = $13 THEN user_membership_grants.expires_at
				ELSE EXCLUDED.expires_at
			END,
			status = CASE
				WHEN user_membership_grants.status = $13 THEN user_membership_grants.status
				ELSE EXCLUDED.status
			END,
			revoked_at = CASE
				WHEN user_membership_grants.status = $13 THEN user_membership_grants.revoked_at
				ELSE NULL
			END,
			revoke_reason = CASE
				WHEN user_membership_grants.status = $13 THEN user_membership_grants.revoke_reason
				ELSE NULL
			END,
			updated_at = NOW()
		RETURNING
			id, user_id, tier, source, period_key, period_start, period_end,
			qualified_amount, starts_at, expires_at, status, subscription_id,
			subscription_group_id, source_order_id, revoked_at, revoke_reason,
			created_at, updated_at, (xmax = 0) AS inserted
	`
	out := &service.MembershipGrant{}
	var inserted bool
	err := r.db.QueryRowContext(ctx, query,
		grant.UserID,
		grant.Tier,
		grant.Source,
		grant.PeriodKey,
		grant.PeriodStart,
		grant.PeriodEnd,
		grant.QualifiedAmount,
		grant.StartsAt,
		grant.ExpiresAt,
		grant.Status,
		grant.SubscriptionGroupID,
		grant.SourceOrderID,
		service.MembershipGrantStatusActive,
	).Scan(
		&out.ID,
		&out.UserID,
		&out.Tier,
		&out.Source,
		&out.PeriodKey,
		&out.PeriodStart,
		&out.PeriodEnd,
		&out.QualifiedAmount,
		&out.StartsAt,
		&out.ExpiresAt,
		&out.Status,
		&out.SubscriptionID,
		&out.SubscriptionGroupID,
		&out.SourceOrderID,
		&out.RevokedAt,
		&out.RevokeReason,
		&out.CreatedAt,
		&out.UpdatedAt,
		&inserted,
	)
	if err != nil {
		return nil, false, err
	}
	return out, inserted, nil
}

func (r *membershipRepository) UpdateGrantSubscription(ctx context.Context, grantID, subscriptionID, groupID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_membership_grants
		SET subscription_id = $2,
			subscription_group_id = $3,
			updated_at = NOW()
		WHERE id = $1
	`, grantID, subscriptionID, groupID)
	return err
}

func (r *membershipRepository) ListActiveAutoGrants(ctx context.Context, userID int64, now time.Time) ([]service.MembershipGrant, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, tier, source, period_key, period_start, period_end,
			qualified_amount, starts_at, expires_at, status, subscription_id,
			subscription_group_id, source_order_id, revoked_at, revoke_reason,
			created_at, updated_at
		FROM user_membership_grants
		WHERE user_id = $1
		  AND source = $2
		  AND status = $3
		  AND starts_at <= $4
		  AND expires_at > $4
		ORDER BY expires_at DESC, id DESC
	`, userID, service.MembershipGrantSourceAutoMonthlySpend, service.MembershipGrantStatusActive, now)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var grants []service.MembershipGrant
	for rows.Next() {
		grant, err := scanMembershipGrant(rows)
		if err != nil {
			return nil, err
		}
		grants = append(grants, *grant)
	}
	return grants, rows.Err()
}

func (r *membershipRepository) GetActiveHighestAutoGrant(ctx context.Context, userID int64, now time.Time) (*service.MembershipGrant, error) {
	query := `
		SELECT id, user_id, tier, source, period_key, period_start, period_end,
			qualified_amount, starts_at, expires_at, status, subscription_id,
			subscription_group_id, source_order_id, revoked_at, revoke_reason,
			created_at, updated_at
		FROM user_membership_grants
		WHERE user_id = $1
		  AND source = $2
		  AND status = $3
		  AND starts_at <= $4
		  AND expires_at > $4
		ORDER BY
			CASE tier
				WHEN $5 THEN 3
				WHEN $6 THEN 2
				WHEN $7 THEN 1
				ELSE 0
			END DESC,
			expires_at DESC,
			id DESC
		LIMIT 1
	`
	rows, err := r.db.QueryContext(ctx, query,
		userID,
		service.MembershipGrantSourceAutoMonthlySpend,
		service.MembershipGrantStatusActive,
		now,
		service.MembershipTierSVIP,
		service.MembershipTierVIP,
		service.MembershipTierNormal,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}
	grant, err := scanMembershipGrant(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return grant, nil
}

func (r *membershipRepository) RevokeGrant(ctx context.Context, grantID int64, reason string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_membership_grants
		SET status = $2,
			revoked_at = NOW(),
			revoke_reason = $3,
			updated_at = NOW()
		WHERE id = $1 AND status = $4
	`, grantID, service.MembershipGrantStatusRevoked, reason, service.MembershipGrantStatusActive)
	return err
}

func scanMembershipGrant(scanner interface {
	Scan(dest ...any) error
}) (*service.MembershipGrant, error) {
	out := &service.MembershipGrant{}
	if err := scanner.Scan(
		&out.ID,
		&out.UserID,
		&out.Tier,
		&out.Source,
		&out.PeriodKey,
		&out.PeriodStart,
		&out.PeriodEnd,
		&out.QualifiedAmount,
		&out.StartsAt,
		&out.ExpiresAt,
		&out.Status,
		&out.SubscriptionID,
		&out.SubscriptionGroupID,
		&out.SourceOrderID,
		&out.RevokedAt,
		&out.RevokeReason,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return out, nil
}
