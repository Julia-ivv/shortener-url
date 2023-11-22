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
			token, err := req.Cookie(authorizer.ACCESS_TOKEN)
			if err != nil {
				userID, tokenString, err := authorizer.BuildToken()
				if err != nil {
					http.Error(res, err.Error(), http.StatusInternalServerError)
					return
				}
				newctx = context.WithValue(req.Context(), authorizer.USER_CONTEXT_KEY, userID)
				http.SetCookie(res, &http.Cookie{
					Name:     authorizer.ACCESS_TOKEN,
					Value:    tokenString,
					Expires:  time.Now().Add(authorizer.TOKEN_EXP),
					Path:     "/",
					HttpOnly: true,
				})
			} else {
				userID, err := authorizer.GetUserIDFromToken(token.Value)
				if err != nil {
					http.Error(res, "401 Unauthorized", http.StatusUnauthorized)
					return
				}
				if userID == -1 {
					userID, tokenString, err := authorizer.BuildToken()
					if err != nil {
						http.Error(res, err.Error(), http.StatusInternalServerError)
						return
					}
					newctx = context.WithValue(req.Context(), authorizer.USER_CONTEXT_KEY, userID)
					http.SetCookie(res, &http.Cookie{
						Name:    authorizer.ACCESS_TOKEN,
						Value:   tokenString,
						Expires: time.Now().Add(authorizer.TOKEN_EXP),
					})
				}
				newctx = context.WithValue(req.Context(), authorizer.USER_CONTEXT_KEY, userID)
			}

			h.ServeHTTP(res, req.WithContext(newctx))
		})
}
