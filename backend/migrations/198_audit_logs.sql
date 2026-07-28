-- 管理面操作审计日志（append-only）。
-- 凭证与请求体在进入本表前必须由应用层掩码/脱敏；不提供单条删除。
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor_user_id BIGINT,
    actor_email VARCHAR(255) NOT NULL DEFAULT '',
    actor_role VARCHAR(32) NOT NULL DEFAULT '',
    auth_method VARCHAR(32) NOT NULL DEFAULT '',
    credential_masked VARCHAR(160) NOT NULL DEFAULT '',
    action VARCHAR(128) NOT NULL DEFAULT '',
    method VARCHAR(16) NOT NULL DEFAULT '',
    path VARCHAR(512) NOT NULL DEFAULT '',
    request_id VARCHAR(64) NOT NULL DEFAULT '',
    client_ip VARCHAR(64) NOT NULL DEFAULT '',
    user_agent VARCHAR(512) NOT NULL DEFAULT '',
    request_body TEXT NOT NULL DEFAULT '',
    status_code INT NOT NULL DEFAULT 0,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    extra JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at_id ON audit_logs (created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_created ON audit_logs (actor_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs (action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_client_ip ON audit_logs (client_ip);
