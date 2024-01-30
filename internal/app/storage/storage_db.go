package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/Julia-ivv/shortener-url.git/internal/app/deleter"
	"github.com/Julia-ivv/shortener-url.git/internal/app/logger"
)

// DBURLs stores a pointer to the database.
type DBURLs struct {
	dbHandle *sql.DB
}

// NewConnectDB creates a connection to the database.
func NewConnectDB(DBDSN string) (*DBURLs, error) {
	db, err := sql.Open("pgx", DBDSN)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx,
		"CREATE TABLE IF NOT EXISTS urls (user_id integer, short_url text, original_url text, deleted_flag boolean DEFAULT false, PRIMARY KEY(user_id, original_url))")
	if err != nil {
		return nil, err
	}

	return &DBURLs{dbHandle: db}, nil
}

// GetURL gets the original URL matching the short URL.
func (db *DBURLs) GetURL(ctx context.Context, shortURL string) (originURL string, isDel bool, ok bool) {
	// получить длинный урл не учитывая пользователя
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	row := db.dbHandle.QueryRowContext(ctx,
		"SELECT original_url, deleted_flag FROM urls WHERE short_url=$1", shortURL)

	err := row.Scan(&originURL, &isDel)
	if err != nil {
		return "", false, false
	}

	return originURL, isDel, true
}

// UserURL stores pairs: short URL, original URL.
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// GetAllUserURLs gets all user's short url.
func (db *DBURLs) GetAllUserURLs(ctx context.Context, baseURL string, userID int) (userURLs []UserURL, err error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	rows, err := db.dbHandle.QueryContext(ctx,
		"SELECT short_url, original_url FROM urls WHERE user_id=$1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u UserURL
		err = rows.Scan(&u.ShortURL, &u.OriginalURL)
		if err != nil {
			return nil, err
		}
		userURLs = append(userURLs, UserURL{
			ShortURL:    baseURL + u.ShortURL,
			OriginalURL: u.OriginalURL,
		})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return userURLs, nil
}

// AddURL adds a new short url.
func (db *DBURLs) AddURL(ctx context.Context, originURL string, userID int) (shortURL string, err error) {
	// добавить урл в БД
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	shortURL, err = GenerateRandomString(LengthShortURL)
	if err != nil {
		return "", err
	}
	result, err := db.dbHandle.ExecContext(ctx,
		"INSERT INTO urls VALUES ($1, $2, $3)", userID, shortURL, originURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			row := db.dbHandle.QueryRowContext(ctx,
				"SELECT short_url FROM urls WHERE original_url=$1 AND user_id=$2", originURL, userID)
			errScan := row.Scan(&shortURL)
			if errScan != nil {
				return "", err
			}
			return shortURL, err
		}
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

// AddBatch adds a batch of new short URLs.
func (db *DBURLs) AddBatch(ctx context.Context, originURLBatch []RequestBatch, baseURL string, userID int) (shortURLBatch []ResponseBatch, err error) {
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
		result, err := tx.ExecContext(ctx, "INSERT INTO urls VALUES ($1, $2, $3)", userID, shortURL, v.OriginalURL)
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

// DeleteUserURLs sets the deletion flag to the user URLs sent in the request.
func (db *DBURLs) DeleteUserURLs(ctx context.Context, delURLs []string, userID int) (err error) {
	doneCh := make(chan struct{})
	defer close(doneCh)

	stmt, err := db.dbHandle.Prepare("UPDATE urls SET deleted_flag = true WHERE user_id = $1 AND short_url = $2")
	if err != nil {
		logger.ZapSugar.Infow("prepare context error", err)
		return err
	}
	defer stmt.Close()

	inputCh := deleter.Generator(doneCh, delURLs, userID)
	chans := deleter.FanOut(doneCh, inputCh, stmt)
	resCh := deleter.FanIn(stmt, doneCh, chans...)
	cnt := 0
	for res := range resCh {
		if res.Err == nil {
			cnt += int(res.Rows)
		}
	}
	logger.ZapSugar.Infof("user ID %d - removed %d out of %d", userID, cnt, len(delURLs))

	return nil
}

// PingStor checking access to storage.
func (db *DBURLs) PingStor(ctx context.Context) error {
	return db.dbHandle.PingContext(ctx)
}

// Close closes the storage.
func (db *DBURLs) Close() error {
	return db.dbHandle.Close()
}
