package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var inc int

type testURLs struct {
	originalURLs map[string]string
}

var testRepo testURLs

func (urls *testURLs) GetURL(shortURL string) (originURL string, ok bool) {
	// получить длинный урл
	originURL, ok = urls.originalURLs[shortURL]
	return originURL, ok
}

func (urls *testURLs) AddURL(originURL string) (shortURL string, err error) {
	// добавить новый урл
	inc++
	short := strconv.Itoa(inc)
	urls.originalURLs[short] = originURL
	return short, nil
}

func (urls *testURLs) Close() error {
	return nil
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestHandlerPost(t *testing.T) {
	testRepo.originalURLs = make(map[string]string)

	router := chi.NewRouter()
	hs := NewHandlers(&testRepo)
	router.Post("/", hs.postURL)
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
		want        want
	}{
		{
			name:        "URL added successfully",
			path:        "/",
			originalURL: "https://mail.ru/",
			want: want{
				contentType: "text/plain",
				statusCode:  201,
			},
		},
		{
			name:        "test with empty body",
			path:        "/",
			originalURL: "",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, getBody := testRequest(t, ts, "POST", test.path, strings.NewReader(test.originalURL))
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))
			assert.True(t, assert.NotEmpty(t, getBody))
		})
	}
}

func TestHandlerGet(t *testing.T) {
	testRepo.originalURLs = make(map[string]string)
	testRepo.originalURLs["EwH"] = "https://practicum.yandex.ru/"

	router := chi.NewRouter()
	hs := NewHandlers(&testRepo)
	router.Get("/{shortURL}", hs.getURL)
	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		statusCode int
		originURL  string
	}
	tests := []struct {
		name string
		path string
		want want
	}{
		{
			name: "url exists in repository",
			path: "/EwH",
			want: want{statusCode: 307, originURL: "https://practicum.yandex.ru/"},
		},
		{
			name: "url does not exist in repository",
			path: "/11",
			want: want{statusCode: 400, originURL: ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, _ := testRequest(t, ts, "GET", test.path, nil)
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
			assert.Equal(t, test.want.originURL, resp.Header.Get("Location"))
		})
	}
}

func TestHandlerPostJSON(t *testing.T) {
	testRepo.originalURLs = make(map[string]string)

	router := chi.NewRouter()
	hs := NewHandlers(&testRepo)
	router.Post("/api/shorten", hs.postJSON)
	ts := httptest.NewServer(router)
	defer ts.Close()

	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name string
		path string
		body string
		want want
	}{
		{
			name: "URL added successfully",
			path: "/api/shorten",
			body: `{"url":"https://mail.ru"}`,
			want: want{
				contentType: "application/json",
				statusCode:  201,
			},
		},
		{
			name: "test with empty body",
			path: "/api/shorten",
			body: "",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, getBody := testRequest(t, ts, "POST", test.path, strings.NewReader(test.body))
			defer resp.Body.Close()
			assert.Equal(t, test.want.statusCode, resp.StatusCode)
			assert.Equal(t, test.want.contentType, resp.Header.Get("Content-Type"))
			assert.True(t, assert.NotEmpty(t, getBody))
		})
	}
}
