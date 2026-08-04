-- Pixel Cafe schema foundation. This migration is additive and forward-only.

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS access_mode VARCHAR(20) NOT NULL DEFAULT 'normal';

ALTER TABLE group_buy_plans
    ADD COLUMN IF NOT EXISTS fulfillment_mode VARCHAR(32) NOT NULL DEFAULT 'aggregate_tier',
    ADD COLUMN IF NOT EXISTS room_key_quota_usd DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS room_key_rate_limit_5h DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS room_key_rate_limit_1d DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS room_key_rate_limit_7d DECIMAL(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS auto_create_room_key BOOLEAN NOT NULL DEFAULT TRUE;

CREATE TABLE IF NOT EXISTS cafe_rooms (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL,
    name VARCHAR(120) NOT NULL,
    plan_id BIGINT NOT NULL REFERENCES group_buy_plans(id) ON DELETE RESTRICT,
    account_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    zone_key VARCHAR(32) NOT NULL DEFAULT 'featured',
    theme_key VARCHAR(64) NOT NULL DEFAULT 'warm_wood',
    scene_slot_key VARCHAR(120) NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'enabled', 'maintenance', 'disabled')),
    featured BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

ALTER TABLE group_buy_rounds
    ADD COLUMN IF NOT EXISTS cafe_room_id BIGINT REFERENCES cafe_rooms(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS assigned_account_id BIGINT REFERENCES accounts(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS room_code_snapshot VARCHAR(64),
    ADD COLUMN IF NOT EXISTS room_name_snapshot VARCHAR(120),
    ADD COLUMN IF NOT EXISTS activated_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS entitlement_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS activation_token VARCHAR(120);

ALTER TABLE group_buy_seats
    ADD COLUMN IF NOT EXISTS seat_no INTEGER;

ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS managed_source_type VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS managed_source_id BIGINT;

CREATE TABLE IF NOT EXISTS api_key_account_bindings (
    id BIGSERIAL PRIMARY KEY,
    api_key_id BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    cafe_room_id BIGINT NOT NULL REFERENCES cafe_rooms(id) ON DELETE CASCADE,
    round_id BIGINT NOT NULL REFERENCES group_buy_rounds(id) ON DELETE CASCADE,
    seat_id BIGINT NOT NULL REFERENCES group_buy_seats(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'expired', 'revoked', 'migrated')),
    strict_mode BOOLEAN NOT NULL DEFAULT TRUE,
    starts_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    replaced_by_binding_id BIGINT REFERENCES api_key_account_bindings(id) ON DELETE SET NULL,
    migrated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT api_key_account_bindings_time_check CHECK (starts_at < expires_at)
);

CREATE INDEX IF NOT EXISTS idx_cafe_rooms_zone_status_sort
    ON cafe_rooms(zone_key, status, sort_order)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_cafe_rooms_code_active
    ON cafe_rooms(code)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_cafe_rooms_plan
    ON cafe_rooms(plan_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_cafe_rooms_account
    ON cafe_rooms(account_id)
    WHERE deleted_at IS NULL AND account_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_buy_rounds_cafe_room_live
    ON group_buy_rounds(cafe_room_id)
    WHERE cafe_room_id IS NOT NULL AND status IN ('open', 'activating', 'active');

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_buy_rounds_activation_token
    ON group_buy_rounds(activation_token)
    WHERE activation_token IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_buy_seats_round_seat_no_live
    ON group_buy_seats(round_id, seat_no)
    WHERE seat_no IS NOT NULL AND status IN ('locked', 'paid', 'active', 'refund_pending', 'refund_processing');

CREATE INDEX IF NOT EXISTS idx_group_buy_seats_seat_no
    ON group_buy_seats(round_id, seat_no)
    WHERE seat_no IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_managed_source_active
    ON api_keys(managed_source_type, managed_source_id)
    WHERE managed_source_type <> '' AND managed_source_id IS NOT NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_api_key_account_bindings_active_key
    ON api_key_account_bindings(api_key_id)
    WHERE status = 'active';

CREATE UNIQUE INDEX IF NOT EXISTS idx_api_key_account_bindings_active_seat
    ON api_key_account_bindings(seat_id)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_api_key_account_bindings_user_group_status
    ON api_key_account_bindings(user_id, group_id, status);

CREATE INDEX IF NOT EXISTS idx_api_key_account_bindings_account_status
    ON api_key_account_bindings(account_id, status);

CREATE INDEX IF NOT EXISTS idx_api_key_account_bindings_room_status
    ON api_key_account_bindings(cafe_room_id, status);

CREATE INDEX IF NOT EXISTS idx_api_key_account_bindings_round
    ON api_key_account_bindings(round_id);

CREATE INDEX IF NOT EXISTS idx_api_key_account_bindings_expires_at
    ON api_key_account_bindings(expires_at);
