package authorizer

import (
	"math"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const SecretKey = "byrhtvtyn"
const AccessToken = "accessToken"
const TokenExp = time.Hour * 3

type key string

const UserContextKey key = "user"

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func BuildToken() (id int, tokenString string, err error) {
	id, err = GenerateRandomInt(math.MaxInt32)
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

func GetUserIDFromToken(tokenString string) (int, error) {
	// возвращает -1 и ошибку, если не удалось получить id
	// -1 и nil, если токен невалидный
	// id и nil, если удалось получить id из токена
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return -1, err
	}

	if !token.Valid {
		return -1, nil
	}

	return claims.UserID, nil
}
