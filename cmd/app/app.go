package main

import (
	"github.com/aerosystems/auth-service/internal/config"
	HTTPServer "github.com/aerosystems/auth-service/internal/http"
	"github.com/sirupsen/logrus"
)

type App struct {
	log        *logrus.Logger
	cfg        *config.Config
	httpServer *HTTPServer.Server
}

func NewApp(
	log *logrus.Logger,
	cfg *config.Config,
	httpServer *HTTPServer.Server,
) *App {
	return &App{
		log:        log,
		cfg:        cfg,
		httpServer: httpServer,
	}
}
