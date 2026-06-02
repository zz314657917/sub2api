CREATE TABLE IF NOT EXISTS support_tickets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    ticket_type VARCHAR(20) NOT NULL DEFAULT 'support',
    system_key VARCHAR(80),
    last_message_preview VARCHAR(240) NOT NULL DEFAULT '',
    last_message_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_unread_count INTEGER NOT NULL DEFAULT 0,
    admin_unread_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    CONSTRAINT chk_support_tickets_status CHECK (
        status IN ('open', 'pending_admin', 'pending_user', 'closed')
    ),
    CONSTRAINT chk_support_tickets_type CHECK (
        ticket_type IN ('support', 'system')
    ),
    CONSTRAINT chk_support_tickets_user_unread_nonnegative CHECK (user_unread_count >= 0),
    CONSTRAINT chk_support_tickets_admin_unread_nonnegative CHECK (admin_unread_count >= 0)
);

CREATE TABLE IF NOT EXISTS support_ticket_messages (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
    sender_type VARCHAR(20) NOT NULL,
    sender_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    event_type VARCHAR(80),
    event_key VARCHAR(160),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_support_ticket_messages_sender_type CHECK (
        sender_type IN ('user', 'admin', 'system')
    )
);

CREATE INDEX IF NOT EXISTS idx_support_tickets_user_last_message
    ON support_tickets(user_id, last_message_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_support_tickets_status_last_message
    ON support_tickets(status, last_message_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_support_tickets_type_user_last_message
    ON support_tickets(ticket_type, user_id, last_message_at DESC, id DESC);

CREATE UNIQUE INDEX IF NOT EXISTS ux_support_tickets_user_system_key
    ON support_tickets(user_id, system_key)
    WHERE ticket_type = 'system' AND system_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_support_ticket_messages_ticket_created
    ON support_ticket_messages(ticket_id, created_at ASC, id ASC);

CREATE UNIQUE INDEX IF NOT EXISTS ux_support_ticket_messages_ticket_event_key
    ON support_ticket_messages(ticket_id, event_key)
    WHERE event_key IS NOT NULL AND event_key <> '';

CREATE INDEX IF NOT EXISTS idx_support_ticket_messages_sender_user
    ON support_ticket_messages(sender_user_id);
