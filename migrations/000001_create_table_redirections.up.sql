CREATE TABLE IF NOT EXISTS redirection (
    id BIGSERIAL PRIMARY KEY,
    original_url TEXT NOT NULL,
    redirect_url VARCHAR(8) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_redirection_redirect_url
ON redirection (redirect_url);