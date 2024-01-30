package authorizer

import (
	"crypto/rand"
	"math/big"
)

// GenerateRandomInt generates a random user ID.
func GenerateRandomInt(max int) (int, error) {
	rand, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(rand.Int64()), nil
}
