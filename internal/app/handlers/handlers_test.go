package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Julia-ivv/shortener-url.git/internal/app/authorizer"
	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
	"github.com/Julia-ivv/shortener-url.git/internal/app/storage"
)

const testUserID = 123

var inc int
var cfg config.Flags

func Init() {
	cfg = *config.NewConfig()
}

type testURL struct {
	userID    int
	shortURL  string
	originURL string
}

type testURLs struct {
	originalURLs []testURL
}

func (urls *testURLs) DeleteUserURLs(ctx context.Context, delURLs []string, userID int) (err error) {
	return nil
}

func (urls *testURLs) GetURL(ctx context.Context, shortURL string) (originURL string, isDel bool, ok bool) {
	for _, v := range urls.originalURLs {
		if v.shortURL == shortURL {
			return v.originURL, false, true
		}
	}
	return "", false, false
}

func (urls *testURLs) AddURL(ctx context.Context, originURL string, userID int) (shortURL string, err error) {
	// добавить новый урл
	inc++
	short := strconv.Itoa(inc)
	urls.originalURLs = append(urls.originalURLs, testURL{
		userID:    userID,
		shortURL:  short,
		originURL: originURL,
	})
	return short, nil
}

func (urls *testURLs) AddBatch(ctx context.Context, originURLBatch []storage.RequestBatch, baseURL string, userID int) (shortURLBatch []storage.ResponseBatch, err error) {
	allUrls := make([]testURL, len(originURLBatch))
	shortURLBatch = make([]storage.ResponseBatch, len(originURLBatch))
	for _, v := range originURLBatch {
		sURL := strconv.Itoa(inc)
		shortURLBatch = append(shortURLBatch, storage.ResponseBatch{
			CorrelationID: v.CorrelationID,
			ShortURL:      baseURL + sURL,
		})
		allUrls = append(allUrls, testURL{
			userID:    userID,
			shortURL:  sURL,
			originURL: v.OriginalURL,
		})
	}

	urls.originalURLs = append(urls.originalURLs, allUrls...)
	return shortURLBatch, nil
}

func (urls *testURLs) GetAllUserURLs(ctx context.Context, baseURL string, userID int) (userURLs []storage.UserURL, err error) {

	for _, v := range urls.originalURLs {
		if v.userID == userID {
			userURLs = append(userURLs, storage.UserURL{
				ShortURL:    baseURL + v.shortURL,
				OriginalURL: v.originURL,
			})
		}
	}

	return userURLs, nil
}

func (urls *testURLs) PingStor(ctx context.Context) (err error) {
	return nil
}

func (urls *testURLs) Close() (err error) {
	return nil
}

func AddContext(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			newctx := context.WithValue(req.Context(), authorizer.UserContextKey, testUserID)
			h.ServeHTTP(res, req.WithContext(newctx))
		})
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader, userID int) (*http.Response, string) {
	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequestWithContext(context.WithValue(context.Background(), authorizer.UserContextKey, userID),
		method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func benchRequest(ts *httptest.Server, method, path string, body io.Reader, userID int) error {
	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequestWithContext(context.WithValue(context.Background(), authorizer.UserContextKey, userID),
		method, ts.URL+path, body)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func TestHandlerPostURL(t *testing.T) {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Post("/", AddContext(hs.PostURL))
	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name        string
		path        string
		originalURL string // URL в теле запроса
		userID      int
		want        want
	}{
		{
			name:        "URL added successfully",
			path:        "/",
			originalURL: "https://mail.ru/",
			userID:      testUserID,
			want: want{
				contentType: "text/plain",
				statusCode:  201,
			},
		},
		{
			name:        "test with empty body",
			path:        "/",
			originalURL: "",
			userID:      testUserID,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, getBody := testRequest(t, ts, "POST", test.path, strings.NewReader(test.originalURL), test.userID)
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))
			assert.True(t, assert.NotEmpty(t, getBody))
		})
	}
}

func TestHandlerGetURL(t *testing.T) {
	testR := make([]testURL, 0)
	testR = append(testR, testURL{
		userID:    testUserID,
		shortURL:  "EwH",
		originURL: "https://practicum.yandex.ru/",
	})
	testRepo := &testURLs{originalURLs: testR}

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Get("/{shortURL}", AddContext(hs.GetURL))
	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		statusCode int
		originURL  string
	}
	tests := []struct {
		name   string
		path   string
		userID int
		want   want
	}{
		{
			name:   "url exists in repository",
			path:   "/EwH",
			userID: testUserID,
			want:   want{statusCode: 307, originURL: "https://practicum.yandex.ru/"},
		},
		{
			name:   "url does not exist in repository",
			path:   "/11",
			userID: testUserID,
			want:   want{statusCode: 400, originURL: ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, "GET", test.path, nil, test.userID)
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
			assert.Equal(t, test.want.originURL, resp.Header.Get("Location"))
		})
	}
}

func TestHandlerPostJSON(t *testing.T) {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Post("/api/shorten", AddContext(hs.PostJSON))
	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name   string
		path   string
		body   string
		userID int
		want   want
	}{
		{
			name:   "URL added successfully",
			path:   "/api/shorten",
			body:   `{"url":"https://mail.ru"}`,
			userID: testUserID,
			want: want{
				contentType: "application/json",
				statusCode:  201,
			},
		},
		{
			name:   "test with empty body",
			path:   "/api/shorten",
			body:   "",
			userID: testUserID,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, getBody := testRequest(t, ts, "POST", test.path, strings.NewReader(test.body), test.userID)
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))
			assert.True(t, assert.NotEmpty(t, getBody))
		})
	}
}

func TestHandlerGetUserURLs(t *testing.T) {
	testR := make([]testURL, 0)
	testR = append(testR, testURL{
		userID:    testUserID,
		shortURL:  "EwH",
		originURL: "https://practicum.yandex.ru/",
	})
	testRepo := &testURLs{originalURLs: testR}

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Get("/api/user/urls", AddContext(hs.GetUserURLs))
	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		statusCode int
		userURLs   string
	}
	tests := []struct {
		name   string
		path   string
		userID int
		want   want
	}{
		{
			name:   "url exists in repository",
			path:   "/api/user/urls",
			userID: testUserID,
			want: want{statusCode: 200, userURLs: `[{
				"short_url": "http://localhost:8080/EwH",
				"original_url": "https://practicum.yandex.ru/"
			}]`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, getBody := testRequest(t, ts, "GET", test.path, nil, test.userID)
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
			assert.True(t, assert.NotEmpty(t, getBody))
		})
	}
}

func TestHandlerPostBatch(t *testing.T) {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Post("/api/shorten/batch", AddContext(hs.PostBatch))
	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name   string
		path   string
		body   string
		userID int
		want   want
	}{
		{
			name: "URLs added successfully",
			path: "/api/shorten/batch",
			body: `[
				{
					"correlation_id": "ind1",
					"original_url": "https://pract.ru/url1"
				},
				{
					"correlation_id": "ind2",
					"original_url": "https://pract.ru/url2"
				}
			]`,
			userID: testUserID,
			want: want{
				contentType: "application/json",
				statusCode:  201,
			},
		},
		{
			name:   "test with empty body",
			path:   "/api/shorten/batch",
			body:   "[]",
			userID: testUserID,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, getBody := testRequest(t, ts, "POST", test.path, strings.NewReader(test.body), test.userID)
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))
			assert.True(t, assert.NotEmpty(t, getBody))
		})
	}
}

func BenchmarkPostURL(b *testing.B) {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}
	path := "/"

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Post(path, AddContext(hs.PostURL))
	ts := httptest.NewServer(router)
	defer ts.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchRequest(ts, http.MethodPost, path, strings.NewReader("https://mail.ru/"), testUserID)
	}
}

func BenchmarkPostJSON(b *testing.B) {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}
	path := "/api/shorten"

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Post(path, AddContext(hs.PostJSON))
	ts := httptest.NewServer(router)
	defer ts.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchRequest(ts, http.MethodPost, path, strings.NewReader(`{"url":"https://mail.ru"}`), testUserID)
	}
}

func BenchmarkPostBatch(b *testing.B) {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}
	path := "/api/shorten/batch"

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Post(path, AddContext(hs.PostBatch))
	ts := httptest.NewServer(router)
	defer ts.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchRequest(ts, http.MethodPost, path,
			strings.NewReader(`[
			{
				"correlation_id": "ind1",
				"original_url": "https://pract.ru/url1"
			},
			{
				"correlation_id": "ind2",
				"original_url": "https://pract.ru/url2"
			},
			{
				"correlation_id": "ind3",
				"original_url": "https://pract.ru/url3"
			},
			{
				"correlation_id": "ind4",
				"original_url": "https://pract.ru/url4"
			}
		]`), testUserID)
	}
}

func BenchmarkGetURL(b *testing.B) {
	testR := make([]testURL, 0)
	testR = append(testR, testURL{
		userID:    testUserID,
		shortURL:  "EwH",
		originURL: "https://practicum.yandex.ru/",
	})
	testRepo := &testURLs{originalURLs: testR}
	path := "/"

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Get(path+"{shortURL}", AddContext(hs.GetURL))
	ts := httptest.NewServer(router)
	defer ts.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchRequest(ts, http.MethodGet, path+"EwH", nil, testUserID)
	}
}

func BenchmarkGetUserURLs(b *testing.B) {
	testR := make([]testURL, 0)
	testR = append(testR, testURL{
		userID:    testUserID,
		shortURL:  "EwH",
		originURL: "https://practicum.yandex.ru/",
	})
	testRepo := &testURLs{originalURLs: testR}
	path := "/api/user/urls"

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg)
	router.Get(path, AddContext(hs.GetUserURLs))
	ts := httptest.NewServer(router)
	defer ts.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchRequest(ts, http.MethodGet, path, nil, testUserID)
	}
}
