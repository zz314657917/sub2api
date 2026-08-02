-- Compatibility column for the admin detail contract. Runtime writes keep
-- this field redacted; unredacted prompt text remains in the short-lived Redis
-- payload only and is never persisted to PostgreSQL.
ALTER TABLE prompt_audit_events
    ADD COLUMN IF NOT EXISTS full_prompt TEXT NOT NULL DEFAULT '';
