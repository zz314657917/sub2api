CREATE UNIQUE INDEX IF NOT EXISTS idx_user_affiliate_ledger_first_recharge_reward_once
    ON user_affiliate_ledger(user_id, source_user_id)
    WHERE action = 'first_recharge_reward' AND source_user_id IS NOT NULL;

COMMENT ON INDEX idx_user_affiliate_ledger_first_recharge_reward_once IS 'Ensures an inviter receives the fixed first-recharge reward at most once per invitee.';
