package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Julia-ivv/shortener-url.git/internal/app/config"
)

var cfg config.Flags

func Init() {
	cfg = *config.NewConfig()
}

func TestNewURLs(t *testing.T) {
	t.Run("create map repo", func(t *testing.T) {
		repo, err := NewURLs(cfg)
		assert.NoError(t, err)
		assert.NotEmpty(t, repo)
	})
	t.Run("create file repo", func(t *testing.T) {
		cfg.FileName = "for_tests.json"
		repo, err := NewURLs(cfg)
		assert.NoError(t, err)
		assert.NotEmpty(t, repo)
	})
}
