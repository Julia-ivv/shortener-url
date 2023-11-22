package storage

import (
	"context"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
)

var inc int

type RequestBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type Repositories interface {
	GetURL(ctx context.Context, shortURL string, userID int) (originURL string, ok bool)
	AddURL(ctx context.Context, originURL string, userID int) (shortURL string, err error)
	AddBatch(ctx context.Context, originURLBatch []RequestBatch, baseURL string, userID int) (shortURLBatch []ResponseBatch, err error)
	GetAllUserURLs(ctx context.Context, baseURL string, userID int) (userURLs []UserURL, err error)
	PingStor(ctx context.Context) (err error)
	Close() (err error)
}

type Stor struct {
	Repo Repositories
}

func NewURLs(flags config.Flags) (Stor, error) {
	if flags.DBDSN != "" {
		db, err := NewConnectDB(flags.DBDSN)
		if err != nil {
			return Stor{}, err
		}
		return Stor{
			Repo: db,
		}, nil
	}

	if flags.FileName != "" {
		fUrls, err := NewFileURLs(flags.FileName)
		if err != nil {
			return Stor{}, err
		}
		return Stor{
			Repo: fUrls,
		}, nil
	}

	return Stor{
		Repo: NewMapURLs(),
	}, nil
}
