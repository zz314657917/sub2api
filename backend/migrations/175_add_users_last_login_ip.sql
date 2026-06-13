-- Record the client IP observed during the latest successful login.
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_ip VARCHAR(45) NOT NULL DEFAULT '';
