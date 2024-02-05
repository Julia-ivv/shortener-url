package authorizer

import (
	"math"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/Julia-ivv/shortener-url.git/pkg/randomizer"
)

// SecretKey - the secret key for the token.
const SecretKey = "byrhtvtyn"

// AccessToken - the name of the cookie for the token.
const AccessToken = "accessToken"

// TokenExp - token expiration time.
const TokenExp = time.Hour * 3

type key string

// UserContextKey - name of the key to get the token value from the context.
const UserContextKey key = "user"

// Claims for JWT token.
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

// BuildToken generates a new user ID and a token with this ID.
func BuildToken() (id int, tokenString string, err error) {
	id, err = randomizer.GenerateRandomInt(math.MaxInt32)
	if err != nil {
		return -1, "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
			},
			UserID: id,
		})
	tokenString, err = token.SignedString([]byte(SecretKey))
	if err != nil {
		return -1, "", err
	}
	return id, tokenString, nil
}

// GetUserIDFromToken gets the user ID from the JWT token.
func GetUserIDFromToken(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return -1, NewTokenError(ParseError, err)
	}

	if !token.Valid {
		return -1, NewTokenError(NotValidToken, nil)
	}

	return claims.UserID, nil
}
