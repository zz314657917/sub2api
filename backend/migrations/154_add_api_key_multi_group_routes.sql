-- Add multi-group routing configuration for API keys.
-- api_keys.group_id remains the default/legacy group binding.

ALTER TABLE api_keys
ADD COLUMN IF NOT EXISTS multi_group_routes JSONB NOT NULL DEFAULT '[]'::jsonb;

COMMENT ON COLUMN api_keys.multi_group_routes IS 'API key multi-group routes: [{"group_id":1,"priority":100,"weight":1,"cooldown_seconds":30,"enabled":true}]';

CREATE INDEX IF NOT EXISTS idx_api_keys_multi_group_routes_gin
ON api_keys USING GIN (multi_group_routes);
