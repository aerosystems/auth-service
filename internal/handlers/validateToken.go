package handlers

import (
	"errors"
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
		err := errors.New("could not get token claims from Access Token")
		_ = WriteResponse(w, http.StatusUnauthorized, NewErrorPayload(401001, "could not get token claims from Access Token", err))
		return
	}

	payload := NewResponsePayload("token is active & valid", accessTokenClaims)
	_ = WriteResponse(w, http.StatusAccepted, payload)
	return
}
