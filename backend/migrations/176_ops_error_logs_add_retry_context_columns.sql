-- Backfill retry-context columns for deployments that created ops_error_logs
-- before the squashed Ops vNext schema included request body capture.
ALTER TABLE ops_error_logs
  ADD COLUMN IF NOT EXISTS request_body JSONB,
  ADD COLUMN IF NOT EXISTS request_headers JSONB,
  ADD COLUMN IF NOT EXISTS request_body_truncated BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS request_body_bytes INT,
  ADD COLUMN IF NOT EXISTS is_retryable BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0;

COMMENT ON COLUMN ops_error_logs.request_body IS 'Sanitized client request body stored for retryable error requests.';
COMMENT ON COLUMN ops_error_logs.request_headers IS 'Sanitized client request headers stored for retryable error requests.';
COMMENT ON COLUMN ops_error_logs.request_body_truncated IS 'Whether request_body was truncated before persistence.';
COMMENT ON COLUMN ops_error_logs.request_body_bytes IS 'Original request body size in bytes when captured.';
COMMENT ON COLUMN ops_error_logs.is_retryable IS 'Best-effort retryability classification for this error request.';
COMMENT ON COLUMN ops_error_logs.retry_count IS 'Number of retry attempts recorded for this error request.';
