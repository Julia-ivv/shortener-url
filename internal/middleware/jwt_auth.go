// Package middleware contains middlewares for handlers.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Julia-ivv/shortener-url.git/internal/authorizer"
)

// HandlerWithAuth adds user authentication to the handler.
func HandlerWithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			var newctx context.Context
			var userID int
			var tokenString string
			token, err := req.Cookie(authorizer.AccessToken)
			if err != nil {
				userID, tokenString, err = authorizer.BuildToken()
				if err != nil {
					http.Error(res, err.Error(), http.StatusInternalServerError)
					return
				}
				newctx = context.WithValue(req.Context(), authorizer.UserContextKey, userID)
				http.SetCookie(res, &http.Cookie{
					Name:     authorizer.AccessToken,
					Value:    tokenString,
					Expires:  time.Now().Add(authorizer.TokenExp),
					Path:     "/",
					HttpOnly: true,
				})
			} else {
				userID, err = authorizer.GetUserIDFromToken(token.Value)
				if err != nil {
					var tokenErr *authorizer.TokenErr
					isTokenError := errors.As(err, &tokenErr)
					if isTokenError && (tokenErr.ErrType == authorizer.ParseError) {
						http.Error(res, "401 Unauthorized", http.StatusUnauthorized)
						return
					}
					if isTokenError && (tokenErr.ErrType == authorizer.NotValidToken) {
						userID, tokenString, err = authorizer.BuildToken()
						if err != nil {
							http.Error(res, err.Error(), http.StatusInternalServerError)
							return
						}
						newctx = context.WithValue(req.Context(), authorizer.UserContextKey, userID)
						http.SetCookie(res, &http.Cookie{
							Name:     authorizer.AccessToken,
							Value:    tokenString,
							Expires:  time.Now().Add(authorizer.TokenExp),
							Path:     "/",
							HttpOnly: true,
						})
					}
				} else {
					newctx = context.WithValue(req.Context(), authorizer.UserContextKey, userID)
				}
			}

			h.ServeHTTP(res, req.WithContext(newctx))
		})
}
