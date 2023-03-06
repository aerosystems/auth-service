package handlers

import (
	"encoding/json"
	"errors"
	"github.com/aerosystems/auth-service/internal/models"
	"golang.org/x/oauth2"
	"io"
	"net/http"
)

type BaseHandler struct {
	googleOauthConfig *oauth2.Config
	userRepo          models.UserRepository
	codeRepo          models.CodeRepository
	tokensRepo        models.TokensRepository
}

// Response is the type used for sending JSON around
type Response struct {
	Error   bool   `json:"error" xml:"error"`
	Message string `json:"message" xml:"message"`
	Data    any    `json:"data,omitempty" xml:"data,omitempty"`
}

func NewBaseHandler(googleOauthConfig *oauth2.Config,
	userRepo models.UserRepository,
	codeRepo models.CodeRepository,
	tokensRepo models.TokensRepository,
) *BaseHandler {
	return &BaseHandler{
		googleOauthConfig: googleOauthConfig,
		userRepo:          userRepo,
		codeRepo:          codeRepo,
		tokensRepo:        tokensRepo,
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

func NewErrorPayload(err error) Response {
	return Response{
		Error:   true,
		Message: err.Error(),
	}
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
