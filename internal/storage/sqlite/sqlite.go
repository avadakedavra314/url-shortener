package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/avadakedavra314/url-shortener/internal/storage"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url(
		id INTEGER PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveUrl(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveUrl"

	query := "INSERT INTO url(url, alias) values (?,?)"

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("%s: prepare insert statement: %w", op, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		// Check if we added already existing alias and return common error - storage.ErrUrlExist
		sqliteErr, ok := err.(sqlite3.Error)
		if ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: exec insert statement: %w", op, storage.ErrUrlExist)
		}
		return 0, fmt.Errorf("%s: exec insert statement: %w", op, err)
	}

	id, err := res.LastInsertId()

	if err != nil {
		return 0, fmt.Errorf("%s: failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetUrl(urlAlias string) (string, error) {
	const op = "storage.Sqlite.GetUrl"

	query := "SELECT url FROM url WHERE alias = ?"
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return "", fmt.Errorf("%s: prepare select statement: %w", op, err)
	}

	var resUrl string
	err = stmt.QueryRow(urlAlias).Scan(&resUrl)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%s: exec select statement: %w", op, storage.ErrUrlNotFound)
		}
		return "", fmt.Errorf("%s: exec select statement: %w", op, err)
	}

	return resUrl, nil
}

func (s *Storage) DeleteUrl(urlAlias string) error {
	const op = "storage.Sqlite.DeleteUrl"

	query := "DELETE FROM url WHERE alias = ?"
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: prepare delete statement: %w", op, err)
	}
	res, err := stmt.Exec(urlAlias)
	if err != nil {
		return fmt.Errorf("%s: exec delete statement: %w", op, err)
	}

	if rowsCount, _ := res.RowsAffected(); rowsCount == 0 {
		return fmt.Errorf("%s: failed to delete url, alias: %s %w", op, urlAlias, err)
	}

	return nil
}
