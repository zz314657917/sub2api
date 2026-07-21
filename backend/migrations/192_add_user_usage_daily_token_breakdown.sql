-- Preserve token breakdowns in the daily rollup used by long-period leaderboards.
ALTER TABLE user_usage_daily_stats
    ADD COLUMN IF NOT EXISTS input_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS output_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cache_creation_tokens BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cache_read_tokens BIGINT NOT NULL DEFAULT 0;

WITH usage_breakdown AS (
    SELECT
        user_id,
        (created_at AT TIME ZONE current_setting('TimeZone'))::date AS usage_date,
        COALESCE(SUM(input_tokens), 0) AS input_tokens,
        COALESCE(SUM(output_tokens), 0) AS output_tokens,
        COALESCE(SUM(cache_creation_tokens), 0) AS cache_creation_tokens,
        COALESCE(SUM(cache_read_tokens), 0) AS cache_read_tokens
    FROM usage_logs
    WHERE user_id IS NOT NULL
    GROUP BY user_id, (created_at AT TIME ZONE current_setting('TimeZone'))::date
)
UPDATE user_usage_daily_stats AS stats
SET input_tokens = usage_breakdown.input_tokens,
    output_tokens = usage_breakdown.output_tokens,
    cache_creation_tokens = usage_breakdown.cache_creation_tokens,
    cache_read_tokens = usage_breakdown.cache_read_tokens,
    updated_at = NOW()
FROM usage_breakdown
WHERE stats.user_id = usage_breakdown.user_id
  AND stats.usage_date = usage_breakdown.usage_date;
