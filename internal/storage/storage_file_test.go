package storage

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testFileName = "for_tests.json"

func fillFile() error {
	text := `{"short_url":"OGAE8Q","original_url":"https://ya.ru","is_deleted":false,"user_id":574039855}
{"short_url":"H_O4PA","original_url":"https://pract.ru/url1","is_deleted":false,"user_id":1777238335}
{"short_url":"-YtNlA","original_url":"https://pract.ru/url2","is_deleted":false,"user_id":1777238335}
{"short_url":"1IVh8Q","original_url":"https://pract.ru/url3","is_deleted":false,"user_id":1777238335}
{"short_url":"dfT_vA","original_url":"https://pract.ru/url4","is_deleted":false,"user_id":1777238335}
`
	file, err := os.OpenFile(testFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(text)
	return nil
}

func TestNewFileURLs(t *testing.T) {
	t.Run("create file", func(t *testing.T) {
		file, err := NewFileURLs(testFileName)
		if assert.NoError(t, err) {
			assert.NotEmpty(t, file)
		}
	})
}

func TestFileGetURL(t *testing.T) {
	err := fillFile()
	if err != nil {
		t.Fatal("Unable to create file:", err)
	}
	testRepo, err := NewFileURLs(testFileName)
	tests := []struct {
		name     string
		short    string
		wantURL  string
		wantFind bool
	}{
		{
			name:     "url exists",
			short:    "dfT_vA",
			wantURL:  "https://pract.ru/url4",
			wantFind: true,
		},
		{
			name:     "url not exists",
			short:    "df",
			wantURL:  "",
			wantFind: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if assert.NoError(t, err) {
				u, _, find := testRepo.GetURL(context.Background(), test.short)
				assert.Equal(t, test.wantURL, u)
				assert.Equal(t, test.wantFind, find)
			}
		})
	}
}

func TestFileAddURL(t *testing.T) {
	testRepo, err := NewFileURLs(testFileName)
	t.Run("add url in file", func(t *testing.T) {
		if assert.NoError(t, err) {
			_, err := testRepo.AddURL(context.Background(), "sh", "https://mail.ru", testUserID)
			assert.NoError(t, err)
		}
	})
}

func TestFileAddBatch(t *testing.T) {
	testRepo, errFile := NewFileURLs(testFileName)
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

	t.Run("add batch in file", func(t *testing.T) {
		if assert.NoError(t, errFile) {
			err := testRepo.AddBatch(context.Background(), testResponseBatch, testRequestBatch, testUserID)
			assert.NoError(t, err)
		}
	})
}

func TestFileGetAllUserURLs(t *testing.T) {
	err := fillFile()
	if err != nil {
		t.Fatal("Unable to create file:", err)
	}
	testRepo, errFile := NewFileURLs(testFileName)
	t.Run("get user urls", func(t *testing.T) {
		if assert.NoError(t, errFile) {
			userURLs, err := testRepo.GetAllUserURLs(context.Background(), cfg.URL, 1777238335)
			assert.NoError(t, err)
			assert.NotEmpty(t, userURLs)
		}
	})
}

func TestFileDeleteUserURLs(t *testing.T) {
	err := fillFile()
	if err != nil {
		t.Fatal("Unable to create file:", err)
	}
	testRepo, errFile := NewFileURLs(testFileName)
	t.Run("mark deleted", func(t *testing.T) {
		if assert.NoError(t, errFile) {
			err := testRepo.DeleteUserURLs(context.Background(), []string{"H_O4PA", "-YtNlA"}, 1777238335)
			assert.NoError(t, err)
		}
	})
}

func TestFilePingStor(t *testing.T) {
	testRepo, errFile := NewFileURLs(testFileName)
	t.Run("ping", func(t *testing.T) {
		if assert.NoError(t, errFile) {
			assert.NoError(t, testRepo.PingStor(context.Background()))
		}
		testRepo.fileName = ""
		assert.NotEqual(t, nil, testRepo.PingStor(context.Background()))
	})
}

func TestFileClose(t *testing.T) {
	testRepo, errFile := NewFileURLs(testFileName)
	t.Run("close storage", func(t *testing.T) {
		if assert.NoError(t, errFile) {
			assert.NoError(t, testRepo.Close())
		}
	})
}

func TestFileGetStats(t *testing.T) {
	err := fillFile()
	if err != nil {
		t.Fatal("Unable to create file:", err)
	}
	testRepo, errFile := NewFileURLs(testFileName)
	t.Run("get stats", func(t *testing.T) {
		if assert.NoError(t, errFile) {
			stats, err := testRepo.GetStats(context.Background())
			assert.NoError(t, err)
			assert.NotEmpty(t, stats)
		}
	})
}
