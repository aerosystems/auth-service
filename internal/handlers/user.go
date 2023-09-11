package handlers

import (
	"errors"
	"github.com/aerosystems/auth-service/internal/helpers"
	TokenService "github.com/aerosystems/auth-service/pkg/token_service"
	"gorm.io/gorm"
	"net/http"
)

func (h *BaseHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	accessTokenClaims := r.Context().Value(helpers.ContextKey("accessTokenClaims")).(*TokenService.AccessTokenClaims)
	user, err := h.userRepo.FindByID(accessTokenClaims.UserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(404001, "User not found", err))
		return
	}
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500001, "could not find User", err))
		return
	}
	_ = WriteResponse(w, http.StatusOK, NewResponsePayload("User successfully found", user))
	return

}
