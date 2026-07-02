ALTER TABLE proxies
  ADD COLUMN IF NOT EXISTS owner_user_id BIGINT;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'proxies_owner_user_id_fkey'
  ) THEN
    ALTER TABLE proxies
      ADD CONSTRAINT proxies_owner_user_id_fkey
      FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_proxies_owner_user_id ON proxies(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_proxies_owner_status ON proxies(owner_user_id, status);

COMMENT ON COLUMN proxies.owner_user_id IS 'User owner for self-managed proxies; NULL means admin global proxy.';
