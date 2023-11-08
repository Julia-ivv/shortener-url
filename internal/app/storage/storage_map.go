package storage

import "context"

type MapURLs struct {
	originalURLs map[string]string
}

func (urls *MapURLs) GetURL(ctx context.Context, shortURL string) (originURL string, ok bool) {
	// получить длинный урл
	originURL, ok = urls.originalURLs[shortURL]
	return originURL, ok
}

func (urls *MapURLs) AddURL(ctx context.Context, originURL string) (shortURL string, err error) {
	// добавить новый урл
	short := GenerateRandomString(LengthShortURL)
	urls.originalURLs[short] = originURL
	return short, nil
}
