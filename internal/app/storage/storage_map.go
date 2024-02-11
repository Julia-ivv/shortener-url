package storage

import (
	"context"
	"errors"
	"sync"

	"github.com/Julia-ivv/shortener-url.git/pkg/randomizer"
)

// MemURL stores URL information in memory.
type MemURL struct {
	shortURL    string
	originURL   string
	deletedFlag bool
	userID      int
}

// MemURLs stores information about all URLs in memory.
type MemURLs struct {
	originalURLs []MemURL
	sync.RWMutex
}

// NewMapURLs creates an instance for storing URLs.
func NewMapURLs() *MemURLs {
	return &MemURLs{
		originalURLs: make([]MemURL, 0),
	}
}

// GetURL gets the original URL matching the short URL.
func (urls *MemURLs) GetURL(ctx context.Context, shortURL string) (originURL string, isDel bool, ok bool) {
	urls.RLock()
	defer urls.RUnlock()

	for _, v := range urls.originalURLs {
		if v.shortURL == shortURL {
			return v.originURL, v.deletedFlag, true
		}
	}
	return "", false, false
}

// AddURL adds a new short url.
func (urls *MemURLs) AddURL(ctx context.Context, originURL string, userID int) (shortURL string, err error) {
	short, err := randomizer.GenerateRandomString(randomizer.LengthShortURL)
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

// AddBatch adds a batch of new short URLs.
func (urls *MemURLs) AddBatch(ctx context.Context, originURLBatch []RequestBatch, baseURL string, userID int) (shortURLBatch []ResponseBatch, err error) {
	allUrls := make([]MemURL, len(originURLBatch))
	shortURLBatch = make([]ResponseBatch, len(originURLBatch))
	for _, v := range originURLBatch {
		sURL, err := randomizer.GenerateRandomString(randomizer.LengthShortURL)
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

// GetAllUserURLs gets all user's short url.
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

// DeleteUserURLs sets the deletion flag to the user URLs sent in the request.
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

// PingStor checking access to storage.
func (urls *MemURLs) PingStor(ctx context.Context) error {
	if urls == nil {
		return errors.New("storage storage does not exist")
	}
	return nil
}

// Close closes the storage.
func (urls *MemURLs) Close() error {
	urls = nil
	return nil
}
