package storage

import (
	"database/sql"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
)

var inc int

type Repositories interface {
	GetURL(shortURL string) (originURL string, ok bool)
	AddURL(originURL string) (shortURL string, err error)
}

type Stor struct {
	Repo     Repositories
	DBHandle *sql.DB
}

func NewURLs(flags config.Flags) (Stor, error) {
	if flags.DBDSN != "" {
		db, err := NewConnectDB(flags.DBDSN)
		return Stor{
			Repo:     db,
			DBHandle: db.dbHandle,
		}, err
	}

	if flags.FileName != "" {
		fUrls, err := NewFileURLs(flags.FileName)
		if err != nil {
			return Stor{
				Repo:     nil,
				DBHandle: nil,
			}, err
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
