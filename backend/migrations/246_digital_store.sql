-- Digital goods store.  This migration is deliberately additive: the Ent
-- schema is not regenerated because store operations use the existing
-- transaction client's ExecContext/QueryContext API.

CREATE TABLE IF NOT EXISTS store_files (
    id BIGSERIAL PRIMARY KEY,
    filename TEXT NOT NULL,
    encrypted_content TEXT NOT NULL,
    content_size BIGINT NOT NULL CHECK (content_size >= 0 AND content_size <= 20971520),
    content_sha256 CHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT store_files_sha256_format CHECK (content_sha256 ~ '^[0-9a-f]{64}$')
);

CREATE TABLE IF NOT EXISTS store_products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    description VARCHAR(4000) NOT NULL DEFAULT '',
    kind VARCHAR(16) NOT NULL CHECK (kind IN ('card', 'account', 'file')),
    price_cents INTEGER NOT NULL CHECK (price_cents > 0 AND price_cents <= 100000000),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    file_id BIGINT NULL REFERENCES store_files(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT store_products_file_kind CHECK ((kind = 'file' AND file_id IS NOT NULL) OR (kind <> 'file' AND file_id IS NULL))
);

CREATE TABLE IF NOT EXISTS store_inventory (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES store_products(id) ON DELETE RESTRICT,
    kind VARCHAR(16) NOT NULL CHECK (kind IN ('card', 'account')),
    encrypted_content TEXT NOT NULL,
    state VARCHAR(16) NOT NULL DEFAULT 'available' CHECK (state IN ('available', 'reserved', 'consumed')),
    reservation_order_id BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reserved_at TIMESTAMPTZ NULL,
    consumed_at TIMESTAMPTZ NULL,
    CONSTRAINT store_inventory_reservation_state CHECK ((state = 'available' AND reservation_order_id IS NULL) OR (state IN ('reserved', 'consumed') AND reservation_order_id IS NOT NULL))
);
CREATE INDEX IF NOT EXISTS idx_store_inventory_available ON store_inventory(product_id, id) WHERE state = 'available';
CREATE UNIQUE INDEX IF NOT EXISTS uq_store_inventory_reservation ON store_inventory(reservation_order_id) WHERE reservation_order_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS store_orders (
    id BIGSERIAL PRIMARY KEY,
    payment_order_id BIGINT NOT NULL UNIQUE REFERENCES payment_orders(id) ON DELETE RESTRICT,
    user_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL REFERENCES store_products(id) ON DELETE RESTRICT,
    idempotency_key VARCHAR(64) NOT NULL,
    product_name VARCHAR(120) NOT NULL,
    kind VARCHAR(16) NOT NULL CHECK (kind IN ('card', 'account', 'file')),
    price_cents INTEGER NOT NULL CHECK (price_cents > 0 AND price_cents <= 100000000),
    file_id BIGINT NULL REFERENCES store_files(id) ON DELETE RESTRICT,
    inventory_id BIGINT NULL REFERENCES store_inventory(id) ON DELETE RESTRICT,
    delivery_status VARCHAR(24) NOT NULL DEFAULT 'reserved' CHECK (delivery_status IN ('reserved', 'delivered', 'needs_attention', 'released')),
    encrypted_delivery TEXT NULL,
    encrypted_checkout_response TEXT NULL,
    delivered_at TIMESTAMPTZ NULL,
    released_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT store_orders_delivery_source CHECK ((kind = 'file' AND file_id IS NOT NULL AND inventory_id IS NULL) OR (kind IN ('card', 'account') AND inventory_id IS NOT NULL AND file_id IS NULL)),
    CONSTRAINT uq_store_orders_user_idempotency UNIQUE (user_id, idempotency_key)
);
CREATE INDEX IF NOT EXISTS idx_store_orders_user_created ON store_orders(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_store_orders_product_created ON store_orders(product_id, created_at DESC);

-- The FK cannot be declared inline because store_inventory is created before
-- store_orders.  PostgreSQL has no ADD CONSTRAINT IF NOT EXISTS.
DO $$ BEGIN
    ALTER TABLE store_inventory ADD CONSTRAINT fk_store_inventory_reservation_order
      FOREIGN KEY (reservation_order_id) REFERENCES store_orders(id) ON DELETE RESTRICT;
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
