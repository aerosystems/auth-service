package handlers

import (
	"encoding/json"
	"errors"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/aerosystems/auth-service/internal/services"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strings"
)

type BaseHandler struct {
	log          *logrus.Logger
	codeRepo     models.CodeRepository
	tokenService services.TokenService
	userService  services.UserService
}

func NewBaseHandler(
	log *logrus.Logger,
	codeRepo models.CodeRepository,
	tokenService services.TokenService,
	userService services.UserService,
) *BaseHandler {
	return &BaseHandler{
		log:          log,
		codeRepo:     codeRepo,
		tokenService: tokenService,
		userService:  userService,
	}
}

// Response is the type used for sending JSON around
type Response struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type TokensResponseBody struct {
	AccessToken  string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// ErrorResponse is the type used for sending JSON around
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   any    `json:"error,omitempty"`
}

// ReadRequest tries to read the body of a request and converts it into JSON
func ReadRequest(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // one megabyte
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only a single json value")
	}

	return nil
}

// WriteResponse takes a response status code and arbitrary data and writes a json response to the client
func WriteResponse(w http.ResponseWriter, statusCode int, payload any, headers ...http.Header) error {
	out, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func NewResponsePayload(message string, data interface{}) *Response {
	return &Response{
		Message: message,
		Data:    data,
	}
}

func NewErrorPayload(code int, message string, err error) *ErrorResponse {
	switch strings.ToUpper(os.Getenv("APP_ENV")) {
	case "DEV":
		return &ErrorResponse{
			Code:    code,
			Message: message,
			Error:   err.Error(),
		}
	default:
		return &ErrorResponse{
			Code:    code,
			Message: message,
		}
	}
}
