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
