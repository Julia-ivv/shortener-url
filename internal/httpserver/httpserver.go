package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	mwPkg "github.com/Julia-ivv/shortener-url/pkg/middleware"

	"github.com/Julia-ivv/shortener-url.git/internal/authorizer"
	"github.com/Julia-ivv/shortener-url.git/internal/config"
	mwInt "github.com/Julia-ivv/shortener-url.git/internal/middleware"
	"github.com/Julia-ivv/shortener-url.git/internal/storage"
	"github.com/Julia-ivv/shortener-url.git/pkg/randomizer"
)

// Handlers stores the repository and settings of this application.
type Handlers struct {
	stor storage.Repositories
	cfg  config.Flags
	wg   *sync.WaitGroup
}

// NewHandlers creates an instance with storage and settings for handlers.
func NewHandlers(stor storage.Repositories, cfg config.Flags, wg *sync.WaitGroup) *Handlers {
	h := &Handlers{}
	h.stor = stor
	h.cfg = cfg
	h.wg = wg
	return h
}

// PostURL gets a long URL from the request body.
// Adds it to storage, returns a short URL in the response body.
func (h *Handlers) PostURL(res http.ResponseWriter, req *http.Request) {
	value := req.Context().Value(authorizer.UserContextKey)
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

	shortURL, err := randomizer.GenerateRandomString(randomizer.LengthShortURL)
	if err != nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	findURL, err := h.stor.AddURL(req.Context(), shortURL, string(postURL), id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(http.StatusConflict)
			_, err = res.Write([]byte(h.cfg.URL + "/" + findURL))
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		_, err = res.Write([]byte(h.cfg.URL + "/" + shortURL))
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

// RequestURL stores the URL from the request body for the handler PostJSON.
type RequestURL struct {
	URL string `json:"url"`
}

// ResponseURL stores the response URL for the handler PostJSON.
type ResponseURL struct {
	Result string `json:"result"`
}

// PostJSON receives json with the original url from the request.
// Adds it to storage, returns json with the short URL in the response body.
func (h *Handlers) PostJSON(res http.ResponseWriter, req *http.Request) {
	value := req.Context().Value(authorizer.UserContextKey)
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
	err = json.Unmarshal(reqJSON, &reqURL)
	if err != nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}

	shortURL, err := randomizer.GenerateRandomString(randomizer.LengthShortURL)
	if err != nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	findURL, err := h.stor.AddURL(req.Context(), shortURL, string(reqURL.URL), id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusConflict)
			shortURL = findURL
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

// PostBatch gets a slice of the original URLs from the request body.
// Adds it to storage, returns a slice of the short URLs in the response body.
func (h *Handlers) PostBatch(res http.ResponseWriter, req *http.Request) {
	value := req.Context().Value(authorizer.UserContextKey)
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

	resBatch := make([]storage.ResponseBatch, len(reqBatch))
	for k, v := range reqBatch {
		shortURL, err := randomizer.GenerateRandomString(randomizer.LengthShortURL)
		if err != nil {
			http.Error(res, "500 internal server error", http.StatusInternalServerError)
			return
		}
		resBatch[k].CorrelationID = v.CorrelationID
		resBatch[k].ShortURLFull = h.cfg.URL + "/" + shortURL
		resBatch[k].ShortURL = shortURL
	}

	err = h.stor.AddBatch(req.Context(), resBatch, reqBatch, id)
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
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetURL gets a long URL from the storage using shortURL.
// Returns a long URL in Location.
// No selection by user.
func (h *Handlers) GetURL(res http.ResponseWriter, req *http.Request) {
	shortURL := chi.URLParam(req, "shortURL")
	originURL, isDel, ok := h.stor.GetURL(req.Context(), shortURL)
	if !ok {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}
	if isDel {
		res.WriteHeader(http.StatusGone)
		return
	}
	res.Header().Set("Location", originURL)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

// GetPingDB checks storage access.
func (h *Handlers) GetPingDB(res http.ResponseWriter, req *http.Request) {
	if err := h.stor.PingStor(req.Context()); err != nil {
		http.Error(res, "ping error", http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

// GetUserURLs gets all the user's short urls from the repository.
func (h *Handlers) GetUserURLs(res http.ResponseWriter, req *http.Request) {
	value := req.Context().Value(authorizer.UserContextKey)
	if value == nil {
		http.Error(res, "500 internal server error", http.StatusInternalServerError)
		return
	}
	id := value.(int)

	allURLs, err := h.stor.GetAllUserURLs(req.Context(), h.cfg.URL+"/", id)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(allURLs) == 0 {
		http.Error(res, "it should be 204 No Content", http.StatusUnauthorized)
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
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteUserURLs adds a removal flag for URLs from the request body.
func (h *Handlers) DeleteUserURLs(res http.ResponseWriter, req *http.Request) {
	value := req.Context().Value(authorizer.UserContextKey)
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
	if len(reqJSON) == 0 {
		http.Error(res, "request with empty body", http.StatusBadRequest)
		return
	}
	var reqShortURLs []string
	err = json.Unmarshal(reqJSON, &reqShortURLs)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	h.wg.Add(1)
	go func() {
		h.stor.DeleteUserURLs(req.Context(), reqShortURLs, id)
		h.wg.Done()
	}()

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusAccepted)
}

// GetStats gets the amount of all users and URLs in the service.
// Available only for IP addresses from a trusted subnet.
func (h *Handlers) GetStats(res http.ResponseWriter, req *http.Request) {
	if h.cfg.TrustedSubnet == "" {
		http.Error(res, "403 Forbidden, empty trusted subnet", http.StatusForbidden)
		return
	}

	ipStr := req.Header.Get("X-Real-IP")
	ip := net.ParseIP(ipStr)
	_, ipNet, err := net.ParseCIDR(h.cfg.TrustedSubnet)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if !ipNet.Contains(ip) {
		http.Error(res, "403 Forbidden, not trusted IP", http.StatusForbidden)
		return
	}

	stats, err := h.stor.GetStats(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	resp, err := json.Marshal(stats)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
}

// NewURLRouter creates a router instance.
func NewURLRouter(repo storage.Repositories, cfg config.Flags, wg *sync.WaitGroup) chi.Router {
	hs := NewHandlers(repo, cfg, wg)
	r := chi.NewRouter()
	r.Use(mwPkg.HandlerWithLogging, mwPkg.HandlerWithGzipCompression)
	r.Group(func(r chi.Router) {
		r.Use(mwInt.HandlerWithAuth)
		r.Post("/", hs.PostURL)
		r.Get("/{shortURL}", hs.GetURL)
		r.Post("/api/shorten", hs.PostJSON)
		r.Post("/api/shorten/batch", hs.PostBatch)
		r.Get("/api/user/urls", hs.GetUserURLs)
		r.Delete("/api/user/urls", hs.DeleteUserURLs)
	})
	r.Get("/ping", hs.GetPingDB)
	r.Get("/api/internal/stats", hs.GetStats)
	return r
}
