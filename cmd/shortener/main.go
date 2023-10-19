package main

import (
	"net/http"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
	"github.com/Julia-ivv/shortener-url.git/internal/app/handlers"
	"github.com/Julia-ivv/shortener-url.git/internal/app/logger"
)

func main() {
	config.Flags = *config.NewConfig()

	logger.ZapSugar = logger.NewLogger()
	logger.ZapSugar.Infow("Starting server", "addr", config.Flags.Host)

	err := http.ListenAndServe(config.Flags.Host, handlers.NewURLRouter())
	if err != nil {
		logger.ZapSugar.Fatalw(err.Error(), "event", "start server")
	}
}
