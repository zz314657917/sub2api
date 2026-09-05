-- Persist structured policy explanations without retaining prompt or credential data.
ALTER TABLE prompt_audit_events
    ADD COLUMN IF NOT EXISTS matched_rule_id VARCHAR(128) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS owasp_tags JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE prompt_audit_events
    DROP CONSTRAINT IF EXISTS chk_prompt_audit_events_owasp_tags_json;

ALTER TABLE prompt_audit_events
    ADD CONSTRAINT chk_prompt_audit_events_owasp_tags_json
    CHECK (jsonb_typeof(owasp_tags) = 'array');

CREATE INDEX IF NOT EXISTS idx_prompt_audit_events_matched_rule
    ON prompt_audit_events(matched_rule_id)
    WHERE matched_rule_id <> '';
