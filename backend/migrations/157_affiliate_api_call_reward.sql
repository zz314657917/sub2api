INSERT INTO settings (key, value, updated_at)
VALUES ('affiliate_api_call_reward_amount', '0', NOW())
ON CONFLICT (key) DO NOTHING;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_affiliate_ledger_api_call_reward_once
    ON user_affiliate_ledger(user_id, source_user_id)
    WHERE action = 'api_call_reward' AND source_user_id IS NOT NULL;
