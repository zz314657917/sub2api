INSERT INTO settings (key, value, updated_at)
VALUES
    ('welfare_daily_checkin_min_account_age_hours', '24', NOW())
ON CONFLICT (key) DO NOTHING;
