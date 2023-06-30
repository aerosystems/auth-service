package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/handlers"
	"github.com/aerosystems/auth-service/internal/helpers"
	"github.com/golang-jwt/jwt"
	"net/http"
	"os"
	"strings"
)

func (app *Config) TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accessToken, err := app.GetAccessTokenFromHeader(r)
		if err != nil {
			_ = handlers.WriteResponse(w, http.StatusUnauthorized, handlers.NewErrorPayload(401002, "could not get Access Token from Header Authorization", err))
			return
		}

		token, err := app.VerifyToken(*accessToken)
		if err != nil {
			_ = handlers.WriteResponse(w, http.StatusUnauthorized, handlers.NewErrorPayload(401003, "could not verify Access Token", err))
			return
		}

		if !token.Valid {
			_ = handlers.WriteResponse(w, http.StatusUnauthorized, handlers.NewErrorPayload(401004, "access Token does not valid", err))
			return
		}

		tokenClaims, err := app.TokenService.DecodeAccessToken(*accessToken)
		if err != nil {
			_ = handlers.WriteResponse(w, http.StatusUnauthorized, handlers.NewErrorPayload(401005, "could not decode Access Token", err))
			return
		}

		_, err = app.TokenService.GetCacheValue(tokenClaims.AccessUUID)
		if err != nil {
			_ = handlers.WriteResponse(w, http.StatusUnauthorized, handlers.NewErrorPayload(401006, "could not get User data from storage", err))
			return
		}

		ctx := context.WithValue(r.Context(), helpers.ContextKey("accessTokenClaims"), tokenClaims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Config) GetAccessTokenFromHeader(r *http.Request) (*string, error) {
	headers := r.Header
	_, ok := headers["Authorization"]
	if !ok {
		return nil, errors.New("request must contain Authorization Header")
	}

	rawData := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	if len(rawData) != 2 {
		return nil, errors.New("authorization Header must contain Bearer format token")
	}
	accessToken := rawData[1]
	return &accessToken, nil
}

func (app *Config) VerifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}
