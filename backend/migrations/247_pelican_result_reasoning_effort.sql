-- NULL preserves the fact that older results did not record their effort.
ALTER TABLE pelican_test_results
  ADD COLUMN IF NOT EXISTS reasoning_effort TEXT
  CHECK (reasoning_effort IN ('', 'none', 'minimal', 'low', 'medium', 'high', 'xhigh'));
