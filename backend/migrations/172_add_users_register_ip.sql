-- Record the client IP observed when a user registers.
ALTER TABLE users ADD COLUMN IF NOT EXISTS register_ip VARCHAR(45) NOT NULL DEFAULT '';
