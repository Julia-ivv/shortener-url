package storage

import (
	"context"
	"errors"
	"slices"
	"sync"
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
func (urls *MemURLs) AddURL(ctx context.Context, shortURL string, originURL string, userID int) (findURL string, err error) {
	urls.Lock()
	defer urls.Unlock()

	urls.originalURLs = append(urls.originalURLs, MemURL{
		userID:      userID,
		shortURL:    shortURL,
		originURL:   originURL,
		deletedFlag: false,
	})
	return "", nil
}

// AddBatch adds a batch of new short URLs.
func (urls *MemURLs) AddBatch(ctx context.Context, shortURLBatch []ResponseBatch, originURLBatch []RequestBatch, userID int) (err error) {
	allUrls := make([]MemURL, len(shortURLBatch))
	for k, v := range shortURLBatch {
		allUrls = append(allUrls, MemURL{
			userID:      userID,
			shortURL:    v.ShortURL,
			originURL:   originURLBatch[k].OriginalURL,
			deletedFlag: false,
		})
	}

	urls.Lock()
	defer urls.Unlock()
	urls.originalURLs = append(urls.originalURLs, allUrls...)

	return nil
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

// GetStats gets statistics - amount URLs and users.
func (urls *MemURLs) GetStats(ctx context.Context) (stats ServiceStats, err error) {
	urls.Lock()
	defer urls.Unlock()

	stats.URLs = len(urls.originalURLs)
	stats.Users = 0

	tmp := make([]int, len(urls.originalURLs))
	for _, v := range urls.originalURLs {
		if !slices.Contains(tmp, v.userID) {
			tmp = append(tmp, v.userID)
			stats.Users++
		}
	}

	return stats, nil
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
