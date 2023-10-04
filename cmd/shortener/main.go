package main

import (
	"net/http"

	"github.com/Julia-ivv/shortener-url.git/internal/app/handlers"
)

func main() {
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
		handlers.HandlerPost(res, req)
	case http.MethodGet:
		handlers.HandlerGet(res, req)
	default:
		res.WriteHeader(http.StatusBadRequest)
	}
}
