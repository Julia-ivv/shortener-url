package storage

import (
	"context"
	"database/sql"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
)

var inc int

type Repositories interface {
	GetURL(ctx context.Context, shortURL string) (originURL string, ok bool)
	AddURL(ctx context.Context, originURL string) (shortURL string, err error)
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
		err = db.dbInit()
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

	mapURL := make(map[string]string)
	mapURL["EwHXdJfB"] = "https://practicum.yandex.ru/"
	return Stor{
		Repo:     &MapURLs{originalURLs: mapURL},
		DBHandle: nil,
	}, nil
}
