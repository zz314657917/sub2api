-- Forward-only repair for the reservation-first Cafe lifecycle.
-- Keep all legacy states; do not rewrite rounds, reservations or payment data.
-- The migration runner executes this file in a single transaction.
ALTER TABLE group_buy_rounds
    DROP CONSTRAINT IF EXISTS group_buy_rounds_status_check,
    ADD CONSTRAINT group_buy_rounds_status_check CHECK (status IN (
        'open', 'reserving', 'awaiting_payment', 'awaiting_account',
        'activating', 'active', 'completed', 'refunding', 'refunded',
        'failed', 'cancelled'
    ));
