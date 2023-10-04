package handlers

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Julia-ivv/shortener-url.git/internal/app/storage"
)

func HandlerPost(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.Error(res, "400 bad request", http.StatusBadRequest)
		return
	}
	postURL, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	storage.Inc++
	shortURL := strconv.Itoa(storage.Inc)
	storage.Urls[shortURL] = string(postURL)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write([]byte("http://" + req.Host + "/" + shortURL))
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	return
}

func HandlerGet(res http.ResponseWriter, req *http.Request) {
	parseURL := strings.Split(req.URL.Path, "/")
	shortURL := parseURL[len(parseURL)-1]
	originURL, ok := storage.Urls[shortURL]
	if !ok {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originURL)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusTemporaryRedirect)
}
