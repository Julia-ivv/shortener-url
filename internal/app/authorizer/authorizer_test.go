package authorizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildToken(t *testing.T) {
	t.Run("test for build token", func(t *testing.T) {
		id, tokenStr, err := BuildToken()
		assert.NotEmpty(t, id)
		assert.NotEmpty(t, tokenStr)
		assert.NoError(t, err)
	})
}

func TestGetUserIDFromToken(t *testing.T) {
	t.Run("test with random string", func(t *testing.T) {
		id, err := GetUserIDFromToken("some_string")
		assert.Equal(t, -1, id)
		assert.Error(t, err)
	})
}
