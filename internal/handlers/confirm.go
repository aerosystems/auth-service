package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aerosystems/auth-service/internal/helpers"
)

type CodeRequestBody struct {
	Code int `json:"code" example:"123456"`
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
// @Failure 500 {object} ErrorResponse
// @Router /confirm [post]
func (h *BaseHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	var requestPayload CodeRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400001, "request payload is incorrect", err))
		return
	}

	if err := helpers.ValidateCode(requestPayload.Code); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400004, "claim Code does not valid", err))
		return
	}

	code, err := h.codeRepo.GetByCode(requestPayload.Code)
	if err != nil {
		err = errors.New("claim Code is not found in storage")
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(404001, "claim Code does not exist", err))
		return
	}
	if code.ExpireAt.Before(time.Now()) {
		err := errors.New("code has expired")
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(404002, "claim Code has expired", err))
		return
	}
	if code.IsUsed {
		err := fmt.Errorf("claim Code was used by user %d", code.UserID)
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(404003, "claim Code was used", err))
		return
	}

	user, err := h.userRepo.FindByID(code.UserID)
	if err != nil {
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(404004, "could not find User by Code", err))
		return
	}

	var payload *Response

	switch code.Action {
	case "registration":
		user.IsActive = true
		payload = NewResponsePayload(
			"successfully confirmed registration User",
			user,
		)
	case "reset":
		if !user.IsActive {
			user.IsActive = true
		}
		user.Password = code.Data

		payload = NewResponsePayload(
			"successfully confirmed changing password User",
			user,
		)
	}

	err = h.userRepo.Update(user)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500001, "could not update User data", err))
		return
	}

	code.IsUsed = true
	err = h.codeRepo.Update(code)
	if err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(500002, "could not update Code data", err))
		return
	}

	_ = WriteResponse(w, http.StatusOK, payload)
	return
}
