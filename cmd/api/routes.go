package main

import (
	_ "github.com/aerosystems/auth-service/docs" // docs are generated by Swag CLI, you have to import it.
	"github.com/aerosystems/auth-service/internal/middleware"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func (app *Config) NewRouter() *echo.Echo {
	e := echo.New()

	docsGroup := e.Group("/docs")
	docsGroup.Use(middleware.BasicAuthMiddleware)
	docsGroup.GET("/*", echoSwagger.WrapHandler)

	e.POST("/v1/sign-up", app.baseHandler.SignUp)
	e.POST("/v1/sign-in", app.baseHandler.SignIn)
	e.POST("/v1/token/refresh", app.baseHandler.RefreshToken)
	e.POST("/v1/confirm", app.baseHandler.Confirm)
	e.POST("/v1/reset-password", app.baseHandler.ResetPassword)

	e.POST("/v1/sign-out", app.baseHandler.SignOut, middleware.AuthTokenMiddleware(models.Customer, models.Staff))
	e.GET("/v1/token/validate", app.baseHandler.ValidateToken, middleware.AuthTokenMiddleware(models.Customer, models.Staff))

	return e
}
