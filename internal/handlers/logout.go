package handlers

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/helpers"
	TokenService "github.com/aerosystems/auth-service/pkg/token_service"
	"net/http"
)

// Logout godoc
// @Summary logout user
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /logout [post]
func (h *BaseHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// receive AccessToken Claims from context middleware
	accessTokenClaims, ok := r.Context().Value(helpers.ContextKey("accessTokenClaims")).(*TokenService.AccessTokenClaims)
	if !ok {
		err := errors.New("could not get token claims from Access Token")
		_ = WriteResponse(w, http.StatusUnauthorized, NewErrorPayload(401001, "could not get token claims from Access Token", err))
		return
	}

	err := h.tokenService.DropCacheTokens(*accessTokenClaims)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500003, "could not drop Access Token from storage", err))
		return
	}

	payload := NewResponsePayload(
		fmt.Sprintf("User %s successfully logged out", accessTokenClaims.AccessUUID),
		accessTokenClaims,
	)
	_ = WriteResponse(w, http.StatusOK, payload)
	return
}
