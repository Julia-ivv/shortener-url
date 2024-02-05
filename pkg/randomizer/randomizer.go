// Package randomizer contains functions for random generation.
package randomizer

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
)

// LengthShortURL limits the length of a short URL.
const LengthShortURL = 4

// GenerateRandomBytes generates a slice of random characters.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString generates generates a string of random characters.
func GenerateRandomString(length int) (string, error) {
	b, err := GenerateRandomBytes(length)
	return base64.RawURLEncoding.EncodeToString(b), err
}

// GenerateRandomInt generates a random user ID.
func GenerateRandomInt(max int) (int, error) {
	rand, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(rand.Int64()), nil
}
