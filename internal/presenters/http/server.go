package HttpServer

import (
	"fmt"
	"github.com/aerosystems/auth-service/internal/presenters/http/handlers"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

const webPort = 80

type Server struct {
	log          *logrus.Logger
	echo         *echo.Echo
	accessSecret string
	userHandler  *handlers.UserHandler
	tokenHandler *handlers.TokenHandler
}

func NewServer(
	log *logrus.Logger,
	accessSecret string,
	userHandler *handlers.UserHandler,
	tokenHandler *handlers.TokenHandler,
) *Server {
	return &Server{
		log:          log,
		echo:         echo.New(),
		accessSecret: accessSecret,
		userHandler:  userHandler,
		tokenHandler: tokenHandler,
	}
}

func (s *Server) Run() error {
	s.setupMiddleware()
	s.setupRoutes()
	s.setupValidator()
	s.log.Infof("starting HTTP server auth-service on port %d\n", webPort)
	return s.echo.Start(fmt.Sprintf(":%d", webPort))
}
