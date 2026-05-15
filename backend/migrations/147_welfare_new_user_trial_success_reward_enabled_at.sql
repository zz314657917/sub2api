INSERT INTO settings (key, value, updated_at)
VALUES
    ('welfare_new_user_trial_success_reward_enabled_at', '', NOW())
ON CONFLICT (key) DO NOTHING;
