-- Bounded full prompt retained only for critical-risk events and returned only
-- from the administrator event-detail endpoint. Other events keep this empty.
ALTER TABLE prompt_audit_events
    ADD COLUMN IF NOT EXISTS full_prompt TEXT NOT NULL DEFAULT '';
