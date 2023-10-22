package handlers

import (
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
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(422005, "email does not valid", err))
		return
	}
	email := normalizers.NormalizeEmail(addr)
	if err := validators.ValidatePassword(requestPayload.Password); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(422006, "password does not valid", err))
		return
	}
	user, err := h.userService.MatchPassword(email, requestPayload.Password)
	if err != nil {
		_ = WriteResponse(w, http.StatusUnauthorized, NewErrorPayload(401001, "could not match email and password", err))
		return
	}
	// create a pair of JWT tokens
	ts, err := h.tokenService.CreateToken(int(user.UserId), user.Role)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500003, "could not to create a pair of JWT Tokens", err))
		return
	}
	// add a refresh token UUID to cache
	if err = h.tokenService.CreateCacheKey(int(user.UserId), ts); err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500004, "could not to add a Refresh Token", err))
		return
	}
	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}
	_ = WriteResponse(w, http.StatusOK, NewResponsePayload("user successfully logged in", tokens))
	return
}
