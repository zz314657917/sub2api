ALTER TABLE pelican_test_plans
  ADD COLUMN IF NOT EXISTS reasoning_effort TEXT NOT NULL DEFAULT ''
  CHECK (reasoning_effort IN ('', 'none', 'minimal', 'low', 'medium', 'high', 'xhigh'));
