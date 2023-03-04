package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/aerosystems/nix-trainee-5-6-7-8/internal/models"
	"github.com/labstack/echo/v4"
)

// Logout godoc
// @Summary logout user
// @Tags auth
// @Accept  json
// @Accept  xml
// @Produce application/json
// @Produce application/xml
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Success 202 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /users/logout [post]
func (h *BaseHandler) Logout(c echo.Context) error {
	// recieve AccessToken Claims from context middleware
	accessTokenClaims, ok := c.Get("user").(*models.AccessTokenClaims)
	if !ok {
		err := errors.New("token is untracked")
		return WriteResponse(c, http.StatusUnauthorized, NewErrorPayload(err))
	}

	err := h.tokensRepo.DropCacheTokens(*accessTokenClaims)
	if err != nil {
		return WriteResponse(c, http.StatusUnauthorized, NewErrorPayload(err))
	}

	payload := Response{
		Error:   false,
		Message: fmt.Sprintf("User %s successfully logged out", accessTokenClaims.AccessUUID),
		Data:    accessTokenClaims,
	}
	return WriteResponse(c, http.StatusAccepted, payload)
}
