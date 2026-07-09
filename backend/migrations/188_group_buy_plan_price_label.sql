ALTER TABLE group_buy_plans
    ADD COLUMN IF NOT EXISTS price_label VARCHAR(120) NOT NULL DEFAULT '';

COMMENT ON COLUMN group_buy_plans.price_label IS 'Display-only price copy for TokenPinPinPin share plans; payment uses price_per_share.';
