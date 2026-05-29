CREATE TABLE IF NOT EXISTS prompt_favorites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    prompt_id TEXT NOT NULL,
    source VARCHAR(80) NOT NULL,
    title TEXT NOT NULL,
    preview TEXT NOT NULL DEFAULT '',
    reference_image_urls JSONB NOT NULL DEFAULT '[]'::jsonb,
    prompt TEXT NOT NULL,
    author TEXT NOT NULL DEFAULT '',
    link TEXT NOT NULL DEFAULT '',
    mode VARCHAR(20) NOT NULL DEFAULT 'generate',
    category TEXT NOT NULL DEFAULT '',
    sub_category TEXT NOT NULL DEFAULT '',
    created TEXT NOT NULL DEFAULT '',
    source_label TEXT NOT NULL DEFAULT '',
    is_nsfw BOOLEAN NOT NULL DEFAULT FALSE,
    localizations JSONB NOT NULL DEFAULT '{}'::jsonb,
    favorited_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, source, prompt_id)
);

CREATE INDEX IF NOT EXISTS idx_prompt_favorites_user_time
    ON prompt_favorites(user_id, favorited_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_prompt_favorites_user_source
    ON prompt_favorites(user_id, source, favorited_at DESC, id DESC);
