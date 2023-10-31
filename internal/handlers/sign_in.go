package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// SignIn godoc
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
// @Param login body UserRequestBody true "raw request body"
// @Success 200 {object} Response{data=TokensResponseBody}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/user/login [post]
func (h *BaseHandler) SignIn(c echo.Context) error {
	var requestPayload UserRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return h.ErrorResponse(c, http.StatusUnprocessableEntity, "could not read request body", err)
	}
	user, err := h.userService.GetActiveUserByEmail(requestPayload.Email)
	if err != nil {
		return h.ErrorResponse(c, http.StatusNotFound, "user not found", err)
	}
	if !h.userService.CheckPassword(user, requestPayload.Password) {
		return h.ErrorResponse(c, http.StatusUnauthorized, "invalid password", err)
	}
	// create a pair of JWT tokens
	ts, err := h.tokenService.CreateToken(int(user.Id), user.Role)
	if err != nil {
		return h.ErrorResponse(c, http.StatusInternalServerError, "could not create a pair of JWT tokens", err)
	}
	// add a refresh token UUID to cache
	if err = h.tokenService.CreateCacheKey(int(user.Id), ts); err != nil {
		return h.ErrorResponse(c, http.StatusInternalServerError, "could not create a cache key", err)
	}
	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}
	return h.SuccessResponse(c, http.StatusOK, "user was successfully logged in", tokens)
}
