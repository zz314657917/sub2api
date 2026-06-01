-- Store a one-way payment method fingerprint for self-referral checks.
ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS payment_method_fingerprint VARCHAR(128) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_payment_orders_method_fingerprint_user
    ON payment_orders (payment_method_fingerprint, user_id)
    WHERE payment_method_fingerprint <> '';

ALTER TABLE user_affiliates
    ADD COLUMN IF NOT EXISTS affiliate_revoked_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS affiliate_revoked_reason VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS affiliate_revoked_order_id BIGINT NULL REFERENCES payment_orders(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_user_affiliates_revoked_at
    ON user_affiliates (affiliate_revoked_at)
    WHERE affiliate_revoked_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_user_affiliates_active_inviter
    ON user_affiliates (inviter_id)
    WHERE inviter_id IS NOT NULL AND affiliate_revoked_at IS NULL;

COMMENT ON COLUMN payment_orders.payment_method_fingerprint IS 'One-way hash of stable payer/payment method identifier, used to detect affiliate self-referrals.';
COMMENT ON COLUMN user_affiliates.affiliate_revoked_at IS 'Time when inviter binding was revoked by anti-abuse checks.';
COMMENT ON COLUMN user_affiliates.affiliate_revoked_reason IS 'Machine-readable reason for affiliate revocation.';
COMMENT ON COLUMN user_affiliates.affiliate_revoked_order_id IS 'Payment order that triggered affiliate revocation, when available.';
