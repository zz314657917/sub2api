CREATE TABLE IF NOT EXISTS user_membership_grants (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier VARCHAR(20) NOT NULL,
    source VARCHAR(30) NOT NULL DEFAULT 'auto_monthly_spend',
    period_key VARCHAR(7) NOT NULL,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    qualified_amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    starts_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    subscription_id BIGINT NULL REFERENCES user_subscriptions(id) ON DELETE SET NULL,
    subscription_group_id BIGINT NULL REFERENCES groups(id) ON DELETE SET NULL,
    source_order_id BIGINT NULL REFERENCES payment_orders(id) ON DELETE SET NULL,
    revoked_at TIMESTAMPTZ NULL,
    revoke_reason TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_membership_grants_period_tier
    ON user_membership_grants(user_id, source, period_key, tier);

CREATE INDEX IF NOT EXISTS idx_user_membership_grants_active
    ON user_membership_grants(user_id, status, expires_at);

CREATE INDEX IF NOT EXISTS idx_user_membership_grants_subscription_id
    ON user_membership_grants(subscription_id)
    WHERE subscription_id IS NOT NULL;
