package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *affiliateRepository) ListAffiliateRiskClusters(ctx context.Context, windowStart, windowEnd time.Time) ([]service.AffiliateRiskCluster, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("affiliate risk repository unavailable")
	}
	if windowEnd.IsZero() {
		windowEnd = time.Now().UTC()
	}
	if windowStart.IsZero() || !windowStart.Before(windowEnd) {
		windowStart = windowEnd.Add(-12 * time.Hour)
	}

	rows, err := r.db.QueryContext(ctx, `
WITH candidates AS (
    SELECT ua.user_id AS invitee_id,
           ua.inviter_id,
           ua.affiliate_revoked_at,
           ua.affiliate_revoked_reason
    FROM user_affiliates ua
    JOIN users invitee ON invitee.id = ua.user_id
    WHERE ua.inviter_id IS NOT NULL
      AND (
          invitee.created_at >= $1
          AND invitee.created_at < $2
          OR EXISTS (
              SELECT 1
              FROM user_affiliate_ledger ual
              WHERE ual.user_id = ua.inviter_id
                AND ual.source_user_id = ua.user_id
                AND ual.action = 'api_call_reward'
                AND ual.created_at >= $1
                AND ual.created_at < $2
          )
      )
),
first_usage AS (
    SELECT DISTINCT ON (ul.user_id)
           ul.user_id,
           ul.created_at,
           COALESCE(ul.ip_address, '') AS ip_address
    FROM usage_logs ul
    JOIN candidates c ON c.invitee_id = ul.user_id
    WHERE ul.created_at >= $1
      AND ul.created_at < $2
    ORDER BY ul.user_id, ul.created_at ASC
),
api_rewards AS (
    SELECT ual.user_id AS inviter_id,
           ual.source_user_id AS invitee_id,
           MIN(ual.created_at) AS reward_at
    FROM user_affiliate_ledger ual
    JOIN candidates c
      ON c.inviter_id = ual.user_id
     AND c.invitee_id = ual.source_user_id
    WHERE ual.action = 'api_call_reward'
      AND ual.created_at < $2
    GROUP BY ual.user_id, ual.source_user_id
)
SELECT c.inviter_id,
       COALESCE(inviter.email, '') AS inviter_email,
       COALESCE(inviter.username, '') AS inviter_username,
       COALESCE(inviter.register_ip, '') AS inviter_register_ip,
       COALESCE(inviter.last_login_ip, '') AS inviter_last_login_ip,
       invitee.id AS invitee_id,
       COALESCE(invitee.email, '') AS invitee_email,
       COALESCE(invitee.username, '') AS invitee_username,
       COALESCE(invitee.register_ip, '') AS invitee_register_ip,
       COALESCE(invitee.last_login_ip, '') AS invitee_last_login_ip,
       invitee.created_at,
       c.affiliate_revoked_at,
       COALESCE(c.affiliate_revoked_reason, '') AS affiliate_revoked_reason,
       fu.created_at AS first_usage_at,
       COALESCE(fu.ip_address, '') AS first_usage_ip,
       ar.reward_at AS api_call_reward_at
FROM candidates c
JOIN users inviter ON inviter.id = c.inviter_id
JOIN users invitee ON invitee.id = c.invitee_id
LEFT JOIN first_usage fu ON fu.user_id = c.invitee_id
LEFT JOIN api_rewards ar ON ar.inviter_id = c.inviter_id AND ar.invitee_id = c.invitee_id
ORDER BY c.inviter_id, invitee.created_at ASC`, windowStart, windowEnd)
	if err != nil {
		return nil, fmt.Errorf("query affiliate risk clusters: %w", err)
	}
	defer func() { _ = rows.Close() }()

	clustersByInviter := make(map[int64]*service.AffiliateRiskCluster)
	order := make([]int64, 0)
	for rows.Next() {
		var (
			inviterID          int64
			inviterEmail       string
			inviterUsername    string
			inviterRegisterIP  string
			inviterLastLoginIP string
			invitee            service.AffiliateRiskInvitee
			affiliateRevokedAt sql.NullTime
			firstUsageAt       sql.NullTime
			apiCallRewardAt    sql.NullTime
		)
		if err := rows.Scan(
			&inviterID,
			&inviterEmail,
			&inviterUsername,
			&inviterRegisterIP,
			&inviterLastLoginIP,
			&invitee.UserID,
			&invitee.Email,
			&invitee.Username,
			&invitee.RegisterIP,
			&invitee.LastLoginIP,
			&invitee.CreatedAt,
			&affiliateRevokedAt,
			&invitee.AffiliateRevokedReason,
			&firstUsageAt,
			&invitee.FirstUsageIP,
			&apiCallRewardAt,
		); err != nil {
			return nil, err
		}
		if affiliateRevokedAt.Valid {
			t := affiliateRevokedAt.Time
			invitee.AffiliateRevokedAt = &t
		}
		if firstUsageAt.Valid {
			t := firstUsageAt.Time
			invitee.FirstUsageAt = &t
		}
		if apiCallRewardAt.Valid {
			t := apiCallRewardAt.Time
			invitee.APICallRewardAt = &t
		}

		cluster := clustersByInviter[inviterID]
		if cluster == nil {
			cluster = &service.AffiliateRiskCluster{
				InviterID:          inviterID,
				InviterEmail:       inviterEmail,
				InviterUsername:    inviterUsername,
				InviterRegisterIP:  inviterRegisterIP,
				InviterLastLoginIP: inviterLastLoginIP,
				Invitees:           []service.AffiliateRiskInvitee{},
			}
			clustersByInviter[inviterID] = cluster
			order = append(order, inviterID)
		}
		cluster.Invitees = append(cluster.Invitees, invitee)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]service.AffiliateRiskCluster, 0, len(order))
	for _, inviterID := range order {
		if cluster := clustersByInviter[inviterID]; cluster != nil {
			out = append(out, *cluster)
		}
	}
	return out, nil
}

func (r *affiliateRepository) UpsertAffiliateRiskFreeze(ctx context.Context, freeze service.AffiliateRiskFreeze) (bool, error) {
	if r == nil || r.db == nil {
		return false, fmt.Errorf("affiliate risk repository unavailable")
	}
	if freeze.InviterID <= 0 || freeze.Fingerprint == "" {
		return false, nil
	}
	res, err := r.db.ExecContext(ctx, `
INSERT INTO affiliate_risk_freezes (
    inviter_id,
    fingerprint,
    severity,
    score,
    reason,
    source_window_start,
    source_window_end,
    active,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, true, NOW(), NOW())
ON CONFLICT (inviter_id, fingerprint) WHERE active = true DO UPDATE
SET severity = EXCLUDED.severity,
    score = GREATEST(affiliate_risk_freezes.score, EXCLUDED.score),
    reason = EXCLUDED.reason,
    source_window_start = EXCLUDED.source_window_start,
    source_window_end = EXCLUDED.source_window_end,
    updated_at = NOW()
WHERE affiliate_risk_freezes.score <> EXCLUDED.score
   OR affiliate_risk_freezes.severity <> EXCLUDED.severity
   OR affiliate_risk_freezes.reason <> EXCLUDED.reason`,
		freeze.InviterID,
		freeze.Fingerprint,
		freeze.Severity,
		freeze.Score,
		freeze.Reason,
		freeze.SourceWindowStart,
		freeze.SourceWindowEnd,
	)
	if err != nil {
		return false, fmt.Errorf("upsert affiliate risk freeze: %w", err)
	}
	affected, _ := res.RowsAffected()
	return affected > 0, nil
}

func (r *affiliateRepository) HasActiveRiskFreeze(ctx context.Context, inviterID int64) (bool, error) {
	if r == nil || r.db == nil || inviterID <= 0 {
		return false, nil
	}
	row := r.db.QueryRowContext(ctx, `
SELECT EXISTS(
    SELECT 1
    FROM affiliate_risk_freezes
    WHERE inviter_id = $1
      AND active = true
    LIMIT 1
)`, inviterID)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("query affiliate risk freeze: %w", err)
	}
	return exists, nil
}
