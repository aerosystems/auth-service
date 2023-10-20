package handlers

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/pkg/validators"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type CodeRequestBody struct {
	Code string `json:"code" example:"012345"`
}

// Confirm godoc
// @Summary confirm registration/reset password with 6-digit code from email/sms
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param code body handlers.CodeRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 410 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/user/confirm [post]
func (h *BaseHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	var requestPayload CodeRequestBody
	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422001, "could not read request body", err))
		return
	}
	if err := validators.ValidateCode(requestPayload.Code); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(422004, "code does not valid", err))
		return
	}
	code, err := h.codeRepo.GetByCode(requestPayload.Code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(404001, "code does not exist", err))
			return
		}
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500001, "could not get code", err))
		return
	}
	if code.ExpireAt.Before(time.Now()) {
		err := fmt.Errorf("code %s has expired at %s", code.Code, code.ExpireAt.String())
		_ = WriteResponse(w, http.StatusGone, NewErrorPayload(410002, "code has expired", err))
		return
	}
	if code.IsUsed {
		err := fmt.Errorf("code was used by user %d", code.User.ID)
		_ = WriteResponse(w, http.StatusGone, NewErrorPayload(410003, "code was used", err))
		return
	}
	if err := h.userService.Confirm(code); err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500001, "could not confirm code", err))
		return
	}

	_ = WriteResponse(w, http.StatusOK, NewResponsePayload("code was confirmed successfully", nil))
	return
}
