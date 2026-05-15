UPDATE settings target
SET value = TO_CHAR(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
    updated_at = NOW()
FROM settings amount
WHERE target.key = 'welfare_new_user_trial_success_reward_enabled_at'
  AND target.value = ''
  AND amount.key = 'welfare_new_user_trial_success_reward_amount'
  AND TRIM(COALESCE(amount.value, '0')) ~ '^[0-9]+(\.[0-9]+)?$'
  AND TRIM(COALESCE(amount.value, '0'))::NUMERIC > 0;
