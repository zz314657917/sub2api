-- Persist Studio Bridge image billing metadata across reserve and commit.

ALTER TABLE studio_bridge_charges
    ADD COLUMN IF NOT EXISTS image_count INTEGER NOT NULL DEFAULT 0;

ALTER TABLE studio_bridge_charges
    ADD COLUMN IF NOT EXISTS image_size VARCHAR(10);

ALTER TABLE studio_bridge_charges
    ADD COLUMN IF NOT EXISTS image_size_source VARCHAR(16);

ALTER TABLE studio_bridge_charges
    ADD COLUMN IF NOT EXISTS image_size_breakdown JSONB;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'studio_bridge_charges_image_count_nonnegative'
          AND conrelid = 'studio_bridge_charges'::regclass
    ) THEN
        ALTER TABLE studio_bridge_charges
            ADD CONSTRAINT studio_bridge_charges_image_count_nonnegative
            CHECK (image_count >= 0) NOT VALID;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'studio_bridge_charges_image_size_source_check'
          AND conrelid = 'studio_bridge_charges'::regclass
    ) THEN
        ALTER TABLE studio_bridge_charges
            ADD CONSTRAINT studio_bridge_charges_image_size_source_check
            CHECK (
                image_size_source IS NULL
                OR image_size_source IN ('output', 'input', 'default', 'legacy')
            ) NOT VALID;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'studio_bridge_charges_image_size_check'
          AND conrelid = 'studio_bridge_charges'::regclass
    ) THEN
        ALTER TABLE studio_bridge_charges
            ADD CONSTRAINT studio_bridge_charges_image_size_check
            CHECK (
                image_size IS NULL
                OR image_size IN ('1K', '2K', '4K', 'mixed')
            ) NOT VALID;
    END IF;
END $$;
