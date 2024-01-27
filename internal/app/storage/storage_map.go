package storage

import (
	"context"
	"errors"
	"sync"
)

type MemURL struct {
	userID      int
	shortURL    string
	originURL   string
	deletedFlag bool
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

func (urls *MemURLs) GetURL(ctx context.Context, shortURL string) (originURL string, isDel bool, ok bool) {
	// получить длинный урл без учета пользователя
	urls.RLock()
	defer urls.RUnlock()

	for _, v := range urls.originalURLs {
		if v.shortURL == shortURL {
			return v.originURL, v.deletedFlag, true
		}
	}
	return "", false, false
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
		userID:      userID,
		shortURL:    short,
		originURL:   originURL,
		deletedFlag: false,
	})
	return short, nil
}

func (urls *MemURLs) AddBatch(ctx context.Context, originURLBatch []RequestBatch, baseURL string, userID int) (shortURLBatch []ResponseBatch, err error) {
	allUrls := make([]MemURL, len(originURLBatch))
	shortURLBatch = make([]ResponseBatch, len(originURLBatch))
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
			userID:      userID,
			shortURL:    sURL,
			originURL:   v.OriginalURL,
			deletedFlag: false,
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

func (urls *MemURLs) DeleteUserURLs(ctx context.Context, delURLs []string, userID int) (err error) {
	urls.Lock()
	defer urls.Unlock()

	for _, delURL := range delURLs {
		for k, curURL := range urls.originalURLs {
			if (delURL == curURL.shortURL) && (userID == curURL.userID) {
				urls.originalURLs[k].deletedFlag = true
				break
			}
		}
	}
	return nil
}

func (urls *MemURLs) PingStor(ctx context.Context) error {
	if urls == nil {
		return errors.New("storage storage does not exist")
	}
	return nil
}

func (urls *MemURLs) Close() error {
	urls = nil
	return nil
}
