package authorizer

import (
	"crypto/rand"
	"math/big"
)

func GenerateRandomInt(max int) (int, error) {
	rand, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(rand.Int64()), nil
}
