CREATE TABLE IF NOT EXISTS group_buy_plans (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(120) NOT NULL,
    description TEXT,
    seat_count INTEGER NOT NULL CHECK (seat_count > 0),
    price_per_seat DECIMAL(20, 2) NOT NULL CHECK (price_per_seat > 0),
    quota_label VARCHAR(255) NOT NULL DEFAULT '',
    target_group_id BIGINT NOT NULL REFERENCES groups(id),
    validity_days INTEGER NOT NULL DEFAULT 30 CHECK (validity_days > 0),
    timeout_minutes INTEGER NOT NULL DEFAULT 1440 CHECK (timeout_minutes > 0),
    refund_mode VARCHAR(32) NOT NULL DEFAULT 'balance_credit' CHECK (refund_mode IN ('balance_credit', 'provider_refund')),
    agreement_text TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    last_round_created_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE group_buy_plans IS 'Platform-managed Pro capacity group-buying plan templates.';
COMMENT ON COLUMN group_buy_plans.target_group_id IS 'Subscription-type group that enforces the real per-seat quota.';
COMMENT ON COLUMN group_buy_plans.quota_label IS 'Display-only quota copy; enforcement comes from target_group_id limits.';

CREATE INDEX IF NOT EXISTS idx_group_buy_plans_status_sort
    ON group_buy_plans(status, sort_order, id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_group_buy_plans_target_group
    ON group_buy_plans(target_group_id);

CREATE TABLE IF NOT EXISTS group_buy_rounds (
    id BIGSERIAL PRIMARY KEY,
    plan_id BIGINT NOT NULL REFERENCES group_buy_plans(id),
    status VARCHAR(20) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'activating', 'active', 'failed', 'cancelled')),
    total_seats INTEGER NOT NULL CHECK (total_seats > 0),
    paid_seats INTEGER NOT NULL DEFAULT 0 CHECK (paid_seats >= 0),
    reserved_seats INTEGER NOT NULL DEFAULT 0 CHECK (reserved_seats >= 0),
    deadline_at TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    close_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT group_buy_rounds_seat_counts_check CHECK (paid_seats + reserved_seats <= total_seats)
);

CREATE INDEX IF NOT EXISTS idx_group_buy_rounds_plan_status
    ON group_buy_rounds(plan_id, status, deadline_at);

CREATE INDEX IF NOT EXISTS idx_group_buy_rounds_status_deadline
    ON group_buy_rounds(status, deadline_at);

CREATE TABLE IF NOT EXISTS group_buy_seats (
    id BIGSERIAL PRIMARY KEY,
    round_id BIGINT NOT NULL REFERENCES group_buy_rounds(id) ON DELETE CASCADE,
    plan_id BIGINT NOT NULL REFERENCES group_buy_plans(id),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    order_id BIGINT REFERENCES payment_orders(id) ON DELETE SET NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'locked' CHECK (status IN ('locked', 'released', 'paid', 'active', 'refund_pending', 'refunded', 'cancelled')),
    subscription_id BIGINT REFERENCES user_subscriptions(id) ON DELETE SET NULL,
    bound_api_key_id BIGINT REFERENCES api_keys(id) ON DELETE SET NULL,
    locked_until TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    activated_at TIMESTAMPTZ,
    bound_at TIMESTAMPTZ,
    refund_processed_at TIMESTAMPTZ,
    refund_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_buy_seats_order
    ON group_buy_seats(order_id)
    WHERE order_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_buy_seats_round_user_active
    ON group_buy_seats(round_id, user_id)
    WHERE status IN ('locked', 'paid', 'active', 'refund_pending');

CREATE INDEX IF NOT EXISTS idx_group_buy_seats_user_created
    ON group_buy_seats(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_group_buy_seats_status_locked_until
    ON group_buy_seats(status, locked_until);

CREATE INDEX IF NOT EXISTS idx_group_buy_seats_round_status
    ON group_buy_seats(round_id, status);

CREATE TABLE IF NOT EXISTS group_buy_events (
    id BIGSERIAL PRIMARY KEY,
    plan_id BIGINT REFERENCES group_buy_plans(id) ON DELETE SET NULL,
    round_id BIGINT REFERENCES group_buy_rounds(id) ON DELETE SET NULL,
    seat_id BIGINT REFERENCES group_buy_seats(id) ON DELETE SET NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    event_type VARCHAR(48) NOT NULL,
    message TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_group_buy_events_created
    ON group_buy_events(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_group_buy_events_round_created
    ON group_buy_events(round_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_group_buy_events_user_created
    ON group_buy_events(user_id, created_at DESC);
