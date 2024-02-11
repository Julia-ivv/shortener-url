package storage

import (
	"context"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
)

// var inc int

// RequestBatch stores data from the request body for adding a batch of URLs.
type RequestBatch struct {
	// CorrelationID - URL ID for correlation with the response.
	CorrelationID string `json:"correlation_id"`
	// OriginalURL - URL for shortening.
	OriginalURL string `json:"original_url"`
}

// ResponseBatch stores response data for the handler to add a pack of URLs.
type ResponseBatch struct {
	// CorrelationID - URL ID for correlation with the request.
	CorrelationID string `json:"correlation_id"`
	// ShortURL - short url.
	ShortURL string `json:"short_url"`
}

// Repositories - the interface contains methods for working with the repository.
type Repositories interface {
	// GetURL gets the original URL matching the short URL.
	GetURL(ctx context.Context, shortURL string) (originURL string, isDel bool, ok bool)
	// AddURL adds a new short url.
	AddURL(ctx context.Context, originURL string, userID int) (shortURL string, err error)
	// AddBatch adds a batch of new short URLs.
	AddBatch(ctx context.Context, originURLBatch []RequestBatch, baseURL string, userID int) (shortURLBatch []ResponseBatch, err error)
	// GetAllUserURLs gets all user's short url.
	GetAllUserURLs(ctx context.Context, baseURL string, userID int) (userURLs []UserURL, err error)
	// DeleteUserURLs sets the deletion flag to the user URLs sent in the request.
	DeleteUserURLs(ctx context.Context, delURLs []string, userID int) (err error)
	// PingStor checking access to storage.
	PingStor(ctx context.Context) (err error)
	// Close closes the storage.
	Close() (err error)
}

// NewURLs creates a storage instance.
func NewURLs(flags config.Flags) (Repositories, error) {
	if flags.DBDSN != "" {
		db, err := NewConnectDB(flags.DBDSN)
		if err != nil {
			return nil, err
		}
		return db, nil
	}

	if flags.FileName != "" {
		fUrls, err := NewFileURLs(flags.FileName)
		if err != nil {
			return nil, err
		}
		return fUrls, nil
	}

	return NewMapURLs(), nil
}
