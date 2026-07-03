-- Welfare check-in voucher wallet.
-- Vouchers are logically expired by expires_at checks on user-scoped reads.

CREATE TABLE IF NOT EXISTS welfare_vouchers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_type VARCHAR(64) NOT NULL,
    source_id BIGINT NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    remaining_amount DECIMAL(20, 8) NOT NULL,
    expires_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    redeem_code_id BIGINT REFERENCES redeem_codes(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT welfare_vouchers_amount_positive CHECK (amount > 0),
    CONSTRAINT welfare_vouchers_remaining_nonnegative CHECK (remaining_amount >= 0),
    CONSTRAINT welfare_vouchers_remaining_not_over_amount CHECK (remaining_amount <= amount),
    CONSTRAINT welfare_vouchers_status_check CHECK (status IN ('active', 'depleted', 'expired'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_welfare_vouchers_source
    ON welfare_vouchers(source_type, source_id);

CREATE INDEX IF NOT EXISTS idx_welfare_vouchers_user_available
    ON welfare_vouchers(user_id, expires_at, id)
    WHERE status = 'active' AND remaining_amount > 0;

CREATE TABLE IF NOT EXISTS welfare_voucher_ledger (
    id BIGSERIAL PRIMARY KEY,
    voucher_id BIGINT REFERENCES welfare_vouchers(id) ON DELETE SET NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    operation VARCHAR(20) NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    remaining_after DECIMAL(20, 8),
    operation_type VARCHAR(64),
    operation_key VARCHAR(255),
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT welfare_voucher_ledger_amount_positive CHECK (amount > 0),
    CONSTRAINT welfare_voucher_ledger_operation_check CHECK (operation IN ('grant', 'consume', 'refund'))
);

CREATE INDEX IF NOT EXISTS idx_welfare_voucher_ledger_user_created
    ON welfare_voucher_ledger(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_welfare_voucher_ledger_operation
    ON welfare_voucher_ledger(operation_type, operation_key)
    WHERE operation_type IS NOT NULL AND operation_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS welfare_voucher_deductions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    voucher_id BIGINT NOT NULL REFERENCES welfare_vouchers(id) ON DELETE CASCADE,
    operation_type VARCHAR(64) NOT NULL,
    operation_key VARCHAR(255) NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    refunded_amount DECIMAL(20, 8) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT welfare_voucher_deductions_amount_positive CHECK (amount > 0),
    CONSTRAINT welfare_voucher_deductions_refunded_nonnegative CHECK (refunded_amount >= 0),
    CONSTRAINT welfare_voucher_deductions_refunded_not_over_amount CHECK (refunded_amount <= amount)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_welfare_voucher_deductions_operation_voucher
    ON welfare_voucher_deductions(operation_type, operation_key, voucher_id);

CREATE INDEX IF NOT EXISTS idx_welfare_voucher_deductions_user_operation
    ON welfare_voucher_deductions(user_id, operation_type, operation_key);

INSERT INTO settings (key, value, updated_at) VALUES
    ('welfare_voucher_valid_days', '0', NOW())
ON CONFLICT (key) DO NOTHING;
