package main

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

var urls map[string]string
var inc int = 100

func main() {
	urls = make(map[string]string)
	urls["EwHXdJfB"] = "https://practicum.yandex.ru/"

	mux := http.NewServeMux()

	mux.HandleFunc(`/`, postURL)

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		panic(err)
	}
}

func postURL(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		if req.URL.Path != "/" {
			http.Error(res, "400 bad request", http.StatusBadRequest)
			return
		}
		postURL, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		inc++
		shortURL := strconv.Itoa(inc)
		urls[shortURL] = string(postURL)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		_, err = res.Write([]byte("http://" + req.Host + "/" + shortURL))
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
	case http.MethodGet:
		parseURL := strings.Split(req.URL.Path, "/")
		shortURL := parseURL[len(parseURL)-1]
		originURL, ok := urls[shortURL]
		if !ok {
			http.Error(res, "URL not found", http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", originURL)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusTemporaryRedirect)
	default:
		res.WriteHeader(http.StatusBadRequest)
	}
}
