package main

import (
	"github.com/aerosystems/auth-service/internal/handlers"
	"github.com/aerosystems/auth-service/internal/services"
)

type Config struct {
	BaseHandler  *handlers.BaseHandler
	TokenService services.TokenService
}
