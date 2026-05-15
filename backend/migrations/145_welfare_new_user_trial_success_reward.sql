INSERT INTO settings (key, value, updated_at)
VALUES
    ('welfare_new_user_trial_success_reward_amount', '0', NOW())
ON CONFLICT (key) DO NOTHING;
