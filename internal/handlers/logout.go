package handlers

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/helpers"
	"net/http"

	"github.com/aerosystems/auth-service/internal/models"
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
func (h *BaseHandler) Logout(w http.ResponseWriter, r *http.Request) error {
	// receive AccessToken Claims from context middleware
	accessTokenClaims, ok := r.Context().Value(helpers.ContextKey("accessTokenClaims")).(*models.AccessTokenClaims)
	if !ok {
		err := errors.New("token is untracked")
		return WriteResponse(w, http.StatusUnauthorized, NewErrorPayload(err))
	}

	err := h.tokensRepo.DropCacheTokens(*accessTokenClaims)
	if err != nil {
		return WriteResponse(w, http.StatusUnauthorized, NewErrorPayload(err))
	}

	payload := Response{
		Error:   false,
		Message: fmt.Sprintf("User %s successfully logged out", accessTokenClaims.AccessUUID),
		Data:    accessTokenClaims,
	}
	return WriteResponse(w, http.StatusAccepted, payload)
}
