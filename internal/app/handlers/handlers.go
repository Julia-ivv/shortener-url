package handlers

import (
	"io"
	"net/http"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
	"github.com/Julia-ivv/shortener-url.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func URLRouter(repo storage.Repositories) chi.Router {
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", HandlerPost(repo))
		r.Get("/{shortURL}", HandlerGet(repo))
	})
	return r
}

func HandlerPost(repo storage.Repositories) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// получает и тела запроса длинный урл - postURL
		// добавляет его в хранилище
		// возвращает в теле ответа короткий урл
		postURL, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		shortURL := repo.AddURL(string(postURL))
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		_, err = res.Write([]byte(config.Flags.URL + "/" + shortURL))
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

func HandlerGet(repo storage.Repositories) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// получает из хранилища длинный урл по shortURL из параметра запроса
		// возвращает длинный урл в Location
		shortURL := chi.URLParam(req, "shortURL")
		originURL, ok := repo.GetURL(shortURL)
		if !ok {
			http.Error(res, "URL not found", http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", originURL)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
