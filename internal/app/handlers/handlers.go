package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
	"github.com/Julia-ivv/shortener-url.git/internal/app/middleware"
	"github.com/Julia-ivv/shortener-url.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

type (
	Handlers struct {
		repo storage.Repositories
		cfg  config.Flags
	}
)

func NewHandlers(repo storage.Repositories, cfg config.Flags) *Handlers {
	h := &Handlers{}
	h.repo = repo
	h.cfg = cfg
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
	shortURL, err := h.repo.AddURL(string(postURL))
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write([]byte(h.cfg.URL + "/" + shortURL))
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
}

type RequestURL struct {
	URL string `json:"url"`
}

type ResponseURL struct {
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
	var reqURL RequestURL
	if len(reqJSON) == 0 {
		http.Error(res, "request with empty body", http.StatusBadRequest)
		return
	}
	json.Unmarshal(reqJSON, &reqURL)
	shortURL, err := h.repo.AddURL(string(reqURL.URL))
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(ResponseURL{Result: h.cfg.URL + "/" + string(shortURL)})
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

func NewURLRouter(repo storage.Repositories, cfg config.Flags) chi.Router {
	hs := NewHandlers(repo, cfg)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(hs.postURL)))
		r.Get("/{shortURL}", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(hs.getURL)))
		r.Post("/api/shorten", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(hs.postJSON)))
	})
	return r
}
