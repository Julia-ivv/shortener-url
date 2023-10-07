package main

import (
	"net/http"

	"github.com/Julia-ivv/shortener-url.git/internal/app/handlers"
	"github.com/Julia-ivv/shortener-url.git/internal/app/storage"
)

func main() {
	err := http.ListenAndServe(":8080", handlers.URLRouter(&storage.Repo))
	if err != nil {
		panic(err)
	}
}
