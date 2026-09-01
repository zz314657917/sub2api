-- Historical billing reconciliation. Review the SELECT result first; setting
-- approved=true is an explicit operator action and only then creates retry jobs.

WITH candidates AS (
    SELECT ul.id AS usage_log_id, ul.request_id, ul.api_key_id
    FROM usage_logs ul
    WHERE ul.total_cost > 0
      AND ul.actual_cost = 0
      AND NOT EXISTS (
          SELECT 1 FROM billing_usage_entries bue WHERE bue.usage_log_id = ul.id
      )
)
SELECT * FROM candidates ORDER BY usage_log_id;

-- Operators must replace the VALUES list after review. Keep the transaction
-- open until the payloads have been checked: this does not deduct funds
-- directly, it only marks rows pending and inserts idempotent outbox jobs.
-- The approved amount and command snapshot are intentionally operator supplied;
-- the historical row has actual_cost = 0 and cannot safely reconstruct the
-- original account/subscription/API-key split without a human decision.
--
-- BEGIN;
-- CREATE TEMP TABLE usage_billing_reconcile_approved (
--     usage_log_id BIGINT PRIMARY KEY,
--     request_id VARCHAR(255) NOT NULL,
--     api_key_id BIGINT NOT NULL,
--     payload JSONB NOT NULL
-- );
-- INSERT INTO usage_billing_reconcile_approved (usage_log_id, request_id, api_key_id, payload)
-- VALUES
--     (<usage_log_id>, '<request_id>', <api_key_id>, '<validated UsageBillingSettlementPayload JSON>'::jsonb);
--
-- UPDATE usage_logs ul
-- SET billing_status = 'pending', billing_error = 'historical reconciliation approved'
-- FROM usage_billing_reconcile_approved a
-- WHERE ul.id = a.usage_log_id
--   AND ul.total_cost > 0
--   AND ul.actual_cost = 0
--   AND NOT EXISTS (SELECT 1 FROM billing_usage_entries b WHERE b.usage_log_id = ul.id);
--
-- INSERT INTO usage_billing_settlement_outbox
--     (usage_log_id, request_id, api_key_id, payload, status, available_at)
-- SELECT a.usage_log_id, a.request_id, a.api_key_id, a.payload, 'pending', NOW()
-- FROM usage_billing_reconcile_approved a
-- JOIN usage_logs ul ON ul.id = a.usage_log_id
-- WHERE ul.billing_status = 'pending'
-- ON CONFLICT (usage_log_id) DO NOTHING;
-- COMMIT;
