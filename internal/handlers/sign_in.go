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
// @Description Response contain pair JWT tokens
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param login body handlers.UserRequestBody true "raw request body"
// @Success 200 {object} handlers.Response{data=handlers.TokensResponseBody}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 422 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /v1/sign-in [post]
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
	ts, err := h.tokenService.CreateToken(user.Uuid.String(), user.Role.String())
	if err != nil {
		return h.ErrorResponse(c, http.StatusInternalServerError, "could not create a pair of JWT tokens", err)
	}
	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}
	return h.SuccessResponse(c, http.StatusOK, "user was successfully logged in", tokens)
}
