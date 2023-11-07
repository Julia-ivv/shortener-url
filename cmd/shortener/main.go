package main

import (
	"net/http"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
	"github.com/Julia-ivv/shortener-url.git/internal/app/handlers"
	"github.com/Julia-ivv/shortener-url.git/internal/app/logger"
	"github.com/Julia-ivv/shortener-url.git/internal/app/storage"
)

func main() {
	cfg := config.NewConfig()

	logger.ZapSugar = logger.NewLogger()
	logger.ZapSugar.Infow("Starting server", "addr", cfg.Host)
	logger.ZapSugar.Infow("flags",
		"base url", cfg.URL,
		"filename", cfg.FileName,
	)

	repo, err := storage.NewURLs(*cfg)
	if err != nil {
		logger.ZapSugar.Fatal(err)
	}

	err = http.ListenAndServe(cfg.Host, handlers.NewURLRouter(repo, *cfg))
	if err != nil {
		logger.ZapSugar.Fatalw(err.Error(), "event", "start server")
	}
}
