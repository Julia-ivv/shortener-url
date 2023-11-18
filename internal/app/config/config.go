package config

import (
	"flag"

	"github.com/caarlos0/env"
)

type Flags struct {
	Host     string `env:"SERVER_ADDRESS"`    // -a адрес запуска HTTP-сервера, например localhost:8080
	URL      string `env:"BASE_URL"`          // -b базовый адрес результирующего сокращённого URL, например  http://localhost:8080
	FileName string `env:"FILE_STORAGE_PATH"` // -f полное имя файла, куда сохраняются данные в формате JSON
	DBDSN    string `env:"DATABASE_DSN"`      // -d строка с адресом подключения к БД
}

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
