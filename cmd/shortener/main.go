package main

import (
	"net/http"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
	"github.com/Julia-ivv/shortener-url.git/internal/app/handlers"
	"github.com/Julia-ivv/shortener-url.git/internal/app/storage"
)

func main() {
	config.Flags = config.InitConfigFromFlags()

	err := http.ListenAndServe(config.Flags.Host, handlers.URLRouter(&storage.Repo))
	if err != nil {
		panic(err)
	}
}
