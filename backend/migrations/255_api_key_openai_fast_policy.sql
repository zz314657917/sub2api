ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS openai_fast_policy VARCHAR(16) NOT NULL DEFAULT 'passthrough';

ALTER TABLE api_keys
    ADD CONSTRAINT api_keys_openai_fast_policy_check
    CHECK (openai_fast_policy IN ('passthrough', 'force', 'filter'));
