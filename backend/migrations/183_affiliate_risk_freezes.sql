INSERT INTO settings (key, value, updated_at)
VALUES ('affiliate_risk_scan_interval_minutes', '20', NOW())
ON CONFLICT (key) DO NOTHING;

CREATE TABLE IF NOT EXISTS affiliate_risk_freezes (
    id BIGSERIAL PRIMARY KEY,
    inviter_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fingerprint VARCHAR(128) NOT NULL,
    severity VARCHAR(16) NOT NULL,
    score INTEGER NOT NULL,
    reason TEXT NOT NULL,
    source_window_start TIMESTAMPTZ NOT NULL,
    source_window_end TIMESTAMPTZ NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    cleared_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE affiliate_risk_freezes IS 'Affiliate reward monetization freezes created by risk scanner.';
COMMENT ON COLUMN affiliate_risk_freezes.fingerprint IS 'Stable risk-cluster fingerprint used to deduplicate scanner output.';
COMMENT ON COLUMN affiliate_risk_freezes.active IS 'Active freezes block affiliate reward claim and transfer.';

CREATE UNIQUE INDEX IF NOT EXISTS idx_affiliate_risk_freezes_active_fingerprint
    ON affiliate_risk_freezes (inviter_id, fingerprint)
    WHERE active = true;

CREATE INDEX IF NOT EXISTS idx_affiliate_risk_freezes_inviter_active
    ON affiliate_risk_freezes (inviter_id, active, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_users_created_at
    ON users (created_at);

CREATE INDEX IF NOT EXISTS idx_user_affiliate_ledger_action_created_at
    ON user_affiliate_ledger (action, created_at);

CREATE INDEX IF NOT EXISTS idx_usage_logs_ip_created_at
    ON usage_logs (ip_address, created_at)
    WHERE ip_address <> '';
