DELETE FROM shortener.redirection a USING shortener.redirection b
WHERE a.id > b.id AND a.original_url = b.original_url;

CREATE UNIQUE INDEX idx_redirection_original_url
ON shortener.redirection (original_url);
