-- S91: administrator-owned model matching rules for API-key group routing.
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS model_match_patterns JSONB NOT NULL DEFAULT '[]'::jsonb;

COMMENT ON COLUMN groups.model_match_patterns IS
    'Administrator-maintained request model patterns; * means all models';
