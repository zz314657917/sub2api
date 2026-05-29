ALTER TABLE image_creator_images
    ADD COLUMN IF NOT EXISTS width INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS height INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_image_creator_images_user_format_created
    ON image_creator_images(user_id, output_format, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_image_creator_images_user_dimensions
    ON image_creator_images(user_id, width, height);
