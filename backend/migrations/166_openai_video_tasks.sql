CREATE TABLE IF NOT EXISTS openai_video_tasks (
	id BIGSERIAL PRIMARY KEY,
	task_id VARCHAR(255) NOT NULL,
	provider VARCHAR(64) NOT NULL DEFAULT 'openai',
	user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	api_key_id BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
	group_id BIGINT REFERENCES groups(id) ON DELETE SET NULL,
	account_id BIGINT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
	model TEXT NOT NULL,
	billing_model TEXT,
	upstream_model TEXT,
	channel_id BIGINT REFERENCES channels(id) ON DELETE SET NULL,
	original_model TEXT,
	channel_mapped_model TEXT,
	billing_model_source VARCHAR(32),
	model_mapping_chain TEXT,
	status VARCHAR(32) NOT NULL DEFAULT 'submitted',
	billing_status VARCHAR(32) NOT NULL DEFAULT 'pending',
	estimated_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
	reserved_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
	refunded_cost DOUBLE PRECISION NOT NULL DEFAULT 0,
	usage_log_id BIGINT REFERENCES usage_logs(id) ON DELETE SET NULL,
	request_payload_hash VARCHAR(64),
	submit_response JSONB,
	last_status_response JSONB,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	completed_at TIMESTAMPTZ,
	billed_at TIMESTAMPTZ,
	UNIQUE (provider, task_id)
);

CREATE INDEX IF NOT EXISTS idx_openai_video_tasks_user_id ON openai_video_tasks(user_id);
CREATE INDEX IF NOT EXISTS idx_openai_video_tasks_api_key_id ON openai_video_tasks(api_key_id);
CREATE INDEX IF NOT EXISTS idx_openai_video_tasks_account_id ON openai_video_tasks(account_id);
CREATE INDEX IF NOT EXISTS idx_openai_video_tasks_billing_status ON openai_video_tasks(billing_status);
CREATE INDEX IF NOT EXISTS idx_openai_video_tasks_created_at ON openai_video_tasks(created_at);
