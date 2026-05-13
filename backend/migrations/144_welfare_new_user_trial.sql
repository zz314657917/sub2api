CREATE TABLE IF NOT EXISTS welfare_new_user_trials (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quota_amount DECIMAL(20, 8) NOT NULL DEFAULT 0.1 CHECK (quota_amount >= 0),
    quota_used DECIMAL(20, 8) NOT NULL DEFAULT 0 CHECK (quota_used >= 0),
    status VARCHAR(32) NOT NULL DEFAULT 'available',
    activated_ip TEXT,
    first_started_at TIMESTAMPTZ,
    first_success_at TIMESTAMPTZ,
    last_request_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id)
);

CREATE INDEX IF NOT EXISTS idx_welfare_new_user_trials_status
    ON welfare_new_user_trials(status);

CREATE INDEX IF NOT EXISTS idx_welfare_new_user_trials_activated_ip_started
    ON welfare_new_user_trials(activated_ip, first_started_at);

CREATE TABLE IF NOT EXISTS welfare_new_user_trial_usages (
    id BIGSERIAL PRIMARY KEY,
    trial_id BIGINT NOT NULL REFERENCES welfare_new_user_trials(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    request_id TEXT NOT NULL,
    amount DECIMAL(20, 8) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    model TEXT,
    api_key_id BIGINT REFERENCES api_keys(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (request_id)
);

CREATE INDEX IF NOT EXISTS idx_welfare_new_user_trial_usages_trial_id
    ON welfare_new_user_trial_usages(trial_id);

CREATE INDEX IF NOT EXISTS idx_welfare_new_user_trial_usages_user_id
    ON welfare_new_user_trial_usages(user_id);

CREATE INDEX IF NOT EXISTS idx_welfare_new_user_trial_usages_created_at
    ON welfare_new_user_trial_usages(created_at);

INSERT INTO settings (key, value, updated_at)
VALUES
    ('welfare_new_user_trial_enabled', 'false', NOW()),
    ('welfare_new_user_trial_quota_amount', '0.1', NOW()),
    ('welfare_new_user_trial_daily_site_quota_amount', '5', NOW()),
    ('welfare_new_user_trial_daily_ip_activation_limit', '3', NOW())
ON CONFLICT (key) DO NOTHING;
