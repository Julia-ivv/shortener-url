// Package main application entry point.
package main

import (
	"fmt"
	"net/http"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
	"github.com/Julia-ivv/shortener-url.git/internal/app/handlers"
	"github.com/Julia-ivv/shortener-url.git/internal/app/storage"
	"github.com/Julia-ivv/shortener-url.git/pkg/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

	cfg := config.NewConfig()

	logger.ZapSugar = logger.NewLogger()
	logger.ZapSugar.Infow("Starting server", "addr", cfg.Host)
	logger.ZapSugar.Infow("flags",
		"base url", cfg.URL,
		"filename", cfg.FileName,
		"db dsn", cfg.DBDSN,
	)

	repo, err := storage.NewURLs(*cfg)
	if err != nil {
		logger.ZapSugar.Fatal(err)
	}

	defer repo.Close()

	err = http.ListenAndServe(cfg.Host, handlers.NewURLRouter(repo, *cfg))
	if err != nil {
		logger.ZapSugar.Fatalw(err.Error(), "event", "start server")
	}
}
