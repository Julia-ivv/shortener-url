package main

import (
	"net/http"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
	"github.com/Julia-ivv/shortener-url.git/internal/app/handlers"
	"github.com/Julia-ivv/shortener-url.git/internal/app/logger"
	"github.com/Julia-ivv/shortener-url.git/internal/app/storage"
)

func main() {
	var err error
	config.Flags = *config.NewConfig()
	repo, err := storage.NewURLs(config.Flags)
	defer repo.Close()
	if err != nil {
		logger.ZapSugar.Fatal(err)
	}

	logger.ZapSugar = logger.NewLogger()
	logger.ZapSugar.Infow("Starting server", "addr", config.Flags.Host)
	logger.ZapSugar.Infow("flags",
		"base url", config.Flags.URL,
		"filename", config.Flags.FileName,
	)

	err = http.ListenAndServe(config.Flags.Host, handlers.NewURLRouter(repo))
	if err != nil {
		logger.ZapSugar.Fatalw(err.Error(), "event", "start server")
	}
}
