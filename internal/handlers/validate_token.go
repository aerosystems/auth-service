package handlers

import (
	"errors"
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
// @Failure 401 {object} handlers.ErrorResponse
// @Router /v1/token/validate [get]
func (h *BaseHandler) ValidateToken(c echo.Context) error {
	// receive AccessToken Claims from context middleware
	accessTokenClaims := c.Get("accessTokenClaims").(services.AccessTokenClaims)
	if len(accessTokenClaims.UserUuid) == 0 {
		return h.ErrorResponse(c, http.StatusUnauthorized, "invalid token", errors.New("invalid token"))
	}
	return h.SuccessResponse(c, http.StatusNoContent, "", nil)
}
