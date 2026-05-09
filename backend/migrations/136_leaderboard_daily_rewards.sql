CREATE TABLE IF NOT EXISTS leaderboard_daily_reward_claims (
    id BIGSERIAL PRIMARY KEY,
    reward_date DATE NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rank INTEGER NOT NULL CHECK (rank BETWEEN 1 AND 3),
    amount DECIMAL(20, 8) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    total_actual_cost DECIMAL(20, 8) NOT NULL DEFAULT 0 CHECK (total_actual_cost >= 0),
    redeem_code_id BIGINT REFERENCES redeem_codes(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (reward_date, user_id)
);

CREATE INDEX IF NOT EXISTS idx_leaderboard_daily_reward_claims_user_id
    ON leaderboard_daily_reward_claims(user_id);

CREATE INDEX IF NOT EXISTS idx_leaderboard_daily_reward_claims_reward_date
    ON leaderboard_daily_reward_claims(reward_date);
