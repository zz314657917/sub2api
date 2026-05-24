-- Add API key account-pool scheduling strategy.

ALTER TABLE api_keys
ADD COLUMN IF NOT EXISTS account_pool_strategy VARCHAR(32) NOT NULL DEFAULT 'shared_only';

COMMENT ON COLUMN api_keys.account_pool_strategy IS 'API key account pool strategy: shared_only, private_first, private_only';

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'api_keys_account_pool_strategy_check'
  ) THEN
    ALTER TABLE api_keys
      ADD CONSTRAINT api_keys_account_pool_strategy_check
      CHECK (account_pool_strategy IN ('shared_only', 'private_first', 'private_only'));
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_api_keys_account_pool_strategy
ON api_keys (account_pool_strategy);
