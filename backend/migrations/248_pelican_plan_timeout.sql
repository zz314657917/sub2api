ALTER TABLE pelican_test_plans
  ADD COLUMN IF NOT EXISTS timeout_seconds INTEGER NOT NULL DEFAULT 180
  CHECK (timeout_seconds BETWEEN 30 AND 3600);
