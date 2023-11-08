package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBURLs struct {
	dbHandle *sql.DB
}

func NewConnectDB(DBDSN string) (*DBURLs, error) {
	db, err := sql.Open("pgx", DBDSN)
	if err != nil {
		return nil, err
	}

	return &DBURLs{dbHandle: db}, nil
}

func (db *DBURLs) dbInit() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.dbHandle.ExecContext(ctx,
		"CREATE TABLE IF NOT EXISTS urls (short_url text, original_url text UNIQUE)")
	if err != nil {
		return err
	}
	return nil
}

func (db *DBURLs) GetURL(ctx context.Context, shortURL string) (originURL string, ok bool) {
	// получить длинный урл
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	row := db.dbHandle.QueryRowContext(ctx,
		"SELECT original_url FROM urls WHERE short_url=$1", shortURL)

	err := row.Scan(&originURL)
	if err != nil {
		return "", false
	}

	return originURL, true
}

func (db *DBURLs) AddURL(ctx context.Context, originURL string) (shortURL string, err error) {
	// добавить урл в БД
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	shortURL = GenerateRandomString(LengthShortURL)
	result, err := db.dbHandle.ExecContext(ctx,
		"INSERT INTO urls VALUES ($1, $2)", shortURL, originURL)
	if err != nil {
		return "", err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return "", err
	}
	if rows != 1 {
		return "", fmt.Errorf("expected to affect 1 row, affected %d", rows)
	}
	return shortURL, nil
}
