package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Julia-ivv/shortener-url.git/internal/app/authorizer"
)

func HandlerWithAuth(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			var newctx context.Context
			token, err := req.Cookie(authorizer.AccessToken)
			if err != nil {
				userID, tokenString, err := authorizer.BuildToken()
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
				userID, err := authorizer.GetUserIDFromToken(token.Value)
				if err != nil {
					http.Error(res, "401 Unauthorized", http.StatusUnauthorized)
					return
				}
				newctx = context.WithValue(req.Context(), authorizer.UserContextKey, userID)
				if userID == -1 {
					userID, tokenString, err := authorizer.BuildToken()
					if err != nil {
						http.Error(res, err.Error(), http.StatusInternalServerError)
						return
					}
					newctx = context.WithValue(req.Context(), authorizer.UserContextKey, userID)
					http.SetCookie(res, &http.Cookie{
						Name:    authorizer.AccessToken,
						Value:   tokenString,
						Expires: time.Now().Add(authorizer.TokenExp),
					})
				}
			}

			h.ServeHTTP(res, req.WithContext(newctx))
		})
}
