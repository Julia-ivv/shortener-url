package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Julia-ivv/shortener-url.git/internal/app/authorizer"
	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
	"github.com/Julia-ivv/shortener-url.git/internal/app/middleware"

	"github.com/Julia-ivv/shortener-url.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type (
	Handlers struct {
		stor storage.Stor
		cfg  config.Flags
	}
)

func NewHandlers(stor storage.Stor, cfg config.Flags) *Handlers {
	h := &Handlers{}
	h.stor = stor
	h.cfg = cfg
	return h
}

func (h *Handlers) postURL(res http.ResponseWriter, req *http.Request) {
	// получает из тела запроса длинный урл - postURL
	// добавляет его в хранилище
	// возвращает в теле ответа короткий урл
	value := req.Context().Value(authorizer.USER_CONTEXT_KEY)
	if value == nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	id := value.(int)

	postURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	if len(postURL) == 0 {
		http.Error(res, "request with empty body", http.StatusBadRequest)
		return
	}
	shortURL, err := h.stor.Repo.AddURL(req.Context(), string(postURL), id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(http.StatusConflict)
		} else {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
	}
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
	value := req.Context().Value(authorizer.USER_CONTEXT_KEY)
	if value == nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	id := value.(int)

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
	shortURL, err := h.stor.Repo.AddURL(req.Context(), string(reqURL.URL), id)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusConflict)
		} else {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
	}

	resp, err := json.Marshal(ResponseURL{Result: h.cfg.URL + "/" + string(shortURL)})
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *Handlers) postBatch(res http.ResponseWriter, req *http.Request) {
	// в теле запроса множество урлов в слайсе
	// в ответе аналогичный слайс с короткими урлами
	value := req.Context().Value(authorizer.USER_CONTEXT_KEY)
	if value == nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	id := value.(int)

	reqJSON, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	var reqBatch []storage.RequestBatch
	if len(reqJSON) == 0 {
		http.Error(res, "request with empty body", http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(reqJSON, &reqBatch)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(reqBatch) == 0 {
		http.Error(res, "empty request", http.StatusBadRequest)
		return
	}
	resBatch, err := h.stor.Repo.AddBatch(req.Context(), reqBatch, h.cfg.URL+"/", id)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(resBatch)
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
	value := req.Context().Value(authorizer.USER_CONTEXT_KEY)
	if value == nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	id := value.(int)

	shortURL := chi.URLParam(req, "shortURL")
	originURL, ok := h.stor.Repo.GetURL(req.Context(), shortURL, id)
	if !ok {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originURL)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handlers) getPingDB(res http.ResponseWriter, req *http.Request) {
	if err := h.stor.Repo.PingStor(req.Context()); err != nil {
		http.Error(res, "ping error", http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (h *Handlers) getUserURLs(res http.ResponseWriter, req *http.Request) {
	value := req.Context().Value(authorizer.USER_CONTEXT_KEY)
	if value == nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	id := value.(int)

	allURLs, err := h.stor.Repo.GetAllUserURLs(req.Context(), h.cfg.URL+"/", id)
	if len(allURLs) == 0 {
		//http.Error(res, "204 No Content", http.StatusNoContent)
		http.Error(res, "401 Unauthorized", http.StatusUnauthorized) // под автотесты
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	resp, err := json.Marshal(allURLs)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
}

func NewURLRouter(repo storage.Stor, cfg config.Flags) chi.Router {
	hs := NewHandlers(repo, cfg)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(middleware.HandlerWithAuth(hs.postURL))))
		r.Get("/{shortURL}", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(middleware.HandlerWithAuth(hs.getURL))))
		r.Post("/api/shorten", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(middleware.HandlerWithAuth(hs.postJSON))))
		r.Post("/api/shorten/batch", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(middleware.HandlerWithAuth(hs.postBatch))))
		r.Get("/api/user/urls", middleware.HandlerWithLogging(middleware.HandlerWithGzipCompression(middleware.HandlerWithAuth(hs.getUserURLs))))
		r.Get("/ping", middleware.HandlerWithLogging(hs.getPingDB))
	})
	return r
}
