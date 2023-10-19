package handlers

import (
	"io"
	"net/http"
	"time"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
	"github.com/Julia-ivv/shortener-url.git/internal/app/logger"
	"github.com/Julia-ivv/shortener-url.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

type (
	Handlers struct {
		repo storage.Repositories
	}

	responseInfo struct {
		status int
		size   int
	}

	logResponseWriter struct {
		http.ResponseWriter
		responseInfo *responseInfo
	}
)

func (res *logResponseWriter) Write(b []byte) (int, error) {
	size, err := res.ResponseWriter.Write(b)
	res.responseInfo.size += size
	return size, err
}

func (res *logResponseWriter) WriteHeader(statusCode int) {
	res.ResponseWriter.WriteHeader(statusCode)
	res.responseInfo.status = statusCode
}

func HandlerWithLogging(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			start := time.Now()
			responseInfo := &responseInfo{
				status: 0,
				size:   0,
			}
			logResponseWriter := logResponseWriter{
				ResponseWriter: res,
				responseInfo:   responseInfo,
			}
			uri := req.RequestURI
			method := req.Method

			h(&logResponseWriter, req)
			duration := time.Since(start)

			logger.ZapSugar.Infoln(
				"uri", uri,
				"method", method,
				"status", responseInfo.status,
				"size", responseInfo.size,
				"duration", duration,
			)
		})
}

func NewHandlers(repo storage.Repositories) *Handlers {
	h := &Handlers{}
	h.repo = repo
	return h
}

func (h *Handlers) postURL(res http.ResponseWriter, req *http.Request) {
	// получает из тела запроса длинный урл - postURL
	// добавляет его в хранилище
	// возвращает в теле ответа короткий урл
	postURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	shortURL := h.repo.AddURL(string(postURL))
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write([]byte(config.Flags.URL + "/" + shortURL))
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handlers) getURL(res http.ResponseWriter, req *http.Request) {
	// получает из хранилища длинный урл по shortURL из параметра запроса
	// возвращает длинный урл в Location
	shortURL := chi.URLParam(req, "shortURL")
	originURL, ok := h.repo.GetURL(shortURL)
	if !ok {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originURL)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func NewURLRouter() chi.Router {
	hs := NewHandlers(&storage.Repo)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", HandlerWithLogging(hs.postURL))
		r.Get("/{shortURL}", HandlerWithLogging(hs.getURL))
	})
	return r
}
