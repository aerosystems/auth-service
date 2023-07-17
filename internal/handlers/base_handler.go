package handlers

import (
	"encoding/json"
	"errors"
	"github.com/aerosystems/auth-service/internal/models"
	TokenService "github.com/aerosystems/auth-service/pkg/token_service"
	"io"
	"net/http"
	"net/rpc"
	"os"
	"strings"
)

type BaseHandler struct {
	userRepo         models.UserRepository
	codeRepo         models.CodeRepository
	tokenService     *TokenService.Service
	projectClientRPC *rpc.Client
	mailClientRPC    *rpc.Client
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

func NewBaseHandler(userRepo models.UserRepository,
	codeRepo models.CodeRepository,
	tokenService *TokenService.Service,
	projectClientRPC *rpc.Client,
	mailClientRPC *rpc.Client,
) *BaseHandler {
	return &BaseHandler{
		userRepo:         userRepo,
		codeRepo:         codeRepo,
		tokenService:     tokenService,
		projectClientRPC: projectClientRPC,
		mailClientRPC:    mailClientRPC,
	}
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
