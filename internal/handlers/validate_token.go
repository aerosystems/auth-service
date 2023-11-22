package handlers

import (
	"github.com/aerosystems/auth-service/internal/services"
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
// @Failure 401 {object} handlers.Response
// @Router /v1/token/validate [get]
func (h *BaseHandler) ValidateToken(c echo.Context) error {
	accessTokenClaims := c.Get("accessTokenClaims").(*services.AccessTokenClaims)
	if _, err := h.tokenService.GetCacheValue(accessTokenClaims.AccessUuid); err != nil {
		return h.ErrorResponse(c, http.StatusUnauthorized, "invalid token", err)
	}
	return h.SuccessResponse(c, http.StatusNoContent, "", nil)
}
