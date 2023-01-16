package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Group(func(mux chi.Router) {
		// Public routes
		mux.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Content-Type"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
		mux.Use(middleware.Heartbeat("/ping"))

		mux.Post("/v1/login", app.Authenticate)
		mux.Post("/v1/register", app.Registration)
		mux.Post("/v1/confirm", app.Confirmation)
		mux.Post("/v1/reset", app.Reset)

		// Private routes
		mux.Group(func(mux chi.Router) {
			mux.Use(cors.Handler(cors.Options{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"POST", "OPTIONS"},
				AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
				ExposedHeaders:   []string{"Link"},
				AllowCredentials: true,
				MaxAge:           300,
			}))

			mux.Use(app.TokenAuthMiddleware)

			mux.Post("/v1/logout", app.Logout)
			mux.Post("/v1/refresh", app.Refresh)
		})
	})

	return mux
}
