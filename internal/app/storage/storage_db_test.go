package storage

import (
	"context"
	"database/sql/driver"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestDBGetURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error occurred while creating mock: %s", err)
	}
	defer db.Close()

	testDB := DBURLs{dbHandle: db}

	tests := []struct {
		name             string
		queryStr         string
		args             string
		expectedOriginal string
		expectedRows     []string
		expectedValues   []driver.Value
		expectedOk       bool
	}{
		{
			name:             "get url",
			queryStr:         "SELECT original_url, deleted_flag FROM urls",
			args:             "EwH",
			expectedOriginal: "https://practicum.yandex.ru/",
			expectedRows:     []string{"original_url", "deleted_flag"},
			expectedValues:   []driver.Value{"https://practicum.yandex.ru/", "false"},
			expectedOk:       true,
		},
	}

	for _, test := range tests {
		rows := sqlmock.NewRows(test.expectedRows).AddRow(test.expectedValues...)

		mock.ExpectQuery(test.queryStr).WithArgs(test.args).WillReturnRows(rows)

		t.Run(test.name, func(t *testing.T) {
			original, _, ok := testDB.GetURL(context.Background(), test.args)
			assert.Equal(t, test.expectedOk, ok)
			assert.Equal(t, test.expectedOriginal, original)
		})
	}
}

func TestDBGetAllUserURLs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error occurred while creating mock: %s", err)
	}
	defer db.Close()

	testDB := DBURLs{dbHandle: db}

	tests := []struct {
		name           string
		queryStr       string
		expectedRows   []string
		expectedValues []driver.Value
		args           int
	}{
		{
			name:           "get all user url",
			queryStr:       "SELECT short_url, original_url FROM url",
			args:           123,
			expectedRows:   []string{"short_url", "original_url"},
			expectedValues: []driver.Value{"EwH", "https://practicum.yandex.ru/"},
		},
	}

	for _, test := range tests {
		rows := sqlmock.NewRows(test.expectedRows).AddRow(test.expectedValues...)

		mock.ExpectQuery(test.queryStr).WithArgs(test.args).WillReturnRows(rows)

		t.Run(test.name, func(t *testing.T) {
			userURLs, err := testDB.GetAllUserURLs(context.Background(), cfg.URL, test.args)
			assert.NoError(t, err)
			assert.EqualValues(t, userURLs, []UserURL{{ShortURL: "EwH", OriginalURL: "https://practicum.yandex.ru/"}})
		})
	}
}

func TestDBPingStor(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error occurred while creating mock: %s", err)
	}
	defer db.Close()

	testDB := DBURLs{dbHandle: db}
	t.Run("ping db", func(t *testing.T) {
		err := testDB.PingStor(context.Background())
		assert.NoError(t, err)
	})
}
