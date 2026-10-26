-- Empty preferences retain automatic scheduling for existing and new plans.
ALTER TABLE pelican_test_plans
    ADD COLUMN IF NOT EXISTS preferred_account_ids BIGINT[] NOT NULL DEFAULT '{}';
