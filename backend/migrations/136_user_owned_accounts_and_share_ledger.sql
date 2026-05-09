-- User-authorized account sharing pool.

ALTER TABLE accounts ADD COLUMN IF NOT EXISTS owner_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS share_mode VARCHAR(20) NOT NULL DEFAULT 'private';
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS share_status VARCHAR(30) NOT NULL DEFAULT 'not_shared';

CREATE INDEX IF NOT EXISTS idx_accounts_owner_user_id ON accounts(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_accounts_share_mode_status ON accounts(share_mode, share_status);
CREATE INDEX IF NOT EXISTS idx_accounts_owner_share ON accounts(owner_user_id, share_mode, share_status);

CREATE TABLE IF NOT EXISTS account_share_ledger (
    id BIGSERIAL PRIMARY KEY,
    owner_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    request_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    request_id VARCHAR(255) NOT NULL,
    api_key_id BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    usage_log_id BIGINT REFERENCES usage_logs(id) ON DELETE SET NULL,
    actual_cost DECIMAL(18, 8) NOT NULL DEFAULT 0,
    owner_rate_percent DECIMAL(8, 4) NOT NULL DEFAULT 80,
    owner_amount DECIMAL(18, 8) NOT NULL DEFAULT 0,
    platform_amount DECIMAL(18, 8) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'frozen',
    freeze_until TIMESTAMPTZ NOT NULL,
    transferred_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT account_share_ledger_status_check CHECK (status IN ('frozen', 'available', 'transferred', 'cancelled'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_account_share_ledger_request_api_key
    ON account_share_ledger(request_id, api_key_id);
CREATE INDEX IF NOT EXISTS idx_account_share_ledger_owner_status
    ON account_share_ledger(owner_user_id, status, freeze_until);
CREATE INDEX IF NOT EXISTS idx_account_share_ledger_account_created
    ON account_share_ledger(account_id, created_at DESC);

INSERT INTO settings (key, value, updated_at) VALUES
    ('account_share_enabled', 'true', NOW()),
    ('account_share_owner_rate', '80', NOW()),
    ('account_share_freeze_hours', '72', NOW()),
    ('account_share_auto_review', 'false', NOW()),
    ('account_share_user_account_limit', '5', NOW())
ON CONFLICT (key) DO NOTHING;
