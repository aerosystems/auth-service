package main

import (
	"github.com/go-chi/cors"
	"net/http"

	_ "github.com/aerosystems/auth-service/docs" // docs are generated by Swag CLI, you have to import it.
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	mux.Group(func(mux chi.Router) {
		mux.Use(middleware.Heartbeat("/ping"))

		mux.Get("/docs/*", httpSwagger.Handler(
			httpSwagger.URL("doc.json"), // The url pointing to API definition
		))

		mux.Post("/v1/user/login", app.BaseHandler.Login)
		mux.Post("/v1/user/register", app.BaseHandler.Register)
		mux.Post("/v1/user/confirm-registration", app.BaseHandler.ConfirmRegistration)
		mux.Post("/v1/user/reset-password", app.BaseHandler.ResetPassword)
		mux.Post("/v1/token/refresh", app.BaseHandler.RefreshToken)

		// Private routes
		mux.Group(func(mux chi.Router) {
			mux.Use(app.TokenAuthMiddleware)

			mux.Post("/v1/user/logout", app.BaseHandler.Logout)
			mux.Get("/v1/token/validate", app.BaseHandler.ValidateToken)
		})
	})

	return mux
}
