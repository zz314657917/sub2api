ALTER TABLE group_buy_plans
    ADD COLUMN IF NOT EXISTS product_key VARCHAR(64) NOT NULL DEFAULT 'token_pinpinpin',
    ADD COLUMN IF NOT EXISTS total_shares INTEGER,
    ADD COLUMN IF NOT EXISTS price_per_share DECIMAL(20, 2),
    ADD COLUMN IF NOT EXISTS quota_per_share_label VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS max_shares_per_user INTEGER NOT NULL DEFAULT 10,
    ADD COLUMN IF NOT EXISTS tier_group_ids JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS launch_mode VARCHAR(16) NOT NULL DEFAULT 'auto';

UPDATE group_buy_plans
SET
    total_shares = COALESCE(total_shares, LEAST(GREATEST(seat_count, 1), 10)),
    price_per_share = COALESCE(price_per_share, price_per_seat),
    quota_per_share_label = CASE
        WHEN quota_per_share_label = '' THEN quota_label
        ELSE quota_per_share_label
    END,
    max_shares_per_user = LEAST(GREATEST(max_shares_per_user, 1), 10),
    tier_group_ids = CASE
        WHEN tier_group_ids IS NULL OR tier_group_ids = '{}'::jsonb THEN jsonb_build_object(
            '1', target_group_id,
            '2', target_group_id,
            '3', target_group_id,
            '4', target_group_id,
            '5', target_group_id,
            '6', target_group_id,
            '7', target_group_id,
            '8', target_group_id,
            '9', target_group_id,
            '10', target_group_id
        )
        ELSE tier_group_ids
    END;

ALTER TABLE group_buy_plans
    ALTER COLUMN product_key SET NOT NULL,
    ALTER COLUMN product_key SET DEFAULT 'token_pinpinpin',
    ALTER COLUMN total_shares SET NOT NULL,
    ALTER COLUMN total_shares SET DEFAULT 10,
    ALTER COLUMN price_per_share SET NOT NULL,
    ALTER COLUMN quota_per_share_label SET DEFAULT '',
    ALTER COLUMN quota_per_share_label SET NOT NULL,
    ALTER COLUMN max_shares_per_user SET DEFAULT 10,
    ALTER COLUMN max_shares_per_user SET NOT NULL,
    ALTER COLUMN tier_group_ids SET DEFAULT '{}'::jsonb,
    ALTER COLUMN tier_group_ids SET NOT NULL,
    ALTER COLUMN launch_mode SET DEFAULT 'auto',
    ALTER COLUMN launch_mode SET NOT NULL;

ALTER TABLE group_buy_plans
    DROP CONSTRAINT IF EXISTS group_buy_plans_total_shares_check,
    DROP CONSTRAINT IF EXISTS group_buy_plans_seat_count_check,
    DROP CONSTRAINT IF EXISTS group_buy_plans_price_per_share_check,
    DROP CONSTRAINT IF EXISTS group_buy_plans_price_per_seat_check,
    DROP CONSTRAINT IF EXISTS group_buy_plans_max_shares_per_user_check,
    DROP CONSTRAINT IF EXISTS group_buy_plans_launch_mode_check;

ALTER TABLE group_buy_plans
    ADD CONSTRAINT group_buy_plans_total_shares_check CHECK (total_shares > 0 AND total_shares <= 10),
    ADD CONSTRAINT group_buy_plans_seat_count_check CHECK (seat_count > 0),
    ADD CONSTRAINT group_buy_plans_price_per_share_check CHECK (price_per_share > 0),
    ADD CONSTRAINT group_buy_plans_price_per_seat_check CHECK (price_per_seat > 0),
    ADD CONSTRAINT group_buy_plans_max_shares_per_user_check CHECK (max_shares_per_user > 0 AND max_shares_per_user <= 10),
    ADD CONSTRAINT group_buy_plans_launch_mode_check CHECK (launch_mode IN ('auto', 'manual'));

COMMENT ON TABLE group_buy_plans IS 'TokenPinPinPin share-car group-buying plan templates.';
COMMENT ON COLUMN group_buy_plans.tier_group_ids IS 'JSON mapping from share count 1..10 to active subscription group id.';
COMMENT ON COLUMN group_buy_plans.target_group_id IS 'Compatibility target group, usually the 10-share tier group.';
COMMENT ON COLUMN group_buy_plans.quota_per_share_label IS 'Display-only quota copy; enforcement comes from tier target group limits.';

CREATE INDEX IF NOT EXISTS idx_group_buy_plans_product_status_sort
    ON group_buy_plans(product_key, status, sort_order, id)
    WHERE deleted_at IS NULL;

ALTER TABLE group_buy_rounds
    ADD COLUMN IF NOT EXISTS total_shares INTEGER,
    ADD COLUMN IF NOT EXISTS paid_shares INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reserved_shares INTEGER NOT NULL DEFAULT 0;

UPDATE group_buy_rounds
SET
    total_shares = COALESCE(total_shares, LEAST(GREATEST(total_seats, 1), 10)),
    paid_shares = COALESCE(paid_shares, paid_seats),
    reserved_shares = COALESCE(reserved_shares, reserved_seats);

ALTER TABLE group_buy_rounds
    ALTER COLUMN total_shares SET NOT NULL,
    ALTER COLUMN paid_shares SET DEFAULT 0,
    ALTER COLUMN reserved_shares SET DEFAULT 0;

ALTER TABLE group_buy_rounds
    DROP CONSTRAINT IF EXISTS group_buy_rounds_total_shares_check,
    DROP CONSTRAINT IF EXISTS group_buy_rounds_total_seats_check,
    DROP CONSTRAINT IF EXISTS group_buy_rounds_paid_shares_check,
    DROP CONSTRAINT IF EXISTS group_buy_rounds_reserved_shares_check,
    DROP CONSTRAINT IF EXISTS group_buy_rounds_share_counts_check;

ALTER TABLE group_buy_rounds
    ADD CONSTRAINT group_buy_rounds_total_shares_check CHECK (total_shares > 0 AND total_shares <= 10),
    ADD CONSTRAINT group_buy_rounds_total_seats_check CHECK (total_seats > 0),
    ADD CONSTRAINT group_buy_rounds_paid_shares_check CHECK (paid_shares >= 0),
    ADD CONSTRAINT group_buy_rounds_reserved_shares_check CHECK (reserved_shares >= 0),
    ADD CONSTRAINT group_buy_rounds_share_counts_check CHECK (paid_shares + reserved_shares <= total_shares);

ALTER TABLE group_buy_seats
    ADD COLUMN IF NOT EXISTS share_count INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;

ALTER TABLE group_buy_seats
    DROP CONSTRAINT IF EXISTS group_buy_seats_share_count_check;

ALTER TABLE group_buy_seats
    ADD CONSTRAINT group_buy_seats_share_count_check CHECK (share_count > 0 AND share_count <= 10);

DROP INDEX IF EXISTS idx_group_buy_seats_round_user_active;

CREATE INDEX IF NOT EXISTS idx_group_buy_seats_status_expires_at
    ON group_buy_seats(status, expires_at);

CREATE INDEX IF NOT EXISTS idx_group_buy_seats_round_user_status
    ON group_buy_seats(round_id, user_id, status);

CREATE TABLE IF NOT EXISTS group_buy_entitlements (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_key VARCHAR(64) NOT NULL DEFAULT 'token_pinpinpin',
    status VARCHAR(20) NOT NULL DEFAULT 'inactive' CHECK (status IN ('active', 'inactive')),
    active_share_count INTEGER NOT NULL DEFAULT 0 CHECK (active_share_count >= 0 AND active_share_count <= 10),
    target_group_id BIGINT REFERENCES groups(id) ON DELETE SET NULL,
    subscription_id BIGINT REFERENCES user_subscriptions(id) ON DELETE SET NULL,
    bound_api_key_id BIGINT REFERENCES api_keys(id) ON DELETE SET NULL,
    last_activated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    refreshed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deactivated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT group_buy_entitlements_user_product_unique UNIQUE (user_id, product_key)
);

CREATE INDEX IF NOT EXISTS idx_group_buy_entitlements_status
    ON group_buy_entitlements(status);

CREATE INDEX IF NOT EXISTS idx_group_buy_entitlements_target_group
    ON group_buy_entitlements(target_group_id);

CREATE INDEX IF NOT EXISTS idx_group_buy_entitlements_subscription
    ON group_buy_entitlements(subscription_id);

CREATE INDEX IF NOT EXISTS idx_group_buy_entitlements_bound_key
    ON group_buy_entitlements(bound_api_key_id);

CREATE INDEX IF NOT EXISTS idx_group_buy_entitlements_expires_at
    ON group_buy_entitlements(expires_at);
