// Package config receives settings when the application starts.
package config

import (
	"flag"

	"github.com/caarlos0/env"
)

// Flags stores application launch settings.
type Flags struct {
	// Host (flag -a) - HTTP server launch address,  e.g. localhost:8080.
	Host string `env:"SERVER_ADDRESS"`
	// URL (flag -b) - the base address of the resulting shortened URL, e.g.  http://localhost:8080.
	URL string `env:"BASE_URL"`
	// FileName (flag -f) - full name of the JSON file to save data.
	FileName string `env:"FILE_STORAGE_PATH"`
	// DBDSN (flag -d) - database connection address.
	DBDSN string `env:"DATABASE_DSN"`
}

// NewConfig creates an instance with settings from flags or environment variables.
func NewConfig() *Flags {
	c := &Flags{}

	flag.StringVar(&c.Host, "a", ":8080", "HTTP server start address")
	flag.StringVar(&c.URL, "b", "http://localhost:8080", "base address of the resulting URL")
	flag.StringVar(&c.FileName, "f", "/tmp/short-url-db.json", "full filename to save URLs")
	flag.StringVar(&c.DBDSN, "d", "", "database connection address")
	flag.Parse()

	env.Parse(c)

	return c
}
