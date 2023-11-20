package middleware

import (
	"errors"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/aerosystems/auth-service/internal/services"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"strings"
)

func AuthTokenMiddleware(roles ...models.KindRole) echo.MiddlewareFunc {
	AuthorizationConfig := echojwt.Config{
		SigningKey:     []byte(os.Getenv("ACCESS_SECRET")),
		ParseTokenFunc: parseToken,
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		},
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := getTokenFromHeader(c)
			if err != nil {
				return AuthorizationConfig.ErrorHandler(c, err)
			}

			accessTokenClaims, err := services.DecodeAccessToken(token)
			if err != nil {
				return AuthorizationConfig.ErrorHandler(c, err)
			}

			if !isAccess(roles, accessTokenClaims.UserRole) {
				return echo.NewHTTPError(http.StatusForbidden, "access denied")
			}

			echo.Context(c).Set("accessTokenClaims", accessTokenClaims)

			return next(c)
		}
	}
}

func isAccess(roles []models.KindRole, role string) bool {
	for _, r := range roles {
		if r.String() == role {
			return true
		}
	}

	return false
}

func getTokenFromHeader(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if len(authHeader) == 0 {
		return "", errors.New("missing Authorization header")
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return "", errors.New("invalid token format")
	}

	return tokenParts[1], nil
}

func parseToken(c echo.Context, auth string) (interface{}, error) {
	_ = c
	accessTokenClaims, err := services.DecodeAccessToken(auth)
	if err != nil {
		return nil, err
	}

	return accessTokenClaims, nil
}
