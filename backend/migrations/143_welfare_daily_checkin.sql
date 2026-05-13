CREATE TABLE IF NOT EXISTS welfare_daily_checkins (
    id BIGSERIAL PRIMARY KEY,
    checkin_date DATE NOT NULL,
    reward_month CHAR(7) NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(20, 8) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    redeem_code_id BIGINT REFERENCES redeem_codes(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (checkin_date, user_id)
);

CREATE INDEX IF NOT EXISTS idx_welfare_daily_checkins_user_id
    ON welfare_daily_checkins(user_id);

CREATE INDEX IF NOT EXISTS idx_welfare_daily_checkins_reward_month
    ON welfare_daily_checkins(reward_month);

CREATE TABLE IF NOT EXISTS welfare_daily_checkin_milestone_claims (
    id BIGSERIAL PRIMARY KEY,
    reward_month CHAR(7) NOT NULL,
    milestone_day INTEGER NOT NULL CHECK (milestone_day IN (7, 14, 21, 28)),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount DECIMAL(20, 8) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    redeem_code_id BIGINT REFERENCES redeem_codes(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (reward_month, milestone_day, user_id)
);

CREATE INDEX IF NOT EXISTS idx_welfare_daily_checkin_milestone_claims_user_id
    ON welfare_daily_checkin_milestone_claims(user_id);

CREATE INDEX IF NOT EXISTS idx_welfare_daily_checkin_milestone_claims_reward_month
    ON welfare_daily_checkin_milestone_claims(reward_month);

INSERT INTO settings (key, value, updated_at)
VALUES
    ('welfare_enabled', 'false', NOW()),
    ('welfare_daily_checkin_enabled', 'false', NOW()),
    ('welfare_recharge_enabled', 'false', NOW()),
    ('welfare_vip_enabled', 'false', NOW()),
    ('welfare_daily_checkin_reward_min', '0', NOW()),
    ('welfare_daily_checkin_reward_max', '0', NOW()),
    ('welfare_daily_checkin_milestone_7_amount', '0', NOW()),
    ('welfare_daily_checkin_milestone_14_amount', '0', NOW()),
    ('welfare_daily_checkin_milestone_21_amount', '0', NOW()),
    ('welfare_daily_checkin_milestone_28_amount', '0', NOW())
ON CONFLICT (key) DO NOTHING;
