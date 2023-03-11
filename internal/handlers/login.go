package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/aerosystems/auth-service/internal/helpers"
)

type LoginRequestBody struct {
	Email    string `json:"email" xml:"email" example:"example@gmail.com"`
	Password string `json:"password" xml:"password" example:"P@ssw0rd"`
}

type TokensResponseBody struct {
	AccessToken  string `json:"access_token" xml:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	RefreshToken string `json:"refresh_token" xml:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// Login godoc
// @Summary login user by credentials
// @Description Password should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Description Response contain pair JWT tokens, use /token/refresh for updating them
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param login body handlers.LoginRequestBody true "raw request body"
// @Success 200 {object} Response{data=TokensResponseBody}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /login [post]
func (h *BaseHandler) Login(w http.ResponseWriter, r *http.Request) {
	var requestPayload LoginRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	addr, err := helpers.ValidateEmail(requestPayload.Email)
	if err != nil {
		err = errors.New("email is not valid")
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	email := helpers.NormalizeEmail(addr)

	// Minimum of one small case letter
	// Minimum of one upper case letter
	// Minimum of one digit
	// Minimum of one special character
	// Minimum 8 characters length
	if err := helpers.ValidatePassword(requestPayload.Password); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	// validate against database
	user, err := h.userRepo.FindByEmail(email)
	if err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	if !user.IsActive {
		err := errors.New("user has did not confirm registration")
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	valid, err := h.userRepo.PasswordMatches(user, requestPayload.Password)
	if err != nil || !valid {
		err := errors.New("invalid credentials")
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	// create pair JWT tokens
	ts, err := h.tokensRepo.CreateToken(user.ID)
	if err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	// add refresh token UUID to cache
	if err = h.tokensRepo.CreateCacheKey(user.ID, ts); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}

	payload := Response{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", requestPayload.Email),
		Data:    tokens,
	}

	_ = WriteResponse(w, http.StatusAccepted, payload)
	return
}
