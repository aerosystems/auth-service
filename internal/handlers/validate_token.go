package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// ValidateToken godoc
// @Summary validate token
// @Tags api-gateway-special
// @Accept  json
// @Produce application/json
// @Security BearerAuth
// @Success 204 {object} handlers.Response
// @Failure 401 {object} handlers.ErrorResponse
// @Router /v1/token/validate [get]
func (h *BaseHandler) ValidateToken(c echo.Context) error {
	return h.SuccessResponse(c, http.StatusNoContent, "", nil)
}
