package handlers

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/helpers"
	"github.com/aerosystems/auth-service/internal/models"
	"net/http"
)

// ValidateToken godoc
// @Summary validate token
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Success 200 {object} Response
// @Failure 401 {object} Response
// @Router /token/validate [get]
func (h *BaseHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	// receive AccessToken Claims from context middleware
	accessTokenClaims, ok := r.Context().Value(helpers.ContextKey("accessTokenClaims")).(*models.AccessTokenClaims)
	if !ok {
		err := errors.New("token is untracked")
		_ = WriteResponse(w, http.StatusUnauthorized, NewErrorPayload(err))
		return
	}

	payload := Response{
		Error:   false,
		Message: fmt.Sprintf("User %s having active token ", accessTokenClaims.AccessUUID),
	}
	_ = WriteResponse(w, http.StatusAccepted, payload)
	return
}
