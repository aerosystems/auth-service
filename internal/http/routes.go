package HttpServer

import (
	"github.com/aerosystems/auth-service/internal/models"
)

func (s *Server) setupRoutes() {
	s.echo.POST("/v1/sign-up", s.userHandler.SignUp)
	s.echo.POST("/v1/sign-in", s.userHandler.SignIn)
	s.echo.POST("/v1/confirm", s.userHandler.Confirm)
	s.echo.POST("/v1/reset-password", s.userHandler.ResetPassword)
	s.echo.POST("/v1/token/refresh", s.tokenHandler.RefreshToken)

	s.echo.GET("/v1/users", s.userHandler.GetUser, s.AuthTokenMiddleware(models.CustomerRole))
	s.echo.POST("/v1/sign-out", s.userHandler.SignOut, s.AuthTokenMiddleware(models.CustomerRole, models.StaffRole))
	s.echo.GET("/v1/token/validate", s.tokenHandler.ValidateToken, s.AuthTokenMiddleware(models.CustomerRole, models.StaffRole))
}
