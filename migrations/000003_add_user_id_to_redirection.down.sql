DROP INDEX IF EXISTS shortener.idx_redirection_user_id;

ALTER TABLE shortener.redirection DROP COLUMN IF EXISTS user_id;
