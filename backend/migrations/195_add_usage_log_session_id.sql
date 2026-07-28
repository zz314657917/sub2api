-- Persist only explicit client-provided session/conversation identifiers for
-- usage correlation. The column remains nullable for older requests.
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS session_id VARCHAR(255);
