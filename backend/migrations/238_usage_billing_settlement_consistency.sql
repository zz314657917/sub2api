-- 238_usage_billing_settlement_consistency.sql
-- Persist the outcome of post-response billing independently from the usage cost.

ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS billing_status VARCHAR(16) NOT NULL DEFAULT 'applied',
    ADD COLUMN IF NOT EXISTS billing_error TEXT;

ALTER TABLE usage_logs
    DROP CONSTRAINT IF EXISTS usage_logs_billing_status_check;
ALTER TABLE usage_logs
    ADD CONSTRAINT usage_logs_billing_status_check
    CHECK (billing_status IN ('pending', 'failed', 'applied'));

CREATE INDEX IF NOT EXISTS idx_usage_logs_billing_status_created_at
    ON usage_logs (billing_status, created_at);

CREATE TABLE IF NOT EXISTS usage_billing_settlement_outbox (
    id BIGSERIAL PRIMARY KEY,
    usage_log_id BIGINT NOT NULL REFERENCES usage_logs(id) ON DELETE CASCADE,
    request_id VARCHAR(255) NOT NULL,
    api_key_id BIGINT NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    lease_until TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('pending', 'processing', 'failed', 'applied')),
    UNIQUE (usage_log_id)
);

CREATE INDEX IF NOT EXISTS idx_usage_billing_settlement_outbox_claim
    ON usage_billing_settlement_outbox (status, available_at, lease_until, id);

CREATE UNIQUE INDEX IF NOT EXISTS usage_billing_settlement_outbox_request_api_key_unique
    ON usage_billing_settlement_outbox (request_id, api_key_id);
