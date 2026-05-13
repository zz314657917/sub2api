-- Daily per-user usage rollups for leaderboard badges.
CREATE TABLE IF NOT EXISTS user_usage_daily_stats (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    usage_date DATE NOT NULL,
    requests BIGINT NOT NULL DEFAULT 0,
    tokens BIGINT NOT NULL DEFAULT 0,
    actual_cost NUMERIC(20, 10) NOT NULL DEFAULT 0,
    night_requests BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, usage_date)
);

CREATE INDEX IF NOT EXISTS idx_user_usage_daily_stats_usage_date
    ON user_usage_daily_stats (usage_date);

CREATE INDEX IF NOT EXISTS idx_user_usage_daily_stats_tokens
    ON user_usage_daily_stats (tokens DESC);

CREATE INDEX IF NOT EXISTS idx_user_usage_daily_stats_night_requests
    ON user_usage_daily_stats (night_requests DESC);

INSERT INTO user_usage_daily_stats (
    user_id,
    usage_date,
    requests,
    tokens,
    actual_cost,
    night_requests,
    updated_at
)
SELECT
    user_id,
    (created_at AT TIME ZONE current_setting('TimeZone'))::date AS usage_date,
    COUNT(*) AS requests,
    COALESCE(SUM(input_tokens + output_tokens + cache_creation_tokens + cache_read_tokens), 0) AS tokens,
    COALESCE(SUM(actual_cost), 0) AS actual_cost,
    COUNT(*) FILTER (
        WHERE EXTRACT(HOUR FROM created_at AT TIME ZONE current_setting('TimeZone')) >= 0
          AND EXTRACT(HOUR FROM created_at AT TIME ZONE current_setting('TimeZone')) < 6
    ) AS night_requests,
    NOW() AS updated_at
FROM usage_logs
WHERE user_id IS NOT NULL
GROUP BY user_id, (created_at AT TIME ZONE current_setting('TimeZone'))::date
ON CONFLICT (user_id, usage_date) DO UPDATE SET
    requests = EXCLUDED.requests,
    tokens = EXCLUDED.tokens,
    actual_cost = EXCLUDED.actual_cost,
    night_requests = EXCLUDED.night_requests,
    updated_at = NOW();
