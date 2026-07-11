-- Allows administrators to opt a platform user out of leaderboard participation.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS exclude_from_leaderboard BOOLEAN NOT NULL DEFAULT FALSE;
