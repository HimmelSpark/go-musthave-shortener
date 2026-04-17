package repository

import (
	"database/sql"
	"errors"
)

type (
	URLRepository interface {
		FindURLShortID(string) (string, error)
		CreateURL(originalURL string, shortID string) (bool, error)
	}

	urlRepository struct {
		db *sql.DB
	}
)

func NewURLRepository(db *sql.DB) URLRepository {
	return &urlRepository{db: db}
}

func (u urlRepository) FindURLShortID(shortURLID string) (string, error) {
	row := u.db.QueryRow("SELECT original_url FROM shortener.redirection WHERE redirect_url = $1", shortURLID)

	var originalURL string

	err := row.Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	return originalURL, nil
}

func (u urlRepository) CreateURL(originalURL string, shortID string) (bool, error) {

	res, err := u.db.Exec(
		`INSERT INTO shortener.redirection (original_url, redirect_url) 
				VALUES ($1, $2) ON CONFLICT (redirect_url) DO NOTHING`,
		originalURL,
		shortID,
	)

	if err != nil {
		return false, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}
