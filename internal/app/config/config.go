package config

import (
	"flag"
)

type Config struct {
	Host string // -a адрес запуска HTTP-сервера например localhost:8888
	URL  string // -b базовый адрес результирующего сокращённого URL, например  http://localhost:8000
}

type ConfigBuilder struct {
	config Config
}

var Flags Config

func (cb ConfigBuilder) SetHost(host string) ConfigBuilder {
	cb.config.Host = host
	return cb
}

func (cb ConfigBuilder) SetURL(url string) ConfigBuilder {
	cb.config.URL = url
	return cb
}

func InitConfigFromFlags() Config {
	var host string
	var url string
	var cb ConfigBuilder

	flag.StringVar(&host, "a", ":8080", "HTTP server start address")
	flag.StringVar(&url, "b", "http://localhost:8000", "base address of the resulting URL")
	flag.Parse()
	cb = cb.SetHost(host).SetURL(url)
	return cb.config
}
