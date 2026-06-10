ALTER TABLE IF EXISTS leaderboard_daily_reward_claims
    ALTER COLUMN reward_date TYPE TEXT USING reward_date::TEXT;
