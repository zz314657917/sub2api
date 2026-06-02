ALTER TABLE support_tickets
    ADD COLUMN IF NOT EXISTS ticket_type VARCHAR(20) NOT NULL DEFAULT 'support';

ALTER TABLE support_tickets
    ADD COLUMN IF NOT EXISTS system_key VARCHAR(80);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_support_tickets_type'
    ) THEN
        ALTER TABLE support_tickets
            ADD CONSTRAINT chk_support_tickets_type CHECK (ticket_type IN ('support', 'system'));
    END IF;
END $$;

ALTER TABLE support_ticket_messages
    ADD COLUMN IF NOT EXISTS event_type VARCHAR(80);

ALTER TABLE support_ticket_messages
    ADD COLUMN IF NOT EXISTS event_key VARCHAR(160);

ALTER TABLE support_ticket_messages
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE INDEX IF NOT EXISTS idx_support_tickets_type_user_last_message
    ON support_tickets(ticket_type, user_id, last_message_at DESC, id DESC);

CREATE UNIQUE INDEX IF NOT EXISTS ux_support_tickets_user_system_key
    ON support_tickets(user_id, system_key)
    WHERE ticket_type = 'system' AND system_key IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_support_ticket_messages_ticket_event_key
    ON support_ticket_messages(ticket_id, event_key)
    WHERE event_key IS NOT NULL AND event_key <> '';
