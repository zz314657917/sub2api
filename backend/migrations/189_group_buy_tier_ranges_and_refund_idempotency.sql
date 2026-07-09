ALTER TABLE group_buy_plans
    ADD COLUMN IF NOT EXISTS tier_rules JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE group_buy_plans
SET tier_rules = COALESCE((
    WITH expanded AS (
        SELECT
            share::int AS share_count,
            (tier_group_ids ->> share)::bigint AS group_id
        FROM generate_series(1, 10) AS share
        WHERE tier_group_ids ? share::text
          AND NULLIF(tier_group_ids ->> share, '') IS NOT NULL
          AND (tier_group_ids ->> share) ~ '^[0-9]+$'
    ),
    islands AS (
        SELECT
            share_count,
            group_id,
            share_count - row_number() OVER (PARTITION BY group_id ORDER BY share_count) AS island_key
        FROM expanded
    ),
    ranges AS (
        SELECT
            min(share_count) AS min_shares,
            max(share_count) AS max_shares,
            group_id
        FROM islands
        GROUP BY group_id, island_key
    )
    SELECT jsonb_agg(jsonb_build_object(
        'min_shares', min_shares,
        'max_shares', max_shares,
        'target_group_id', group_id,
        'label', min_shares::text || CASE WHEN min_shares = max_shares THEN '份' ELSE '-' || max_shares::text || '份' END
    ) ORDER BY min_shares)
    FROM ranges
), jsonb_build_array(jsonb_build_object(
    'min_shares', 1,
    'max_shares', LEAST(GREATEST(total_shares, 1), 10),
    'target_group_id', target_group_id,
    'label', '默认档位'
)))
WHERE tier_rules = '[]'::jsonb;

COMMENT ON COLUMN group_buy_plans.tier_rules IS 'Inclusive share range rules mapping purchased active shares to subscription groups.';
COMMENT ON COLUMN group_buy_plans.tier_group_ids IS 'Legacy JSON mapping from exact share count to subscription group id; new code uses tier_rules.';

ALTER TABLE group_buy_seats
    ADD COLUMN IF NOT EXISTS policy_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN group_buy_seats.policy_snapshot IS 'Snapshot of group-buy entitlement policy captured when the share batch is locked.';

ALTER TABLE group_buy_seats
    DROP CONSTRAINT IF EXISTS group_buy_seats_status_check;

ALTER TABLE group_buy_seats
    ADD CONSTRAINT group_buy_seats_status_check CHECK (status IN ('locked', 'released', 'paid', 'active', 'refund_pending', 'refund_processing', 'refunded', 'cancelled'));

ALTER TABLE group_buy_entitlements
    ADD COLUMN IF NOT EXISTS managed_subscription_id BIGINT REFERENCES user_subscriptions(id) ON DELETE SET NULL;

ALTER TABLE user_subscriptions
    ADD COLUMN IF NOT EXISTS source_type VARCHAR(64) NOT NULL DEFAULT 'standard',
    ADD COLUMN IF NOT EXISTS source_id BIGINT,
    ADD COLUMN IF NOT EXISTS managed_by_group_buy BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE user_subscriptions
SET source_type = 'standard'
WHERE source_type IS NULL OR source_type = '';

COMMENT ON COLUMN user_subscriptions.source_type IS 'Subscription source namespace. group_buy is isolated from standard user subscriptions.';
COMMENT ON COLUMN user_subscriptions.source_id IS 'Source record id within source_type, such as group_buy entitlement id.';
COMMENT ON COLUMN user_subscriptions.managed_by_group_buy IS 'Whether this subscription is created and rotated by group-buy entitlement sync.';

UPDATE group_buy_entitlements
SET managed_subscription_id = subscription_id
WHERE managed_subscription_id IS NULL
  AND subscription_id IS NOT NULL;

UPDATE user_subscriptions us
SET source_type = 'group_buy',
    source_id = gbe.id,
    managed_by_group_buy = TRUE
FROM group_buy_entitlements gbe
WHERE us.id = gbe.managed_subscription_id
  AND (us.source_type = 'standard' OR us.source_type IS NULL);

DROP INDEX IF EXISTS user_subscriptions_user_group_unique_active;

CREATE UNIQUE INDEX IF NOT EXISTS user_subscriptions_user_group_source_unique_active
    ON user_subscriptions(user_id, group_id, source_type)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_group_buy_entitlements_managed_subscription
    ON group_buy_entitlements(managed_subscription_id);

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_source
    ON user_subscriptions(source_type, source_id);

CREATE INDEX IF NOT EXISTS idx_user_subscriptions_managed_by_group_buy
    ON user_subscriptions(managed_by_group_buy);

CREATE TABLE IF NOT EXISTS group_buy_refunds (
    id BIGSERIAL PRIMARY KEY,
    seat_id BIGINT NOT NULL REFERENCES group_buy_seats(id) ON DELETE CASCADE,
    order_id BIGINT REFERENCES payment_orders(id) ON DELETE SET NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mode VARCHAR(32) NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'processing' CHECK (status IN ('processing', 'succeeded', 'pending_provider', 'failed')),
    amount DECIMAL(20, 2) NOT NULL CHECK (amount >= 0),
    idempotency_key VARCHAR(120) NOT NULL,
    note TEXT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_buy_refunds_seat
    ON group_buy_refunds(seat_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_buy_refunds_idempotency
    ON group_buy_refunds(idempotency_key);

CREATE INDEX IF NOT EXISTS idx_group_buy_refunds_order
    ON group_buy_refunds(order_id);

CREATE INDEX IF NOT EXISTS idx_group_buy_refunds_status
    ON group_buy_refunds(status);

CREATE INDEX IF NOT EXISTS idx_group_buy_refunds_user
    ON group_buy_refunds(user_id);
