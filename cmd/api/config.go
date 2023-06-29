package main

import (
	"github.com/aerosystems/auth-service/internal/handlers"
	TokenService "github.com/aerosystems/auth-service/pkg/token_service"
)

type Config struct {
	WebPort      string
	BaseHandler  *handlers.BaseHandler
	TokenService *TokenService.Service
}
