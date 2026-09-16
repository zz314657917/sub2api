CREATE TABLE IF NOT EXISTS pelican_test_plans (
 id BIGSERIAL PRIMARY KEY, group_id BIGINT NOT NULL, group_name TEXT NOT NULL DEFAULT '', model_id TEXT NOT NULL,
 interval_minutes INTEGER NOT NULL CHECK (interval_minutes BETWEEN 15 AND 1440), enabled BOOLEAN NOT NULL DEFAULT FALSE,
 max_results INTEGER NOT NULL DEFAULT 20 CHECK (max_results BETWEEN 1 AND 50), min_chars INTEGER NOT NULL DEFAULT 9366 CHECK (min_chars BETWEEN 100 AND 50000),
 last_run_at TIMESTAMPTZ, next_run_at TIMESTAMPTZ, running_until TIMESTAMPTZ, run_generation BIGINT NOT NULL DEFAULT 0, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_pelican_test_plans_due ON pelican_test_plans (next_run_at) WHERE enabled;
CREATE TABLE IF NOT EXISTS pelican_test_results (
 id BIGSERIAL PRIMARY KEY, plan_id BIGINT NOT NULL REFERENCES pelican_test_plans(id) ON DELETE CASCADE, group_id BIGINT NOT NULL, account_id BIGINT NOT NULL,
 model_id TEXT NOT NULL, prompt_version TEXT NOT NULL, status TEXT NOT NULL CHECK (status IN ('success','failed','skipped')),
 error_message TEXT NOT NULL DEFAULT '', latency_ms BIGINT NOT NULL DEFAULT 0, char_count INTEGER NOT NULL DEFAULT 0, min_chars INTEGER NOT NULL,
 run_generation BIGINT NOT NULL DEFAULT 0, started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), finished_at TIMESTAMPTZ, html TEXT
);
CREATE INDEX IF NOT EXISTS idx_pelican_test_results_group_finished ON pelican_test_results (group_id, finished_at DESC);
CREATE INDEX IF NOT EXISTS idx_pelican_test_results_plan_account_finished ON pelican_test_results (plan_id, account_id, finished_at DESC);
