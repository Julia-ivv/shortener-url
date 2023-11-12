package storage

import (
	"context"
	"sync"
)

type MapURLs struct {
	sync.RWMutex
	originalURLs map[string]string
}

func NewMapURLs() *MapURLs {
	mapURL := make(map[string]string)
	mapURL["EwHXdJfB"] = "https://practicum.yandex.ru/"
	return &MapURLs{originalURLs: mapURL}
}

func (urls *MapURLs) GetURL(ctx context.Context, shortURL string) (originURL string, ok bool) {
	// получить длинный урл
	urls.RLock()
	defer urls.RUnlock()

	originURL, ok = urls.originalURLs[shortURL]
	return originURL, ok
}

func (urls *MapURLs) AddURL(ctx context.Context, originURL string) (shortURL string, err error) {
	// добавить новый урл
	short, err := GenerateRandomString(LengthShortURL)
	if err != nil {
		return "", err
	}

	urls.Lock()
	defer urls.Unlock()

	urls.originalURLs[short] = originURL
	return short, nil
}

func (urls *MapURLs) AddBatch(ctx context.Context, originURLBatch []RequestBatch, baseURL string) (shortURLBatch []ResponseBatch, err error) {
	allUrls := make(map[string]string)
	for _, v := range originURLBatch {
		sURL, err := GenerateRandomString(LengthShortURL)
		if err != nil {
			return nil, err
		}
		shortURLBatch = append(shortURLBatch, ResponseBatch{
			CorrelationId: v.CorrelationId,
			ShortURL:      baseURL + sURL,
		})
		allUrls[sURL] = v.OriginalURL
	}

	urls.Lock()
	defer urls.Unlock()
	for k, v := range allUrls {
		urls.originalURLs[k] = v
	}

	return shortURLBatch, nil
}
