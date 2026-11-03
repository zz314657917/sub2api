-- Additive administrator-owned group policy; existing groups remain unchanged.
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS force_openai_fast BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN groups.force_openai_fast IS
    'Force OpenAI priority before the global Fast policy for OpenAI and Composite groups';
