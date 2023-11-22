package storage

import (
	"context"
	"sync"
)

type url struct {
	userID    int
	shortURL  string
	originURL string
}

type MapURLs struct {
	sync.RWMutex
	originalURLs []url
}

func NewMapURLs() *MapURLs {
	return &MapURLs{
		originalURLs: make([]url, 0),
	}
}

func (urls *MapURLs) GetURL(ctx context.Context, shortURL string, userID int) (originURL string, ok bool) {
	// получить длинный урл
	urls.RLock()
	defer urls.RUnlock()

	for _, v := range urls.originalURLs {
		if (v.shortURL == shortURL) && (v.userID == userID) {
			return v.originURL, true
		}
	}
	return "", false
}

func (urls *MapURLs) AddURL(ctx context.Context, originURL string, userID int) (shortURL string, err error) {
	// добавить новый урл
	short, err := GenerateRandomString(LengthShortURL)
	if err != nil {
		return "", err
	}

	urls.Lock()
	defer urls.Unlock()

	urls.originalURLs = append(urls.originalURLs, url{
		userID:    userID,
		shortURL:  short,
		originURL: originURL,
	})
	return short, nil
}

func (urls *MapURLs) AddBatch(ctx context.Context, originURLBatch []RequestBatch, baseURL string, userID int) (shortURLBatch []ResponseBatch, err error) {
	allUrls := make([]url, 0)
	for _, v := range originURLBatch {
		sURL, err := GenerateRandomString(LengthShortURL)
		if err != nil {
			return nil, err
		}
		shortURLBatch = append(shortURLBatch, ResponseBatch{
			CorrelationID: v.CorrelationID,
			ShortURL:      baseURL + sURL,
		})
		allUrls = append(allUrls, url{
			userID:    userID,
			shortURL:  sURL,
			originURL: v.OriginalURL,
		})
	}

	urls.Lock()
	defer urls.Unlock()
	for _, v := range allUrls {
		urls.originalURLs = append(urls.originalURLs, v)
	}

	return shortURLBatch, nil
}

func (urls *MapURLs) GetAllUserURLs(ctx context.Context, baseURL string, userID int) (userURLs []UserURL, err error) {
	urls.RLock()
	defer urls.RUnlock()

	for _, v := range urls.originalURLs {
		if v.userID == userID {
			userURLs = append(userURLs, UserURL{
				ShortURL:    baseURL + v.shortURL,
				OriginalURL: v.originURL,
			})
		}
	}

	return userURLs, nil
}

func (urls *MapURLs) PingStor(ctx context.Context) error {
	return nil
}

func (urls *MapURLs) Close() error {
	return nil
}
