ALTER TABLE shortener.redirection ADD COLUMN IF NOT EXISTS user_id VARCHAR(64) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_redirection_user_id
ON shortener.redirection (user_id);
