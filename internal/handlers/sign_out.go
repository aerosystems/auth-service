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
// @Success 200 {object} handlers.Response
// @Failure 401 {object} handlers.Response
// @Failure 500 {object} handlers.Response
// @Router /v1/sign-out [post]
func (h *BaseHandler) SignOut(c echo.Context) error {
	accessTokenClaims := c.Get("accessTokenClaims").(*services.AccessTokenClaims)
	if err := h.tokenService.DropCacheTokens(*accessTokenClaims); err != nil {
		return h.ErrorResponse(c, http.StatusInternalServerError, "could not logout user", err)
	}
	return h.SuccessResponse(c, http.StatusOK, "user was successfully logged out", nil)
}
