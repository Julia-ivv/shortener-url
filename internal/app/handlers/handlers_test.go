package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerPost(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		body        string
	}
	tests := []struct {
		name        string
		path        string
		originalURL string // URL в body
		want        want
	}{
		{
			name:        "positive test",
			path:        "/",
			originalURL: "https://mail.ru/",
			want:        want{contentType: "text/plain", statusCode: 201, body: "http://example.com/101"},
		},
		{
			name:        "wrong path",
			path:        "/11",
			originalURL: "https://mail.ru/",
			want:        want{contentType: "text/plain; charset=utf-8", statusCode: 400, body: "wrong path\n"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, test.path, strings.NewReader(test.originalURL))
			w := httptest.NewRecorder()
			HandlerPost(w, req)
			res := w.Result()
			assert.Equal(t, test.want.statusCode, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.body, string(resBody))
		})
	}
}

func TestHandlerGet(t *testing.T) {
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
			name: "positive test",
			path: "/EwHXdJfB",
			want: want{statusCode: 307, originURL: "https://practicum.yandex.ru/"},
		},
		{
			name: "not found test",
			path: "/11",
			want: want{statusCode: 400, originURL: ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.path, nil)
			w := httptest.NewRecorder()
			HandlerGet(w, req)
			res := w.Result()
			assert.Equal(t, test.want.statusCode, res.StatusCode)
			assert.Equal(t, test.want.originURL, res.Header.Get("Location"))
		})
	}
}
