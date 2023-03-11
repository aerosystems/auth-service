package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aerosystems/auth-service/internal/helpers"
)

type CodeRequestBody struct {
	Code int `json:"code" xml:"code" example:"123456"`
}

// Confirm godoc
// @Summary confirm registration/reset password with 6-digit code from email/sms
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param code body handlers.CodeRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /confirm [post]
func (h *BaseHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	var requestPayload CodeRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	if err := helpers.ValidateCode(requestPayload.Code); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	code, err := h.codeRepo.GetByCode(requestPayload.Code)
	if err != nil {
		err = errors.New("code is not found")
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(err))
		return
	}
	if code.ExpireAt.Before(time.Now()) {
		err := errors.New("code is expired")
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(err))
		return
	}
	if code.IsUsed {
		err := errors.New("code was used")
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(err))
		return
	}

	user, err := h.userRepo.FindByID(code.UserID)
	if err != nil {
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(err))
		return
	}

	var payload Response

	switch code.Action {
	case "registration":
		user.IsActive = true
		payload = Response{
			Error:   false,
			Message: fmt.Sprintf("Succesfuly confirmed registration user with Id: %d", user.ID),
			Data:    user,
		}
	case "reset":
		if !user.IsActive {
			user.IsActive = true
		}
		user.Password = code.Data

		payload = Response{
			Error:   false,
			Message: fmt.Sprintf("Succesfuly confirmed changing user password with Id: %d", user.ID),
			Data:    user,
		}
	}

	err = h.userRepo.Update(user)
	if err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	code.IsUsed = true
	err = h.codeRepo.Update(code)
	if err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	_ = WriteResponse(w, http.StatusAccepted, payload)
	return
}
