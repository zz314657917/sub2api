-- Persist Studio Bridge charge idempotency in PostgreSQL.

CREATE TABLE IF NOT EXISTS studio_bridge_charges (
    id BIGSERIAL PRIMARY KEY,
    app_id VARCHAR(64) NOT NULL,
    charge_key VARCHAR(255) NOT NULL,
    refund_for_charge_key VARCHAR(255),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(20, 8) NOT NULL,
    refunded_amount DECIMAL(20, 8) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL,
    fingerprint VARCHAR(255) NOT NULL,
    reason TEXT,
    task_id VARCHAR(255),
    mode VARCHAR(64),
    model VARCHAR(255),
    actor_user_id VARCHAR(255),
    team_id VARCHAR(255),
    balance_after DECIMAL(20, 8),
    usage_logged_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT studio_bridge_charges_amount_positive CHECK (amount > 0),
    CONSTRAINT studio_bridge_charges_refunded_nonnegative CHECK (refunded_amount >= 0),
    CONSTRAINT studio_bridge_charges_refunded_not_over_amount CHECK (refunded_amount <= amount),
    CONSTRAINT studio_bridge_charges_status_check CHECK (status IN ('reserved', 'committed', 'refunded'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_studio_bridge_charges_app_key
    ON studio_bridge_charges(app_id, charge_key);

CREATE INDEX IF NOT EXISTS idx_studio_bridge_charges_user_created
    ON studio_bridge_charges(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_studio_bridge_charges_refund_for
    ON studio_bridge_charges(app_id, refund_for_charge_key)
    WHERE refund_for_charge_key IS NOT NULL AND refund_for_charge_key <> '';
