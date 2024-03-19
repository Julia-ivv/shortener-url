package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Julia-ivv/shortener-url.git/internal/authorizer"
)

func hFunc(w http.ResponseWriter, r *http.Request) {
	value := r.Context().Value(authorizer.UserContextKey)
	if value == nil {
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}
	id := value.(int)
	resp, _ := json.Marshal(id)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func TestHandlerWithAuth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h := HandlerWithAuth(http.HandlerFunc(hFunc))
	h.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.NotEmpty(t, body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
