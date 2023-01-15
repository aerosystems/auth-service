package main

import (
	"context"
	"net/http"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

func (app *Config) TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accessToken, err := app.GetAccessTokenFromHeader(r)
		if err != nil {
			_ = app.errorJSON(w, err, http.StatusUnauthorized)
			return
		}

		token, err := app.VerifyToken(*accessToken)
		if err != nil {
			_ = app.errorJSON(w, err, http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			_ = app.errorJSON(w, err, http.StatusUnauthorized)
			return
		}

		tokenClaims, err := app.decodeAccessToken(*accessToken)
		if err != nil {
			_ = app.errorJSON(w, err, http.StatusUnauthorized)
			return
		}

		_, err = app.GetCacheValue(tokenClaims.AccessUUID)
		if err != nil {
			_ = app.errorJSON(w, err, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), contextKey("accessTokenClaims"), tokenClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
