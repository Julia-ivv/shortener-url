package config

import (
	"flag"

	"github.com/caarlos0/env"
)

type Config struct {
	Host string `env:"SERVER_ADDRESS"` // -a адрес запуска HTTP-сервера, например localhost:8080
	URL  string `env:"BASE_URL"`       // -b базовый адрес результирующего сокращённого URL, например  http://localhost:8080
}

var Flags Config

func NewConfig() *Config {
	c := &Config{}

	flag.StringVar(&c.Host, "a", ":8080", "HTTP server start address")
	flag.StringVar(&c.URL, "b", "http://localhost:8080", "base address of the resulting URL")
	flag.Parse()

	env.Parse(&c)

	return c
}
