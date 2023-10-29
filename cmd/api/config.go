package main

import (
	"github.com/aerosystems/auth-service/internal/handlers"
	"github.com/aerosystems/auth-service/internal/services"
)

type Config struct {
	baseHandler  *handlers.BaseHandler
	tokenService services.TokenService
}

func NewConfig(baseHandler *handlers.BaseHandler, tokenService services.TokenService) *Config {
	return &Config{
		baseHandler:  baseHandler,
		tokenService: tokenService,
	}
}
