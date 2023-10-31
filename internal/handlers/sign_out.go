package handlers

import (
	"github.com/aerosystems/auth-service/internal/services"
	"github.com/labstack/echo/v4"
	"net/http"
)

// SignOut godoc
// @Summary logout user
// @Tags auth
// @Accept  json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/user/logout [post]
func (h *BaseHandler) SignOut(c echo.Context) error {
	// receive AccessToken Claims from context middleware
	accessTokenClaims := c.Get("accessTokenClaims").(services.AccessTokenClaims)
	err := h.tokenService.DropCacheTokens(accessTokenClaims)
	if err != nil {
		return h.ErrorResponse(c, http.StatusInternalServerError, "could not logout user", err)
	}
	return h.ErrorResponse(c, http.StatusOK, "user was successfully logged out", nil)
}
