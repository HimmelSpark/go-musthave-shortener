package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type URLBatchItem struct {
	OriginalURL string
	ShortID     string
}

var ErrBatchCollision = errors.New("batch insert collision")

type (
	URLRepository interface {
		FindURLShortID(string) (string, error)
		CreateURL(originalURL string, shortID string) (bool, error)
		CreateURLBatch(items []URLBatchItem) error
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

func (u urlRepository) CreateURLBatch(items []URLBatchItem) error {
	tx, err := u.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var b strings.Builder
	b.WriteString("INSERT INTO shortener.redirection (original_url, redirect_url) VALUES ")

	args := make([]any, 0, len(items)*2)
	for i, item := range items {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(fmt.Sprintf("($%d,$%d)", i*2+1, i*2+2))
		args = append(args, item.OriginalURL, item.ShortID)
	}
	b.WriteString(" ON CONFLICT (redirect_url) DO NOTHING")

	res, err := tx.Exec(b.String(), args...)
	if err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if int(n) != len(items) {
		return ErrBatchCollision
	}

	return tx.Commit()
}
