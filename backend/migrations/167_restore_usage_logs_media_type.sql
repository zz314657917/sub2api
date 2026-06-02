-- Restore usage_logs.media_type required by usage log repository and video billing.
ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS media_type VARCHAR(16);
