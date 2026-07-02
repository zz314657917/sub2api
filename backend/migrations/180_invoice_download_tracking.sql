-- Track whether an issued invoice file has been claimed by the user.
ALTER TABLE invoice_requests
  ADD COLUMN IF NOT EXISTS downloaded_at TIMESTAMPTZ;

ALTER TABLE invoice_requests
  ADD COLUMN IF NOT EXISTS download_count INTEGER NOT NULL DEFAULT 0;
