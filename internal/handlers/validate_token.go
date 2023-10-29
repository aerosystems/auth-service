package handlers

//
//import (
//	"errors"
//	"github.com/aerosystems/auth-service/internal/helpers"
//	"github.com/aerosystems/auth-service/internal/services"
//	"net/http"
//)
//
//// ValidateToken godoc
//// @Summary validate token
//// @Tags api-gateway-special
//// @Accept  json
//// @Produce application/json
//// @Security BearerAuth
//// @Success 204 {object} Response
//// @Failure 401 {object} ErrorResponse
//// @Router /v1/token/validate [get]
//func (h *BaseHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
//	// receive AccessToken Claims from context middleware
//	_, ok := r.Context().Value(helpers.ContextKey("accessTokenClaims")).(*services.AccessTokenClaims)
//	if !ok {
//		err := errors.New("could not get token claims from Access Token")
//		_ = WriteResponse(w, http.StatusUnauthorized, NewErrorPayload(401001, "could not get token claims from Access Token", err))
//		return
//	}
//	_ = WriteResponse(w, http.StatusNoContent, nil)
//	return
//}
