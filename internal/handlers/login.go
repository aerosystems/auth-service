package handlers

import (
	"fmt"
	"github.com/aerosystems/auth-service/pkg/normalizers"
	"github.com/aerosystems/auth-service/pkg/validators"
	"net/http"
)

type LoginRequestBody struct {
	Email    string `json:"email" example:"example@gmail.com"`
	Password string `json:"password" example:"P@ssw0rd"`
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
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/user/login [post]
func (h *BaseHandler) Login(w http.ResponseWriter, r *http.Request) {
	var requestPayload LoginRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422001, "could not read request body", err))
		return
	}

	addr, err := validators.ValidateEmail(requestPayload.Email)
	if err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422005, "Email does not valid", err))
		return
	}

	email := normalizers.NormalizeEmail(addr)

	if err := validators.ValidatePassword(requestPayload.Password); err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422006, "Password does not valid", err))
		return
	}

	user, err := h.userRepo.FindByEmail(email)
	if err != nil {
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(404007, "could not find User by Email", err))
		return
	}

	if !user.IsActive {
		err := fmt.Errorf("user %d did not confirm registration yet", user.ID)
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500002, "user did not confirm registration yet", err))
		return
	}

	valid, err := h.userRepo.PasswordMatches(user, requestPayload.Password)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(401009, "could not to check password", err))
		return
	}
	if !valid {
		if err == nil {
			err = fmt.Errorf("user %d entered invalid password", user.ID)
		}
		_ = WriteResponse(w, http.StatusUnauthorized, NewErrorPayload(401010, "invalid credentials", err))
		return
	}

	// create a pair of JWT tokens
	ts, err := h.tokenService.CreateToken(user.ID, user.Role)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500003, "could not to create a pair of JWT Tokens", err))
		return
	}

	// add a refresh token UUID to cache
	if err = h.tokenService.CreateCacheKey(user.ID, ts); err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500004, "could not to add a Refresh Token", err))
		return
	}

	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}

	_ = WriteResponse(w, http.StatusOK, NewResponsePayload(fmt.Sprintf("logged in User %s successfully", requestPayload.Email), tokens))
	return
}
