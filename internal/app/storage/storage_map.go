package storage

import (
	"github.com/Julia-ivv/shortener-url.git/internal/app/tools"
)

type MapURLs struct {
	originalURLs map[string]string
}

func (urls *MapURLs) GetURL(shortURL string) (originURL string, ok bool) {
	// получить длинный урл
	originURL, ok = urls.originalURLs[shortURL]
	return originURL, ok
}

func (urls *MapURLs) AddURL(originURL string) (shortURL string, err error) {
	// добавить новый урл
	short := tools.GenerateRandomString(tools.LengthShortURL)
	urls.originalURLs[short] = originURL
	return short, nil
}

func (urls *MapURLs) Close() error {
	return nil
}
