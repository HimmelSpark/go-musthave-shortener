package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type URLBatchItem struct {
	OriginalURL string
	ShortID     string
	UserID      string
}

type UserURL struct {
	ShortID     string
	OriginalURL string
}

var ErrBatchCollision = errors.New("batch insert collision")

type ErrDuplicateURL struct {
	ExistingShortID string
}

func (e *ErrDuplicateURL) Error() string {
	return "original URL already exists"
}

type (
	URLRepository interface {
		FindURLShortID(shortID string) (originalURL string, isDeleted bool, err error)
		CreateURL(originalURL string, shortID string, userID string) (bool, error)
		CreateURLBatch(items []URLBatchItem) error
		GetUserURLs(userID string) ([]UserURL, error)
		DeleteUserURLs(userID string, shortIDs []string) error
	}

	urlRepository struct {
		db *sql.DB
	}
)

func NewURLRepository(db *sql.DB) URLRepository {
	return &urlRepository{db: db}
}

func (u urlRepository) FindURLShortID(shortURLID string) (string, bool, error) {
	row := u.db.QueryRow(
		"SELECT original_url, is_deleted FROM shortener.redirection WHERE redirect_url = $1",
		shortURLID,
	)

	var originalURL string
	var isDeleted bool

	err := row.Scan(&originalURL, &isDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}

	return originalURL, isDeleted, nil
}

func (u urlRepository) CreateURL(originalURL string, shortID string, userID string) (bool, error) {
	var returnedShortID string
	err := u.db.QueryRow(
		`INSERT INTO shortener.redirection (original_url, redirect_url, user_id)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
		 RETURNING redirect_url`,
		originalURL,
		shortID,
		userID,
	).Scan(&returnedShortID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return false, nil
		}
		return false, err
	}

	if returnedShortID != shortID {
		return false, &ErrDuplicateURL{ExistingShortID: returnedShortID}
	}

	return true, nil
}

func (u urlRepository) CreateURLBatch(items []URLBatchItem) error {
	tx, err := u.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var b strings.Builder
	b.WriteString("INSERT INTO shortener.redirection (original_url, redirect_url, user_id) VALUES ")

	args := make([]any, 0, len(items)*3)
	for i, item := range items {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(fmt.Sprintf("($%d,$%d,$%d)", i*3+1, i*3+2, i*3+3))
		args = append(args, item.OriginalURL, item.ShortID, item.UserID)
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

func (u urlRepository) DeleteUserURLs(userID string, shortIDs []string) error {
	if userID == "" || len(shortIDs) == 0 {
		return nil
	}

	var b strings.Builder
	b.WriteString("UPDATE shortener.redirection SET is_deleted = TRUE WHERE user_id = $1 AND redirect_url IN (")

	args := make([]any, 0, len(shortIDs)+1)
	args = append(args, userID)
	for i, sid := range shortIDs {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(fmt.Sprintf("$%d", i+2))
		args = append(args, sid)
	}
	b.WriteString(")")

	if _, err := u.db.Exec(b.String(), args...); err != nil {
		return fmt.Errorf("failed to soft-delete user urls: %w", err)
	}
	return nil
}

func (u urlRepository) GetUserURLs(userID string) ([]UserURL, error) {
	if userID == "" {
		return nil, nil
	}

	rows, err := u.db.Query(
		"SELECT redirect_url, original_url FROM shortener.redirection WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []UserURL
	for rows.Next() {
		var item UserURL
		if err := rows.Scan(&item.ShortID, &item.OriginalURL); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

