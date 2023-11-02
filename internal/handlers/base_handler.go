package handlers

import (
	"github.com/aerosystems/auth-service/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"strings"
)

type BaseHandler struct {
	mode         string
	log          *logrus.Logger
	tokenService services.TokenService
	userService  services.UserService
	codeService  services.CodeService
}

func NewBaseHandler(mode string, log *logrus.Logger, tokenService services.TokenService, userService services.UserService, codeService services.CodeService) *BaseHandler {
	return &BaseHandler{
		mode:         mode,
		log:          log,
		tokenService: tokenService,
		userService:  userService,
		codeService:  codeService,
	}
}

type TokensResponseBody struct {
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

type RefreshTokenRequestBody struct {
	RefreshToken string `json:"refreshToken" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

type CodeRequestBody struct {
	Code string `json:"code" validate:"required,min=6,max=6" example:"012345"`
}

type UserRequestBody struct {
	Email    string `json:"email" validate:"required,email" example:"example@gmail.com"`
	Password string `json:"password" validate:"required,customPasswordValidator" example:"P@ssw0rd"`
}

// Response is the type used for sending JSON around
type Response struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ErrorResponse is the type used for sending JSON around
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   any    `json:"error,omitempty"`
}

// SuccessResponse takes a response status code and arbitrary data and writes a json response to the client
func (h *BaseHandler) SuccessResponse(c echo.Context, statusCode int, message string, data any) error {
	payload := Response{
		Message: message,
		Data:    data,
	}
	return c.JSON(statusCode, payload)
}

// ErrorResponse takes a response status code and arbitrary data and writes a json response to the client. It depends on the mode whether the error is included in the response.
func (h *BaseHandler) ErrorResponse(c echo.Context, statusCode int, message string, err error) error {
	payload := Response{Message: message}
	if strings.ToLower(h.mode) == "dev" {
		payload.Data = err.Error()
	}
	return c.JSON(statusCode, payload)
}
