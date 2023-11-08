package storage

import (
	"database/sql"

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

func (db *DBURLs) GetURL(shortURL string) (originURL string, ok bool) {
	return "", true
}

func (db *DBURLs) AddURL(originURL string) (shortURL string, err error) {
	return "", nil
}
