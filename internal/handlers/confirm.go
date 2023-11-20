package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// Confirm godoc
// @Summary confirm registration/reset password with 6-digit code from email/sms
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param code body handlers.CodeRequestBody true "raw request body"
// @Success 200 {object} handlers.Response
// @Failure 400 {object} handlers.Response
// @Failure 422 {object} handlers.Response
// @Failure 500 {object} handlers.Response
// @Router /v1/confirm [post]
func (h *BaseHandler) Confirm(c echo.Context) error {
	var requestPayload CodeRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return h.ErrorResponse(c, http.StatusUnprocessableEntity, "could not read request body", err)
	}
	if err := c.Validate(requestPayload); err != nil {
		return err
	}
	code, err := h.codeService.GetCode(requestPayload.Code)
	if err != nil {
		return h.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
	}
	if err := h.userService.Confirm(code); err != nil {
		return h.ErrorResponse(c, http.StatusInternalServerError, "could not confirm user", err)
	}
	return h.SuccessResponse(c, http.StatusOK, "code was successfully confirmed", nil)
}
