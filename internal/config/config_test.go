package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	flags := NewConfig()
	if assert.NotEmpty(t, flags) {
		assert.NotEmpty(t, flags.Host)
		assert.NotEmpty(t, flags.URL)
		assert.NotEmpty(t, flags.FileName)
	}
}

func TestReadFromConf(t *testing.T) {
	c := Flags{
		Host:           "",
		URL:            "",
		FileName:       "",
		DBDSN:          "",
		ConfigFileName: "for_tests.json",
		EnableHTTPS:    false,
		TrustedSubnet:  "",
	}
	err := readFromConf(&c)
	assert.NoError(t, err)
}
