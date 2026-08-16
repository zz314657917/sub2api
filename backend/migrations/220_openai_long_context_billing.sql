-- Keep the account veto and its usage audit trail consistent across mixed
-- application versions. This checkout has no persisted Spark shadow schema;
-- account-level normalization is therefore intentionally limited to accounts.
ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS long_context_billing_applied BOOLEAN NOT NULL DEFAULT FALSE;

CREATE OR REPLACE FUNCTION public.enforce_openai_long_context_billing_extra()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.platform IS DISTINCT FROM 'openai' THEN
        RETURN NEW;
    END IF;

    NEW.extra := COALESCE(NEW.extra, '{}'::jsonb);
    IF NOT (NEW.extra ? 'openai_long_context_billing_enabled')
        AND TG_OP = 'UPDATE'
        AND OLD.platform = 'openai'
        AND jsonb_typeof(COALESCE(OLD.extra, '{}'::jsonb)->'openai_long_context_billing_enabled') = 'boolean' THEN
        NEW.extra := jsonb_set(
            NEW.extra,
            '{openai_long_context_billing_enabled}',
            OLD.extra->'openai_long_context_billing_enabled',
            true
        );
    ELSIF NOT (NEW.extra ? 'openai_long_context_billing_enabled') THEN
        NEW.extra := jsonb_set(
            NEW.extra,
            '{openai_long_context_billing_enabled}',
            'false'::jsonb,
            true
        );
    END IF;

    IF jsonb_typeof(NEW.extra->'openai_long_context_billing_enabled') IS DISTINCT FROM 'boolean' THEN
        RAISE EXCEPTION 'openai_long_context_billing_enabled must be a boolean'
            USING ERRCODE = '22023';
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS accounts_enforce_openai_long_context_billing_extra ON accounts;

-- Normalize legacy/malformed values before later application versions rely on
-- the strict trigger. Missing values default off so upgrades never increase a
-- customer's charge without an explicit opt-in.
UPDATE accounts
SET extra = jsonb_set(
    COALESCE(extra, '{}'::jsonb),
    '{openai_long_context_billing_enabled}',
    'false'::jsonb,
    true
)
WHERE platform = 'openai'
  AND (
    NOT (COALESCE(extra, '{}'::jsonb) ? 'openai_long_context_billing_enabled')
    OR jsonb_typeof(extra->'openai_long_context_billing_enabled') IS DISTINCT FROM 'boolean'
  );

CREATE TRIGGER accounts_enforce_openai_long_context_billing_extra
BEFORE INSERT OR UPDATE OF platform, extra
ON accounts
FOR EACH ROW
EXECUTE FUNCTION public.enforce_openai_long_context_billing_extra();
