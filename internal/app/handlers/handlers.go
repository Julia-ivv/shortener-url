package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Julia-ivv/shortener-url.git/internal/app/compressing"
	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
	"github.com/Julia-ivv/shortener-url.git/internal/app/logger"
	"github.com/Julia-ivv/shortener-url.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

type (
	Handlers struct {
		repo storage.Repositories
	}
)

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
	if len(postURL) == 0 {
		http.Error(res, "request with empty body", http.StatusBadRequest)
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

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func (h *Handlers) postJSON(res http.ResponseWriter, req *http.Request) {
	// в теле запроса JSON с длинным урлом
	// в ответе JSON с коротким урлом
	reqJSON, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	var reqURL Request
	if len(reqJSON) == 0 {
		http.Error(res, "request with empty body", http.StatusBadRequest)
		return
	}
	json.Unmarshal(reqJSON, &reqURL)
	shortURL := h.repo.AddURL(string(reqURL.URL))

	resp, err := json.Marshal(Response{Result: config.Flags.URL + "/" + string(shortURL)})
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write(resp)
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
		r.Post("/", logger.HandlerWithLogging(compressing.HandlerWithGzipCompression(hs.postURL)))
		r.Get("/{shortURL}", logger.HandlerWithLogging(compressing.HandlerWithGzipCompression(hs.getURL)))
		r.Post("/api/shorten", logger.HandlerWithLogging(compressing.HandlerWithGzipCompression(hs.postJSON)))
	})
	return r
}
