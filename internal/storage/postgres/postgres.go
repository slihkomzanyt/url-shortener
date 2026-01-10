package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(dsn string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func initSchema(db *sql.DB) error {
	const createTable = `
 CREATE TABLE IF NOT EXISTS urls (
  id SERIAL PRIMARY KEY,
  alias TEXT NOT NULL UNIQUE,
  url TEXT NOT NULL
 );`

	if _, err := db.Exec(createTable); err != nil {
		return err
	}

	const createIndex = `
 CREATE INDEX IF NOT EXISTS idx_alias ON urls(alias);`

	_, err := db.Exec(createIndex)
	return err
}

func (s *Storage) SaveURL(urlToSave, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"

	const query = `
  INSERT INTO urls(url, alias)
  VALUES ($1, $2)
  RETURNING id`

	var id int64
	err := s.db.QueryRow(query, urlToSave, alias).Scan(&id)
	if err != nil {
		// duplicate key (unique alias)
		if strings.Contains(err.Error(), "duplicate key") {
			return 0, storage.ErrURLExists
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgres.GetURL"

	var url string
	err := s.db.QueryRow(
		`SELECT url FROM urls WHERE alias = $1`,
		alias,
	).Scan(&url)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgres.DeleteURL"

	res, err := s.db.Exec(
		`DELETE FROM urls WHERE alias = $1`,
		alias,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if affected == 0 {
		return storage.ErrURLNotFound
	}

	return nil
}
