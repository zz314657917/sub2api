ALTER TABLE IF EXISTS leaderboard_daily_reward_claims
    ADD COLUMN IF NOT EXISTS reward_mode TEXT NOT NULL DEFAULT 'red_packet';

ALTER TABLE IF EXISTS leaderboard_daily_reward_claims
    ADD COLUMN IF NOT EXISTS packet_id BIGINT;

ALTER TABLE IF EXISTS leaderboard_daily_reward_claims
    ADD COLUMN IF NOT EXISTS lottery_run_id BIGINT;

ALTER TABLE IF EXISTS leaderboard_daily_reward_claims
    DROP CONSTRAINT IF EXISTS leaderboard_daily_reward_claims_rank_check;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'leaderboard_daily_reward_claims_rank_top10_check'
    ) THEN
        ALTER TABLE leaderboard_daily_reward_claims
            ADD CONSTRAINT leaderboard_daily_reward_claims_rank_top10_check
            CHECK (rank BETWEEN 1 AND 10);
    END IF;
END $$;

ALTER TABLE IF EXISTS leaderboard_daily_reward_claims
    DROP CONSTRAINT IF EXISTS leaderboard_daily_reward_claims_reward_date_user_id_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_leaderboard_daily_reward_claims_period_mode_user
    ON leaderboard_daily_reward_claims(reward_date, reward_mode, user_id);

CREATE INDEX IF NOT EXISTS idx_leaderboard_daily_reward_claims_period_mode
    ON leaderboard_daily_reward_claims(reward_date, reward_mode);

CREATE TABLE IF NOT EXISTS leaderboard_red_packets (
    id BIGSERIAL PRIMARY KEY,
    reward_date TEXT NOT NULL,
    packet_no INTEGER NOT NULL CHECK (packet_no BETWEEN 1 AND 10),
    amount DECIMAL(20, 8) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    claimed_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    claim_id BIGINT REFERENCES leaderboard_daily_reward_claims(id) ON DELETE SET NULL,
    claimed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (reward_date, packet_no)
);

ALTER TABLE IF EXISTS leaderboard_red_packets
    DROP CONSTRAINT IF EXISTS leaderboard_red_packets_claimed_by_unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx_leaderboard_red_packets_claim_id
    ON leaderboard_red_packets(claim_id)
    WHERE claim_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_leaderboard_red_packets_claimed_by
    ON leaderboard_red_packets(reward_date, claimed_by)
    WHERE claimed_by IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_leaderboard_red_packets_reward_date
    ON leaderboard_red_packets(reward_date);

CREATE TABLE IF NOT EXISTS leaderboard_lottery_runs (
    id BIGSERIAL PRIMARY KEY,
    reward_date TEXT NOT NULL UNIQUE,
    winner_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    winner_rank INTEGER NOT NULL CHECK (winner_rank BETWEEN 1 AND 10),
    amount DECIMAL(20, 8) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    total_actual_cost DECIMAL(20, 8) NOT NULL DEFAULT 0 CHECK (total_actual_cost >= 0),
    redeem_code_id BIGINT REFERENCES redeem_codes(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_leaderboard_lottery_runs_winner_user
    ON leaderboard_lottery_runs(winner_user_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'leaderboard_daily_reward_claims_packet_fk'
    ) THEN
        ALTER TABLE leaderboard_daily_reward_claims
            ADD CONSTRAINT leaderboard_daily_reward_claims_packet_fk
            FOREIGN KEY (packet_id)
            REFERENCES leaderboard_red_packets(id)
            ON DELETE SET NULL
            NOT VALID;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'leaderboard_daily_reward_claims_lottery_run_fk'
    ) THEN
        ALTER TABLE leaderboard_daily_reward_claims
            ADD CONSTRAINT leaderboard_daily_reward_claims_lottery_run_fk
            FOREIGN KEY (lottery_run_id)
            REFERENCES leaderboard_lottery_runs(id)
            ON DELETE SET NULL
            NOT VALID;
    END IF;
END $$;

INSERT INTO settings (key, value, updated_at)
VALUES (
    'leaderboard_daily_reward_mode',
    COALESCE(
        (
            SELECT CASE WHEN value = 'true' THEN 'red_packet' ELSE 'disabled' END
            FROM settings
            WHERE key = 'leaderboard_daily_reward_enabled'
        ),
        'disabled'
    ),
    NOW()
)
ON CONFLICT (key) DO NOTHING;

INSERT INTO settings (key, value, updated_at) VALUES
    ('leaderboard_daily_reward_red_packet_pool_amount', '0', NOW()),
    ('leaderboard_daily_reward_red_packet_min_amount', '0', NOW()),
    ('leaderboard_daily_reward_red_packet_max_amount', '0', NOW()),
    ('leaderboard_daily_reward_lottery_amount', '0', NOW()),
    ('leaderboard_daily_reward_lottery_cron', '0 12 * * 4', NOW())
ON CONFLICT (key) DO NOTHING;
