-- Pixel Cafe S252 share fulfillment. This migration is additive and does not
-- rewrite legacy Cafe seat/key/binding rows.

ALTER TABLE group_buy_plans
    ADD COLUMN IF NOT EXISTS subscription_tier VARCHAR(16) NOT NULL DEFAULT 'plus',
    ADD COLUMN IF NOT EXISTS max_buyers INTEGER NOT NULL DEFAULT 4,
    ADD COLUMN IF NOT EXISTS fulfillment_timeout_minutes INTEGER NOT NULL DEFAULT 1440;

ALTER TABLE group_buy_plans
    DROP CONSTRAINT IF EXISTS group_buy_plans_subscription_tier_check,
    DROP CONSTRAINT IF EXISTS group_buy_plans_max_buyers_check,
    DROP CONSTRAINT IF EXISTS group_buy_plans_max_shares_per_user_check,
    DROP CONSTRAINT IF EXISTS group_buy_plans_fulfillment_timeout_minutes_check;

ALTER TABLE group_buy_plans
    ADD CONSTRAINT group_buy_plans_subscription_tier_check CHECK (fulfillment_mode <> 'room_subscription' OR subscription_tier IN ('plus', 'pro')),
    ADD CONSTRAINT group_buy_plans_max_buyers_check CHECK (fulfillment_mode <> 'room_subscription' OR (max_buyers >= 1 AND max_buyers <= total_shares)),
    ADD CONSTRAINT group_buy_plans_max_shares_per_user_check CHECK (fulfillment_mode <> 'room_subscription' OR (max_shares_per_user >= 1 AND max_shares_per_user <= total_shares)),
    ADD CONSTRAINT group_buy_plans_fulfillment_timeout_minutes_check CHECK (fulfillment_mode <> 'room_subscription' OR fulfillment_timeout_minutes > 0);

ALTER TABLE group_buy_rounds
    ADD COLUMN IF NOT EXISTS cafe_fulfillment_version VARCHAR(32) NOT NULL DEFAULT 'legacy_seat',
    ADD COLUMN IF NOT EXISTS subscription_tier VARCHAR(16),
    ADD COLUMN IF NOT EXISTS max_buyers INTEGER,
    ADD COLUMN IF NOT EXISTS max_shares_per_user INTEGER,
    ADD COLUMN IF NOT EXISTS fulfillment_timeout_minutes INTEGER,
    ADD COLUMN IF NOT EXISTS validity_days_snapshot INTEGER,
    ADD COLUMN IF NOT EXISTS target_group_id_snapshot BIGINT,
    ADD COLUMN IF NOT EXISTS platform_snapshot VARCHAR(64),
    ADD COLUMN IF NOT EXISTS quota_per_share_snapshot DECIMAL(20,8),
    ADD COLUMN IF NOT EXISTS rate_limit_5h_per_share_snapshot DECIMAL(20,8),
    ADD COLUMN IF NOT EXISTS rate_limit_1d_per_share_snapshot DECIMAL(20,8),
    ADD COLUMN IF NOT EXISTS rate_limit_7d_per_share_snapshot DECIMAL(20,8),
    ADD COLUMN IF NOT EXISTS fulfillment_deadline_at TIMESTAMPTZ;

UPDATE group_buy_rounds
SET cafe_fulfillment_version = 'legacy_seat'
WHERE cafe_fulfillment_version IS NULL OR cafe_fulfillment_version = '';

ALTER TABLE group_buy_rounds
    DROP CONSTRAINT IF EXISTS group_buy_rounds_status_check;

ALTER TABLE group_buy_rounds
    ADD CONSTRAINT group_buy_rounds_status_check CHECK (status IN ('open', 'awaiting_account', 'activating', 'active', 'completed', 'refunding', 'refunded', 'failed', 'cancelled'));

CREATE TABLE IF NOT EXISTS cafe_round_memberships (
    id BIGSERIAL PRIMARY KEY,
    round_id BIGINT NOT NULL REFERENCES group_buy_rounds(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(24) NOT NULL DEFAULT 'locked' CHECK (status IN ('locked', 'paid', 'active', 'refunding', 'refunded', 'cancelled')),
    paid_shares INTEGER NOT NULL DEFAULT 0 CHECK (paid_shares >= 0),
    reserved_shares INTEGER NOT NULL DEFAULT 0 CHECK (reserved_shares >= 0),
    bound_api_key_id BIGINT REFERENCES api_keys(id) ON DELETE SET NULL,
    activated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT cafe_round_memberships_share_counts_check CHECK (paid_shares >= 0 AND reserved_shares >= 0),
    CONSTRAINT cafe_round_memberships_time_check CHECK (activated_at IS NULL OR expires_at IS NULL OR activated_at < expires_at),
    UNIQUE (round_id, user_id)
);

-- Keep historical Cafe seat/key/binding rows untouched, but give every
-- historical Cafe participant a durable aggregate record for read-side and
-- audit compatibility. A second run is harmless because of the unique key.
INSERT INTO cafe_round_memberships (round_id, user_id, status, paid_shares, reserved_shares, bound_api_key_id, activated_at, expires_at)
SELECT
    s.round_id,
    s.user_id,
    CASE WHEN BOOL_OR(s.status = 'active') THEN 'active'
         WHEN BOOL_OR(s.status = 'paid') THEN 'paid'
         ELSE 'locked' END,
    COALESCE(SUM(s.share_count) FILTER (WHERE s.status IN ('paid', 'active', 'refund_pending', 'refund_processing', 'refunded')), 0),
    COALESCE(SUM(s.share_count) FILTER (WHERE s.status = 'locked'), 0),
    MAX(s.bound_api_key_id),
    MIN(s.activated_at),
    MAX(s.expires_at)
FROM group_buy_seats s
JOIN group_buy_rounds r ON r.id = s.round_id AND r.cafe_room_id IS NOT NULL
WHERE s.status IN ('locked', 'paid', 'active', 'refund_pending', 'refund_processing', 'refunded')
GROUP BY s.round_id, s.user_id
ON CONFLICT (round_id, user_id) DO NOTHING;

ALTER TABLE group_buy_seats
    ADD COLUMN IF NOT EXISTS membership_id BIGINT REFERENCES cafe_round_memberships(id) ON DELETE SET NULL;

UPDATE group_buy_seats s
SET membership_id = m.id
FROM cafe_round_memberships m, group_buy_rounds r
WHERE s.membership_id IS NULL
  AND m.round_id = s.round_id
  AND m.user_id = s.user_id
  AND r.id = s.round_id
  AND r.cafe_room_id IS NOT NULL;

ALTER TABLE api_key_account_bindings
    ADD COLUMN IF NOT EXISTS membership_id BIGINT REFERENCES cafe_round_memberships(id) ON DELETE CASCADE;

ALTER TABLE api_key_account_bindings
    ALTER COLUMN seat_id DROP NOT NULL;

ALTER TABLE api_key_account_bindings
    DROP CONSTRAINT IF EXISTS api_key_account_bindings_subject_check;

ALTER TABLE api_key_account_bindings
    ADD CONSTRAINT api_key_account_bindings_subject_check CHECK ((seat_id IS NOT NULL) <> (membership_id IS NOT NULL));

DROP INDEX IF EXISTS idx_group_buy_rounds_cafe_room_live;
CREATE UNIQUE INDEX IF NOT EXISTS idx_group_buy_rounds_cafe_room_live
    ON group_buy_rounds(cafe_room_id)
    WHERE cafe_room_id IS NOT NULL AND status IN ('open', 'awaiting_account', 'activating', 'active');

CREATE INDEX IF NOT EXISTS idx_group_buy_rounds_cafe_fulfillment_deadline
    ON group_buy_rounds(fulfillment_deadline_at)
    WHERE cafe_room_id IS NOT NULL AND status = 'awaiting_account';

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_buy_rounds_assigned_account_live
    ON group_buy_rounds(assigned_account_id)
    WHERE assigned_account_id IS NOT NULL AND status IN ('activating', 'active');

CREATE INDEX IF NOT EXISTS idx_cafe_round_memberships_round_status
    ON cafe_round_memberships(round_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cafe_round_memberships_bound_key
    ON cafe_round_memberships(bound_api_key_id)
    WHERE bound_api_key_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_cafe_round_memberships_expires
    ON cafe_round_memberships(expires_at)
    WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_group_buy_seats_membership
    ON group_buy_seats(membership_id)
    WHERE membership_id IS NOT NULL;

DROP INDEX IF EXISTS idx_api_key_account_bindings_active_seat;
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_key_account_bindings_active_seat
    ON api_key_account_bindings(seat_id)
    WHERE status = 'active' AND seat_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_api_key_account_bindings_active_membership
    ON api_key_account_bindings(membership_id)
    WHERE status = 'active' AND membership_id IS NOT NULL;
