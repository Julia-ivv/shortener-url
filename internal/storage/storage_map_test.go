package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testUserID = 123

func TestNewMapURLs(t *testing.T) {
	t.Run("create new map storage", func(t *testing.T) {
		mapURLs := NewMapURLs()
		assert.NotEmpty(t, mapURLs)
	})
}

func TestGetURL(t *testing.T) {
	testRepo := NewMapURLs()
	testR := make([]MemURL, 0)
	testR = append(testR, MemURL{
		userID:      testUserID,
		shortURL:    "EwH",
		originURL:   "https://practicum.yandex.ru/",
		deletedFlag: false,
	})
	testR = append(testR, MemURL{
		userID:      testUserID,
		shortURL:    "Ert",
		originURL:   "https://mail.ru/",
		deletedFlag: false,
	})
	testRepo.originalURLs = testR

	tests := []struct {
		name     string
		shortURL string
		wantURL  string
	}{
		{
			name:     "url exists",
			shortURL: "EwH",
			wantURL:  "https://practicum.yandex.ru/",
		},
		{
			name:     "url not exists",
			shortURL: "Euu",
			wantURL:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			orig, _, _ := testRepo.GetURL(context.Background(), test.shortURL)
			assert.Equal(t, test.wantURL, orig)
		})
	}
}

func TestAddURL(t *testing.T) {
	testRepo := NewMapURLs()
	t.Run("add url in storage", func(t *testing.T) {
		_, err := testRepo.AddURL(context.Background(), "rtt", "https://mail.ru/", testUserID)
		assert.NoError(t, err)
	})
}

func TestAddBatch(t *testing.T) {
	testRepo := NewMapURLs()
	testRequestBatch := []RequestBatch{
		{
			CorrelationID: "ind1",
			OriginalURL:   "https://pract.ru/url1",
		},
		{
			CorrelationID: "ind2",
			OriginalURL:   "https://pract.ru/url2",
		},
	}
	testResponseBatch := []ResponseBatch{
		{
			CorrelationID: "ind1",
			ShortURLFull:  "ggg",
			ShortURL:      cfg.URL + "ggg",
		},
		{
			CorrelationID: "ind2",
			ShortURLFull:  "rrr",
			ShortURL:      cfg.URL + "rrr",
		},
	}
	t.Run("add batch url in storage", func(t *testing.T) {
		err := testRepo.AddBatch(context.Background(), testResponseBatch, testRequestBatch, testUserID)
		assert.NoError(t, err)
	})
}

func TestGetAllUserURLS(t *testing.T) {
	testRepo := NewMapURLs()
	testR := make([]MemURL, 0)
	testR = append(testR, MemURL{
		userID:      testUserID,
		shortURL:    "EwH",
		originURL:   "https://practicum.yandex.ru/",
		deletedFlag: false,
	})
	testR = append(testR, MemURL{
		userID:      testUserID,
		shortURL:    "Ert",
		originURL:   "https://mail.ru/",
		deletedFlag: false,
	})
	testR = append(testR, MemURL{
		userID:      88,
		shortURL:    "Ert",
		originURL:   "https://mail.ru/",
		deletedFlag: false,
	})
	testRepo.originalURLs = testR

	t.Run("get urls", func(t *testing.T) {
		userURLs, err := testRepo.GetAllUserURLs(context.Background(), cfg.URL, testUserID)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, len(userURLs))
		}
	})
}

func TestDeleteUserURLs(t *testing.T) {
	testRepo := NewMapURLs()
	testR := make([]MemURL, 0)
	testR = append(testR, MemURL{
		userID:      testUserID,
		shortURL:    "EwH",
		originURL:   "https://practicum.yandex.ru/",
		deletedFlag: false,
	})
	testR = append(testR, MemURL{
		userID:      testUserID,
		shortURL:    "Ert",
		originURL:   "https://mail.ru/",
		deletedFlag: false,
	})
	testR = append(testR, MemURL{
		userID:      88,
		shortURL:    "Ert",
		originURL:   "https://mail.ru/",
		deletedFlag: false,
	})
	testRepo.originalURLs = testR

	t.Run("mark deletet", func(t *testing.T) {
		del := "EwH"
		err := testRepo.DeleteUserURLs(context.Background(), []string{del}, testUserID)
		assert.NoError(t, err)
		for _, u := range testRepo.originalURLs {
			if u.shortURL == del {
				assert.Equal(t, true, u.deletedFlag)
				continue
			}
			assert.Equal(t, false, u.deletedFlag)
		}
	})
}

func TestPingStor(t *testing.T) {
	testRepo := NewMapURLs()
	t.Run("ping", func(t *testing.T) {
		assert.NoError(t, testRepo.PingStor(context.Background()))
	})
}

func TestClose(t *testing.T) {
	testRepo := NewMapURLs()
	t.Run("close stor", func(t *testing.T) {
		assert.NoError(t, testRepo.Close())
	})
}

func TestGetStats(t *testing.T) {
	testRepo := NewMapURLs()
	testR := make([]MemURL, 0)
	testR = append(testR, MemURL{
		userID:      testUserID,
		shortURL:    "EwH",
		originURL:   "https://practicum.yandex.ru/",
		deletedFlag: false,
	})
	testR = append(testR, MemURL{
		userID:      testUserID,
		shortURL:    "Ert",
		originURL:   "https://mail.ru/",
		deletedFlag: false,
	})
	testR = append(testR, MemURL{
		userID:      88,
		shortURL:    "Ert",
		originURL:   "https://mail.ru/",
		deletedFlag: false,
	})
	testRepo.originalURLs = testR

	t.Run("get stats", func(t *testing.T) {
		stats, err := testRepo.GetStats(context.Background())
		if assert.NoError(t, err) {
			assert.Equal(t, ServiceStats{Users: 2, URLs: 3}, stats)
		}
	})
}
