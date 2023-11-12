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

func (db *DBURLs) CreateAllTables() error {
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

	shortURL, err = GenerateRandomString(LengthShortURL)
	if err != nil {
		return "", err
	}
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

func (db *DBURLs) AddBatch(ctx context.Context, originURLBatch []RequestBatch, baseURL string) (shortURLBatch []ResponseBatch, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx, err := db.dbHandle.Begin()
	if err != nil {
		return nil, err
	}

	for _, v := range originURLBatch {
		shortURL, err := GenerateRandomString(LengthShortURL)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		result, err := tx.ExecContext(ctx, "INSERT INTO urls VALUES ($1, $2)", shortURL, v.OriginalURL)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		rows, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		if rows != 1 {
			tx.Rollback()
			return nil, fmt.Errorf("expected to affect 1 row, affected %d", rows)
		}

		shortURLBatch = append(shortURLBatch, ResponseBatch{
			CorrelationID: v.CorrelationID,
			ShortURL:      baseURL + shortURL,
		})
	}

	return shortURLBatch, tx.Commit()
}
