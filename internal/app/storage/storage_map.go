package storage

import (
	"context"
	"sync"
)

type MemURL struct {
	userID    int
	shortURL  string
	originURL string
}

type MemURLs struct {
	sync.RWMutex
	originalURLs []MemURL
}

func NewMapURLs() *MemURLs {
	return &MemURLs{
		originalURLs: make([]MemURL, 0),
	}
}

func (urls *MemURLs) GetURL(ctx context.Context, shortURL string, userID int) (originURL string, ok bool) {
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

func (urls *MemURLs) AddURL(ctx context.Context, originURL string, userID int) (shortURL string, err error) {
	// добавить новый урл
	short, err := GenerateRandomString(LengthShortURL)
	if err != nil {
		return "", err
	}

	urls.Lock()
	defer urls.Unlock()

	urls.originalURLs = append(urls.originalURLs, MemURL{
		userID:    userID,
		shortURL:  short,
		originURL: originURL,
	})
	return short, nil
}

func (urls *MemURLs) AddBatch(ctx context.Context, originURLBatch []RequestBatch, baseURL string, userID int) (shortURLBatch []ResponseBatch, err error) {
	allUrls := make([]MemURL, 0)
	for _, v := range originURLBatch {
		sURL, err := GenerateRandomString(LengthShortURL)
		if err != nil {
			return nil, err
		}
		shortURLBatch = append(shortURLBatch, ResponseBatch{
			CorrelationID: v.CorrelationID,
			ShortURL:      baseURL + sURL,
		})
		allUrls = append(allUrls, MemURL{
			userID:    userID,
			shortURL:  sURL,
			originURL: v.OriginalURL,
		})
	}

	urls.Lock()
	defer urls.Unlock()
	urls.originalURLs = append(urls.originalURLs, allUrls...)

	return shortURLBatch, nil
}

func (urls *MemURLs) GetAllUserURLs(ctx context.Context, baseURL string, userID int) (userURLs []UserURL, err error) {
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

func (urls *MemURLs) PingStor(ctx context.Context) error {
	return nil
}

func (urls *MemURLs) Close() error {
	return nil
}
