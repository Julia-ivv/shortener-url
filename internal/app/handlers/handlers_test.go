package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
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
	shortURL    string
	originURL   string
	deletedFlag bool
	userID      int
}

type testURLs struct {
	originalURLs []testURL
}

func (urls *testURLs) DeleteUserURLs(ctx context.Context, delURLs []string, userID int) (err error) {
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

func (urls *testURLs) GetURL(ctx context.Context, shortURL string) (originURL string, isDel bool, ok bool) {
	for _, v := range urls.originalURLs {
		if v.shortURL == shortURL {
			return v.originURL, v.deletedFlag, true
		}
	}
	return "", false, false
}

func (urls *testURLs) AddURL(ctx context.Context, shortURL string, originURL string, userID int) (err error) {
	inc++
	short := strconv.Itoa(inc)
	urls.originalURLs = append(urls.originalURLs, testURL{
		userID:    userID,
		shortURL:  short,
		originURL: originURL,
	})
	return nil
}

func (urls *testURLs) AddBatch(ctx context.Context, shortURLBatch []storage.ResponseBatch, originURLBatch []storage.RequestBatch, userID int) (err error) {
	allUrls := make([]testURL, len(originURLBatch))
	for k, v := range shortURLBatch {
		allUrls = append(allUrls, testURL{
			userID:    userID,
			shortURL:  v.ShortURL,
			originURL: originURLBatch[k].OriginalURL,
		})
	}

	urls.originalURLs = append(urls.originalURLs, allUrls...)
	return nil
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
	if urls == nil {
		return errors.New("storage storage does not exist")
	}
	return nil
}

func (urls *testURLs) Close() (err error) {
	return nil
}

func (urls *testURLs) GetStats(ctx context.Context) (stats storage.ServiceStats, err error) {
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
	req.Header.Add("X-Real-IP", "192.168.0.1")

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
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("without context", func(t *testing.T) {
		w := httptest.NewRecorder()
		hs.PostURL(w, httptest.NewRequest("GET", ts.URL+"/", nil))
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, 500, res.StatusCode)
	})

	router.Post("/", AddContext(hs.PostURL))
	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name        string
		path        string
		originalURL string
		want        want
		userID      int
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
		userID:      testUserID,
		shortURL:    "EwH",
		deletedFlag: false,
		originURL:   "https://practicum.yandex.ru/",
	})
	testR = append(testR, testURL{
		userID:      testUserID,
		shortURL:    "Eorp",
		deletedFlag: true,
		originURL:   "https://yandex.ru/",
	})
	testRepo := &testURLs{originalURLs: testR}

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
	ts := httptest.NewServer(router)
	defer ts.Close()

	router.Get("/{shortURL}", AddContext(hs.GetURL))
	type want struct {
		originURL  string
		statusCode int
	}
	tests := []struct {
		name   string
		path   string
		want   want
		userID int
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
		{
			name:   "url deleted",
			path:   "/Eorp",
			userID: testUserID,
			want:   want{statusCode: 410, originURL: ""},
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

func TestHandlerDeleteUserURLs(t *testing.T) {
	path := "/api/user/urls"
	testR := make([]testURL, 0)
	testR = append(testR, testURL{
		userID:      testUserID,
		shortURL:    "EwH",
		originURL:   "https://practicum.yandex.ru/",
		deletedFlag: false,
	})
	testR = append(testR, testURL{
		userID:      testUserID,
		shortURL:    "Ert",
		originURL:   "https://mail.ru/",
		deletedFlag: false,
	})

	testRepo := &testURLs{originalURLs: testR}

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("without context", func(t *testing.T) {
		w := httptest.NewRecorder()
		hs.DeleteUserURLs(w, httptest.NewRequest("DELETE", path, nil))
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, 500, res.StatusCode)
	})

	router.Delete("/api/user/urls", AddContext(hs.DeleteUserURLs))
	tests := []struct {
		name       string
		body       string
		userID     int
		wantStatus int
	}{
		{
			name:       "with body",
			body:       `["EwH"]`,
			userID:     testUserID,
			wantStatus: 202,
		},
		{
			name:       "empty body",
			body:       "",
			userID:     testUserID,
			wantStatus: 400,
		},
		{
			name:       "bad body",
			body:       "ghgh",
			userID:     testUserID,
			wantStatus: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, "DELETE", path, strings.NewReader(test.body), test.userID)
			defer resp.Body.Close()
			assert.Equal(t, test.wantStatus, resp.StatusCode)
		})
	}
}

func TestHandlerPostJSON(t *testing.T) {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("without context", func(t *testing.T) {
		w := httptest.NewRecorder()
		hs.PostJSON(w, httptest.NewRequest("POST", ts.URL+"/api/shorten", nil))
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, 500, res.StatusCode)
	})

	router.Post("/api/shorten", AddContext(hs.PostJSON))
	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name   string
		path   string
		body   string
		want   want
		userID int
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
	testRepo := &testURLs{originalURLs: nil}

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("without context", func(t *testing.T) {
		w := httptest.NewRecorder()
		hs.GetUserURLs(w, httptest.NewRequest("GET", ts.URL+"/api/user/urls", nil))
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, 500, res.StatusCode)
	})
	router.Get("/api/user/urls", AddContext(hs.GetUserURLs))

	type want struct {
		userURLs   string
		statusCode int
	}
	type tests struct {
		name   string
		path   string
		want   want
		userID int
	}
	test := tests{
		name:   "no content",
		path:   "/api/user/urls",
		userID: 888,
		want:   want{statusCode: 401, userURLs: ""},
	}

	t.Run(test.name, func(t *testing.T) {
		resp, _ := testRequest(t, ts, "GET", test.path, nil, test.userID)
		defer resp.Body.Close()
		assert.Equal(t, test.want.statusCode, resp.StatusCode)
	})

	testR := make([]testURL, 0)
	testR = append(testR, testURL{
		userID:      testUserID,
		shortURL:    "EwH",
		deletedFlag: false,
		originURL:   "https://practicum.yandex.ru/",
	})
	testRepo.originalURLs = testR
	test = tests{
		name:   "url exists in repository",
		path:   "/api/user/urls",
		userID: testUserID,
		want: want{statusCode: 200, userURLs: `[{
				"short_url": "http://localhost:8080/EwH",
				"original_url": "https://practicum.yandex.ru/"
			}]`},
	}

	t.Run(test.name, func(t *testing.T) {
		resp, getBody := testRequest(t, ts, "GET", test.path, nil, test.userID)
		defer resp.Body.Close()
		assert.Equal(t, test.want.statusCode, resp.StatusCode)
		assert.True(t, assert.NotEmpty(t, getBody))
	})
}

func TestHandlerPostBatch(t *testing.T) {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("without context", func(t *testing.T) {
		w := httptest.NewRecorder()
		hs.PostBatch(w, httptest.NewRequest("POST", ts.URL+"/api/shorten/batch", nil))
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, 500, res.StatusCode)
	})

	router.Post("/api/shorten/batch", AddContext(hs.PostBatch))
	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name   string
		path   string
		body   string
		want   want
		userID int
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
			name:   "test with empty batch",
			path:   "/api/shorten/batch",
			body:   "[]",
			userID: testUserID,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
		},
		{
			name:   "test with empty body",
			path:   "/api/shorten/batch",
			body:   "",
			userID: testUserID,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
		},
		{
			name: "bad batch",
			path: "/api/shorten/batch",
			body: `[
				{
					cor: "ind1",
					origin: "https://pract.ru/url1"
				},
				{
					cor: "ind2",
					origin: "https://pract.ru/url2"
				}
			]`,
			userID: testUserID,
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  500,
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

func TestPing(t *testing.T) {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}
	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
	ts := httptest.NewServer(router)
	defer ts.Close()

	t.Run("ping", func(t *testing.T) {
		w := httptest.NewRecorder()
		hs.GetPingDB(w, httptest.NewRequest("GET", ts.URL+"/ping", nil))
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, 200, res.StatusCode)
	})

	testRepo = nil
	hs = NewHandlers(testRepo, cfg, &sync.WaitGroup{})
	ts = httptest.NewServer(router)
	defer ts.Close()
	t.Run("no ping", func(t *testing.T) {
		w := httptest.NewRecorder()
		hs.GetPingDB(w, httptest.NewRequest("GET", ts.URL+"/ping", nil))
		res := w.Result()
		defer res.Body.Close()
		assert.Equal(t, 500, res.StatusCode)
	})
}

func TestNewURLRouter(t *testing.T) {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}
	t.Run("create router", func(t *testing.T) {
		res := NewURLRouter(testRepo, cfg, &sync.WaitGroup{})
		assert.NotEmpty(t, res)
	})
}

func TestHandlerGetStats(t *testing.T) {
	testR := make([]testURL, 0)
	testR = append(testR, testURL{
		userID:      testUserID,
		shortURL:    "EwH",
		deletedFlag: false,
		originURL:   "https://practicum.yandex.ru/",
	})
	testR = append(testR, testURL{
		userID:      testUserID,
		shortURL:    "Eorp",
		deletedFlag: false,
		originURL:   "https://yandex.ru/",
	})
	testRepo := &testURLs{originalURLs: testR}

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
	ts := httptest.NewServer(router)
	defer ts.Close()

	path := "/api/internal/stats"
	router.Get(path, AddContext(hs.GetStats))
	type want struct {
		urls       int
		users      int
		statusCode int
	}
	tests := []struct {
		name          string
		path          string
		trustedSubnet string
		want          want
		userID        int
	}{
		{
			name:          "status OK",
			path:          path,
			trustedSubnet: "192.168.0.0/24",
			userID:        testUserID,
			want:          want{statusCode: 200, urls: 2, users: 1},
		},
		{
			name:          "status forbidden",
			path:          path,
			trustedSubnet: "192.168.1.0/24",
			userID:        testUserID,
			want:          want{statusCode: 403, urls: 0, users: 0},
		},
		{
			name:          "status forbidden, empty subnet",
			path:          path,
			trustedSubnet: "",
			userID:        testUserID,
			want:          want{statusCode: 403, urls: 0, users: 0},
		},
		{
			name:          "wrong subnet",
			path:          path,
			trustedSubnet: "19216810/24",
			userID:        testUserID,
			want:          want{statusCode: 500, urls: 0, users: 0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hs.cfg.TrustedSubnet = test.trustedSubnet
			resp, _ := testRequest(t, ts, "GET", test.path, nil, test.userID)
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
		})
	}
}

func BenchmarkPostURL(b *testing.B) {
	testRepo := &testURLs{originalURLs: make([]testURL, 0)}
	path := "/"

	router := chi.NewRouter()
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
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
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
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
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
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
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
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
	hs := NewHandlers(testRepo, cfg, &sync.WaitGroup{})
	router.Get(path, AddContext(hs.GetUserURLs))
	ts := httptest.NewServer(router)
	defer ts.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchRequest(ts, http.MethodGet, path, nil, testUserID)
	}
}
