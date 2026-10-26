-- Additive V1 channel monitoring: legacy clients retain probe/chat/zero jitter.
ALTER TABLE channel_monitors DROP CONSTRAINT IF EXISTS channel_monitors_provider_check;
ALTER TABLE channel_monitors ADD CONSTRAINT channel_monitors_provider_check
    CHECK (provider IN ('openai','anthropic','gemini','grok','antigravity','kimi','zhipu','deepseek','minimax'));
ALTER TABLE channel_monitor_request_templates DROP CONSTRAINT IF EXISTS channel_monitor_request_templates_provider_check;
ALTER TABLE channel_monitor_request_templates ADD CONSTRAINT channel_monitor_request_templates_provider_check
    CHECK (provider IN ('openai','anthropic','gemini','grok','antigravity','kimi','zhipu','deepseek','minimax'));

ALTER TABLE channel_monitors
    ADD COLUMN IF NOT EXISTS api_mode VARCHAR(32) NOT NULL DEFAULT 'chat_completions',
    ADD COLUMN IF NOT EXISTS check_mode VARCHAR(32) NOT NULL DEFAULT 'probe',
    ADD COLUMN IF NOT EXISTS account_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS jitter_seconds INTEGER NOT NULL DEFAULT 0;
ALTER TABLE channel_monitor_request_templates
    ADD COLUMN IF NOT EXISTS api_mode VARCHAR(32) NOT NULL DEFAULT 'chat_completions';
ALTER TABLE channel_monitor_histories ADD COLUMN IF NOT EXISTS quota JSONB;

ALTER TABLE channel_monitors ADD CONSTRAINT channel_monitors_api_mode_check
    CHECK (api_mode IN ('chat_completions','responses') AND (api_mode <> 'responses' OR provider = 'openai'));
ALTER TABLE channel_monitor_request_templates ADD CONSTRAINT channel_monitor_request_templates_api_mode_check
    CHECK (api_mode IN ('chat_completions','responses') AND (api_mode <> 'responses' OR provider = 'openai'));
ALTER TABLE channel_monitors ADD CONSTRAINT channel_monitors_check_mode_check
    CHECK (check_mode IN ('probe','quota','quota_probe')
        AND (provider <> 'antigravity' OR check_mode = 'quota')
        AND (provider <> 'minimax' OR check_mode = 'probe'));
ALTER TABLE channel_monitors ADD CONSTRAINT channel_monitors_jitter_check
    CHECK (jitter_seconds >= 0 AND interval_seconds - jitter_seconds >= 15);

CREATE INDEX IF NOT EXISTS idx_channel_monitors_account_id ON channel_monitors(account_id);
CREATE INDEX IF NOT EXISTS idx_channel_monitors_provider_api_mode ON channel_monitors(provider,api_mode);
CREATE INDEX IF NOT EXISTS idx_channel_monitor_templates_provider_api_mode ON channel_monitor_request_templates(provider,api_mode);
