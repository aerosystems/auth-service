package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// ResetPassword godoc
// @Summary resetting password
// @Description Password should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param registration body handlers.UserRequestBody true "raw request body"
// @Success 200 {object} handlers.Response
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 422 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /v1/reset-password [post]
func (h *BaseHandler) ResetPassword(c echo.Context) error {
	var requestPayload UserRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return h.ErrorResponse(c, http.StatusUnprocessableEntity, "could not read request body", err)
	}
	if err := h.userService.ResetPassword(requestPayload.Email, requestPayload.Password); err != nil {
		return h.ErrorResponse(c, http.StatusInternalServerError, "could not reset password", err)
	}
	return h.SuccessResponse(c, http.StatusOK, "password was successfully reset", nil)
}
