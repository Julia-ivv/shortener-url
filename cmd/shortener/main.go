// Package main application entry point.
package main

import (
	"fmt"
	"net/http"

	"github.com/Julia-ivv/shortener-url.git/cmd/certgenerator"
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
		"https enabled", cfg.EnableHTTPS,
		"config file", cfg.ConfigFileName,
	)

	repo, err := storage.NewURLs(*cfg)
	if err != nil {
		logger.ZapSugar.Fatal(err)
	}

	defer repo.Close()

	certFile, privateKeyFile, err := certgenerator.GenCert(4096)
	if err != nil {
		logger.ZapSugar.Fatalw(err.Error(), "event", "create certificate or private key")
	}
	if cfg.EnableHTTPS {
		err := http.ListenAndServeTLS(cfg.Host, certFile.Name(), privateKeyFile.Name(), handlers.NewURLRouter(repo, *cfg))
		if err != nil {
			logger.ZapSugar.Fatalw(err.Error(), "event", "start server")
		}
	}
	err = http.ListenAndServe(cfg.Host, handlers.NewURLRouter(repo, *cfg))
	if err != nil {
		logger.ZapSugar.Fatalw(err.Error(), "event", "start server")
	}
}
