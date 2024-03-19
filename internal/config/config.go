// Package config receives settings when the application starts.
package config

import (
	"bufio"
	"encoding/json"
	"flag"
	"os"

	"github.com/caarlos0/env"

	"github.com/Julia-ivv/shortener-url.git/pkg/logger"
)

// Flags stores application launch settings.
type Flags struct {
	// Host (flag -a) - HTTP server launch address,  e.g. localhost:8080.
	Host string `env:"SERVER_ADDRESS" json:"server_address"`
	// URL (flag -b) - the base address of the resulting shortened URL, e.g.  http://localhost:8080.
	URL string `env:"BASE_URL" json:"base_url"`
	// FileName (flag -f) - full name of the JSON file to save data.
	FileName string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	// DBDSN (flag -d) - database connection address.
	DBDSN string `env:"DATABASE_DSN" json:"database_dsn"`
	// ConfigFileName (flag -c/-config) - the name of configuration file
	ConfigFileName string `env:"CONFIG"`
	// EnableHTTPS (flag -s) - if true, https enabled.
	EnableHTTPS bool `env:"ENABLE_HTTPS" json:"enable_https"`
	// TrustedSubnet (flag -t) - CIDR.
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	// GRPC (flag -g) - port for gRPC, e.g. :3200.
	GRPC string `env:"GRPC_PORT" json:"grpc"`
}

// Default values for flags.
const (
	defHost     string = ":8080"
	defURL      string = "http://localhost:8080"
	defFileName string = "/tmp/short-url-db.json"
	defHTTPS    bool   = false
	defGRPC     string = ":3200"
)

// readFromConf reads flag values from configuration file.
func readFromConf(c *Flags) error {
	f, err := os.Open(c.ConfigFileName)
	if err != nil {
		return err
	}
	defer f.Close()

	scan := bufio.NewScanner(f)
	allData := []byte{}
	for scan.Scan() {
		allData = append(allData, scan.Bytes()...)
	}
	if err = scan.Err(); err != nil {
		return err
	}

	var conf Flags
	err = json.Unmarshal(allData, &conf)
	if err != nil {
		return err
	}

	if c.Host == "" {
		c.Host = conf.Host
	}
	if c.URL == "" {
		c.URL = conf.URL
	}
	if c.FileName == "" {
		c.FileName = conf.FileName
	}
	if c.DBDSN == "" {
		c.DBDSN = conf.DBDSN
	}
	if c.TrustedSubnet == "" {
		c.TrustedSubnet = conf.TrustedSubnet
	}
	if c.GRPC == "" {
		c.GRPC = conf.GRPC
	}

	return nil
}

// NewConfig creates an instance with settings from flags or environment variables.
func NewConfig() *Flags {
	c := &Flags{}

	flag.StringVar(&c.Host, "a", defHost, "HTTP server start address")
	flag.StringVar(&c.URL, "b", defURL, "base address of the resulting URL")
	flag.StringVar(&c.FileName, "f", defFileName, "full filename to save URLs")
	flag.StringVar(&c.DBDSN, "d", "", "database connection address")
	flag.StringVar(&c.ConfigFileName, "c", "", "the name of configuration file")
	flag.StringVar(&c.ConfigFileName, "config", "", "the name of configuration file")
	flag.BoolVar(&c.EnableHTTPS, "s", defHTTPS, "https enabled")
	flag.StringVar(&c.TrustedSubnet, "t", "", "CIDR string")
	flag.StringVar(&c.GRPC, "g", defGRPC, "gRPC port")
	flag.Parse()

	env.Parse(c)

	if c.ConfigFileName != "" {
		err := readFromConf(c)
		if err != nil {
			logger.ZapSugar.Infow("reading configuration file", err)
		}
	}

	return c
}
