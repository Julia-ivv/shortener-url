package storage

import (
	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
)

var inc int

type Repositories interface {
	GetURL(shortURL string) (originURL string, ok bool)
	AddURL(originURL string) (shortURL string, err error)
}

func NewURLs(flags config.Flags) (Repositories, error) {
	if flags.FileName == "" {
		mapURL := make(map[string]string)
		mapURL["EwHXdJfB"] = "https://practicum.yandex.ru/"
		return &MapURLs{
			originalURLs: mapURL,
		}, nil
	}

	fUrls, err := NewFileURLs(flags.FileName)
	if err != nil {
		return nil, err
	}
	return fUrls, nil
}
