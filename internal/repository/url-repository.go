package repository

import (
	"database/sql"
	"errors"
)

type (
	UrlRepository interface {
		FindUrl(string) (string, error)
		CreateUrl(originalUrl string, shortId string) (bool, error)
	}

	urlRepository struct {
		db *sql.DB
	}
)

func NewUrlRepository(db *sql.DB) UrlRepository {
	return &urlRepository{db: db}
}

func (u urlRepository) FindUrl(s string) (string, error) {
	row := u.db.QueryRow("SELECT original_url FROM shortener.redirection WHERE redirect_url = $1", s)

	var originalUrl string

	err := row.Scan(&originalUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	return originalUrl, nil
}

func (u urlRepository) CreateUrl(originalUrl string, shortId string) (bool, error) {

	res, err := u.db.Exec(
		`INSERT INTO shortener.redirection (original_url, redirect_url) 
				VALUES ($1, $2) ON CONFLICT (redirect_url) DO NOTHING`,
		originalUrl,
		shortId,
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
