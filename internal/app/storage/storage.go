package storage

import (
	"context"
	"database/sql"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
)

var inc int

type RequestBatch struct {
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseBatch struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type Repositories interface {
	GetURL(ctx context.Context, shortURL string) (originURL string, ok bool)
	AddURL(ctx context.Context, originURL string) (shortURL string, err error)
	AddBatch(ctx context.Context, originURLBatch []RequestBatch, baseURL string) (shortURLBatch []ResponseBatch, err error)
}

type Stor struct {
	Repo     Repositories
	DBHandle *sql.DB
}

func NewURLs(flags config.Flags) (Stor, error) {
	if flags.DBDSN != "" {
		db, err := NewConnectDB(flags.DBDSN)
		if err != nil {
			return Stor{}, err
		}
		err = db.CreateAllTables()
		if err != nil {
			return Stor{}, err
		}
		return Stor{
			Repo:     db,
			DBHandle: db.dbHandle,
		}, nil
	}

	if flags.FileName != "" {
		fUrls, err := NewFileURLs(flags.FileName)
		if err != nil {
			return Stor{}, err
		}
		return Stor{
			Repo:     fUrls,
			DBHandle: nil,
		}, nil
	}

	return Stor{
		Repo:     NewMapURLs(),
		DBHandle: nil,
	}, nil
}
