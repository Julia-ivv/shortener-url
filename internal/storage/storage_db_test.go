package storage

import (
	"context"
	"database/sql/driver"
	"errors"
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

func TestDBGetStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error occurred while creating mock: %s", err)
	}
	defer db.Close()

	testDB := DBURLs{dbHandle: db}

	tests := []struct {
		name           string
		queryStr       string
		args           string
		expectedRows   []string
		expectedValues []driver.Value
		expectedStats  ServiceStats
	}{
		{
			name:           "get stats",
			queryStr:       "SELECT COUNT(.+) AS urls, COUNT(.+) AS users FROM urls",
			expectedRows:   []string{"urls", "users"},
			expectedValues: []driver.Value{2, 1},
			expectedStats:  ServiceStats{URLs: 2, Users: 1},
		},
	}

	for _, test := range tests {
		rows := sqlmock.NewRows(test.expectedRows).AddRow(test.expectedValues...)

		mock.ExpectQuery(test.queryStr).WithoutArgs().WillReturnRows(rows)

		t.Run(test.name, func(t *testing.T) {
			stats, err := testDB.GetStats(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, test.expectedStats, stats)
		})
	}
}

func TestDBAddURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error occurred while creating mock: %s", err)
	}
	defer db.Close()

	testDB := DBURLs{dbHandle: db}
	type mockBehavior func(short string, origin string, id int)

	tests := []struct {
		name            string
		testShortURL    string
		testOriginalURL string
		mockBehavior    mockBehavior
		wantErr         bool
	}{
		{
			name:            "add url ok",
			testShortURL:    "EwH",
			testOriginalURL: "https://practicum.yandex.ru/",
			mockBehavior: func(short string, origin string, id int) {
				mock.ExpectExec("INSERT INTO urls").
					WithArgs([]driver.Value{testUserID, short, origin}...).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:            "insert error",
			testShortURL:    "EwH",
			testOriginalURL: "https://practicum.yandex.ru/",
			mockBehavior: func(short string, origin string, id int) {
				mock.ExpectExec("INSERT INTO urls").
					WithArgs([]driver.Value{testUserID, short, origin}...).
					WillReturnError(errors.New("some error"))
			},
			wantErr: true,
		},
		{
			name:            "insert error rows",
			testShortURL:    "EwH",
			testOriginalURL: "https://practicum.yandex.ru/",
			mockBehavior: func(short string, origin string, id int) {
				mock.ExpectExec("INSERT INTO urls").
					WithArgs([]driver.Value{testUserID, short, origin}...).
					WillReturnResult(sqlmock.NewResult(2, 2))
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(test.testShortURL, test.testOriginalURL, testUserID)
			_, err := testDB.AddURL(context.Background(), test.testShortURL, test.testOriginalURL, testUserID)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBAddBatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error occurred while creating mock: %s", err)
	}
	defer db.Close()

	testDB := DBURLs{dbHandle: db}

	type mockBehavior func(tResp []ResponseBatch, tReq []RequestBatch, id int)

	tests := []struct {
		name              string
		mockBehavior      mockBehavior
		testRequestBatch  []RequestBatch
		testResponseBatch []ResponseBatch
		wantErr           bool
	}{
		{
			name: "add batch OK",
			testRequestBatch: []RequestBatch{
				{
					CorrelationID: "ind1",
					OriginalURL:   "https://pract.ru/url1",
				},
				{
					CorrelationID: "ind2",
					OriginalURL:   "https://pract.ru/url2",
				},
			},
			testResponseBatch: []ResponseBatch{
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
			},
			mockBehavior: func(tResp []ResponseBatch, tReq []RequestBatch, id int) {
				mock.ExpectBegin()
				for k, v := range tResp {
					mock.ExpectExec("INSERT INTO urls").
						WithArgs([]driver.Value{id, v.ShortURL, tReq[k].OriginalURL}...).
						WillReturnResult(sqlmock.NewResult(1, 1))
				}
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			name: "insert error",
			testRequestBatch: []RequestBatch{
				{
					CorrelationID: "ind1",
					OriginalURL:   "https://pract.ru/url1",
				},
				{
					CorrelationID: "ind2",
					OriginalURL:   "https://pract.ru/url2",
				},
			},
			testResponseBatch: []ResponseBatch{
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
			},
			mockBehavior: func(tResp []ResponseBatch, tReq []RequestBatch, id int) {
				mock.ExpectBegin()
				for k, v := range tResp {
					mock.ExpectExec("INSERT INTO urls").
						WithArgs([]driver.Value{id, v.ShortURL, tReq[k].OriginalURL}...).
						WillReturnError(errors.New("some error"))
				}
				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "add error, rows",
			testRequestBatch: []RequestBatch{
				{
					CorrelationID: "ind1",
					OriginalURL:   "https://pract.ru/url1",
				},
				{
					CorrelationID: "ind2",
					OriginalURL:   "https://pract.ru/url2",
				},
			},
			testResponseBatch: []ResponseBatch{
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
			},
			mockBehavior: func(tResp []ResponseBatch, tReq []RequestBatch, id int) {
				mock.ExpectBegin()
				for k, v := range tResp {
					mock.ExpectExec("INSERT INTO urls").
						WithArgs([]driver.Value{id, v.ShortURL, tReq[k].OriginalURL}...).
						WillReturnResult(sqlmock.NewResult(2, 2))
					mock.ExpectRollback()
				}
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.mockBehavior(test.testResponseBatch, test.testRequestBatch, testUserID)

			err := testDB.AddBatch(context.Background(), test.testResponseBatch, test.testRequestBatch, testUserID)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
