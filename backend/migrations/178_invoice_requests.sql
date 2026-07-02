CREATE TABLE IF NOT EXISTS invoice_requests (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  amount DECIMAL(20,2) NOT NULL,
  currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
  invoice_type VARCHAR(30) NOT NULL,
  title VARCHAR(255) NOT NULL,
  tax_number VARCHAR(64) NOT NULL,
  remark TEXT,
  status VARCHAR(30) NOT NULL DEFAULT 'pending',
  admin_note TEXT,
  invoice_no VARCHAR(128),
  file_name VARCHAR(255),
  file_path TEXT,
  file_size BIGINT,
  file_content_type VARCHAR(128),
  reviewed_by BIGINT,
  reviewed_at TIMESTAMPTZ,
  issued_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS invoice_requests_user_id_idx ON invoice_requests(user_id);
CREATE INDEX IF NOT EXISTS invoice_requests_status_idx ON invoice_requests(status);
CREATE INDEX IF NOT EXISTS invoice_requests_created_at_idx ON invoice_requests(created_at);
CREATE INDEX IF NOT EXISTS invoice_requests_currency_idx ON invoice_requests(currency);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'invoice_requests_invoice_type_check'
  ) THEN
    ALTER TABLE invoice_requests
      ADD CONSTRAINT invoice_requests_invoice_type_check
      CHECK (invoice_type IN ('vat_general', 'vat_special')) NOT VALID;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'invoice_requests_status_check'
  ) THEN
    ALTER TABLE invoice_requests
      ADD CONSTRAINT invoice_requests_status_check
      CHECK (status IN ('pending', 'approved', 'rejected', 'issued')) NOT VALID;
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'invoice_requests_amount_positive_check'
  ) THEN
    ALTER TABLE invoice_requests
      ADD CONSTRAINT invoice_requests_amount_positive_check
      CHECK (amount > 0) NOT VALID;
  END IF;
END $$;
