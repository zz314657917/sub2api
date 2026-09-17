ALTER TABLE pelican_test_plans
  ADD COLUMN IF NOT EXISTS daily_call_limit INTEGER NOT NULL DEFAULT 0 CHECK (daily_call_limit BETWEEN 0 AND 100000),
  ADD COLUMN IF NOT EXISTS failure_pause_threshold INTEGER NOT NULL DEFAULT 0 CHECK (failure_pause_threshold BETWEEN 0 AND 100),
  ADD COLUMN IF NOT EXISTS retention_days INTEGER NOT NULL DEFAULT 0 CHECK (retention_days BETWEEN 0 AND 3650),
  ADD COLUMN IF NOT EXISTS usage_day DATE,
  ADD COLUMN IF NOT EXISTS daily_calls_used INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS consecutive_failed_runs INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS pause_reason TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS last_run_calls INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS round_attempted INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS round_successes INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_pelican_test_results_plan_account_started
  ON pelican_test_results (plan_id, account_id, group_id, started_at DESC, id DESC);
