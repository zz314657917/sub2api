CREATE TABLE IF NOT EXISTS canvas_documents (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    viewport JSONB NOT NULL DEFAULT '{}'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS canvas_nodes (
    id BIGSERIAL PRIMARY KEY,
    canvas_id BIGINT NOT NULL REFERENCES canvas_documents(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    node_key VARCHAR(128) NOT NULL,
    type VARCHAR(40) NOT NULL,
    position JSONB NOT NULL DEFAULT '{}'::jsonb,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_canvas_nodes_canvas_key UNIQUE (canvas_id, node_key),
    CONSTRAINT chk_canvas_nodes_type CHECK (
        type IN ('text', 'image', 'prompt', 'loop', 'group', 'text_to_image', 'image_to_image', 'result')
    )
);

CREATE TABLE IF NOT EXISTS canvas_edges (
    id BIGSERIAL PRIMARY KEY,
    canvas_id BIGINT NOT NULL REFERENCES canvas_documents(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    edge_key VARCHAR(128) NOT NULL,
    source_node_key VARCHAR(128) NOT NULL,
    target_node_key VARCHAR(128) NOT NULL,
    source_handle VARCHAR(128),
    target_handle VARCHAR(128),
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_canvas_edges_canvas_key UNIQUE (canvas_id, edge_key)
);

CREATE TABLE IF NOT EXISTS canvas_runs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    canvas_id BIGINT REFERENCES canvas_documents(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    trigger_type VARCHAR(40) NOT NULL DEFAULT 'manual',
    api_key_id BIGINT REFERENCES api_keys(id) ON DELETE SET NULL,
    model VARCHAR(100) NOT NULL DEFAULT '',
    input JSONB NOT NULL DEFAULT '{}'::jsonb,
    output JSONB NOT NULL DEFAULT '{}'::jsonb,
    error_message TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_canvas_runs_status CHECK (
        status IN ('pending', 'running', 'succeeded', 'failed', 'canceled')
    )
);

CREATE INDEX IF NOT EXISTS idx_canvas_documents_user_updated
    ON canvas_documents(user_id, updated_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_canvas_nodes_canvas_order
    ON canvas_nodes(canvas_id, id ASC);

CREATE INDEX IF NOT EXISTS idx_canvas_edges_canvas_order
    ON canvas_edges(canvas_id, id ASC);

CREATE INDEX IF NOT EXISTS idx_canvas_runs_user_created
    ON canvas_runs(user_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_canvas_runs_canvas_created
    ON canvas_runs(canvas_id, created_at DESC, id DESC);
