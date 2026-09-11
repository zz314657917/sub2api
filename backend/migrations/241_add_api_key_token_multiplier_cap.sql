ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS token_multiplier_cap DECIMAL(20,8) NOT NULL DEFAULT 0;
COMMENT ON COLUMN api_keys.token_multiplier_cap IS 'Maximum token billing multiplier for this API key (0 = unlimited)';
