package handlers

//
//import (
//	"errors"
//	"fmt"
//	"github.com/aerosystems/auth-service/internal/helpers"
//	"github.com/aerosystems/auth-service/internal/services"
//	"net/http"
//)
//
//// Logout godoc
//// @Summary logout user
//// @Tags auth
//// @Accept  json
//// @Produce application/json
//// @Security BearerAuth
//// @Success 200 {object} Response
//// @Failure 401 {object} ErrorResponse
//// @Failure 500 {object} ErrorResponse
//// @Router /v1/user/logout [post]
//func (h *BaseHandler) Logout(w http.ResponseWriter, r *http.Request) {
//	// receive AccessToken Claims from context middleware
//	accessTokenClaims, ok := r.Context().Value(helpers.ContextKey("accessTokenClaims")).(*services.AccessTokenClaims)
//	if !ok {
//		err := errors.New("could not get token claims from Access Token")
//		_ = WriteResponse(w, http.StatusUnauthorized, NewErrorPayload(401001, "could not get token claims from Access Token", err))
//		return
//	}
//
//	err := h.tokenService.DropCacheTokens(*accessTokenClaims)
//	if err != nil {
//		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500003, "could not drop Access Token", err))
//		return
//	}
//
//	_ = WriteResponse(w, http.StatusOK, NewResponsePayload(fmt.Sprintf("User %s successfully logged out", accessTokenClaims.AccessUUID), accessTokenClaims))
//	return
//}
