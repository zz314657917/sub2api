CREATE TABLE IF NOT EXISTS image_creator_tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    model VARCHAR(100) NOT NULL DEFAULT 'gpt-image-2',
    prompt TEXT NOT NULL,
    size VARCHAR(40) NOT NULL DEFAULT '',
    quality VARCHAR(40) NOT NULL DEFAULT '',
    output_format VARCHAR(20) NOT NULL DEFAULT 'png',
    background VARCHAR(40) NOT NULL DEFAULT '',
    image_count INTEGER NOT NULL DEFAULT 1,
    reference_image_path TEXT,
    reference_image_mime_type VARCHAR(100),
    reference_image_filename TEXT,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_image_creator_tasks_status
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed')),
    CONSTRAINT chk_image_creator_tasks_count
        CHECK (image_count BETWEEN 1 AND 4)
);

CREATE TABLE IF NOT EXISTS image_creator_images (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES image_creator_tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    output_format VARCHAR(20) NOT NULL DEFAULT 'png',
    mime_type VARCHAR(100) NOT NULL DEFAULT 'image/png',
    byte_size BIGINT NOT NULL DEFAULT 0,
    sha256 VARCHAR(64) NOT NULL DEFAULT '',
    revised_prompt TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_image_creator_tasks_user_created
    ON image_creator_tasks(user_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_image_creator_tasks_status_created
    ON image_creator_tasks(status, created_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_image_creator_tasks_expires_at
    ON image_creator_tasks(expires_at);

CREATE INDEX IF NOT EXISTS idx_image_creator_images_user_created
    ON image_creator_images(user_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_image_creator_images_task_created
    ON image_creator_images(task_id, created_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_image_creator_images_expires_at
    ON image_creator_images(expires_at);
