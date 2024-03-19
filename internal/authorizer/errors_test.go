package authorizer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTokenError(t *testing.T) {
	t.Run("test create error struct", func(t *testing.T) {
		err := NewTokenError(NotValidToken, errors.New("error"))
		assert.NotEmpty(t, err)
	})
}
