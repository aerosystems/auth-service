package handlers

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/pkg/validators"
	"net/http"
	"time"
)

type CodeRequestBody struct {
	Code int `json:"code" example:"123456"`
}

type RPCProjectPayload struct {
	UserID     int
	UserRole   string
	Name       string
	AccessTime time.Time
}

// ConfirmRegistration godoc
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
// @Router /v1/user/confirm-registration [post]
func (h *BaseHandler) ConfirmRegistration(w http.ResponseWriter, r *http.Request) {
	var requestPayload CodeRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422001, "could not read request body", err))
		return
	}

	if err := validators.ValidateCode(requestPayload.Code); err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422004, "Code does not valid", err))
		return
	}

	code, err := h.codeRepo.GetByCode(requestPayload.Code)
	if err != nil {
		err = errors.New("code is not found in storage")
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(404001, "Code does not exist", err))
		return
	}
	if code.ExpireAt.Before(time.Now()) {
		err := fmt.Errorf("code %d has expired at %s", code.Code, code.ExpireAt.String())
		_ = WriteResponse(w, http.StatusGone, NewErrorPayload(410002, "Code has expired", err))
		return
	}
	if code.IsUsed {
		err := fmt.Errorf("code was used by user %d", code.User.ID)
		_ = WriteResponse(w, http.StatusGone, NewErrorPayload(410003, "Code was used", err))
		return
	}

	var payload *Response

	switch code.Action {
	case "registration":
		code.User.IsActive = true
		payload = NewResponsePayload(
			"successfully confirmed registration User",
			nil,
		)
		code.IsUsed = true
		err = h.codeRepo.UpdateWithAssociations(code)
		if err != nil {
			_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500002, "could not confirm registration", err))
			return
		}

		// create default project via RPC
		var result string
		err = h.projectClientRPC.Call("ProjectServer.CreateProject", RPCProjectPayload{
			UserID:     code.User.ID,
			UserRole:   code.User.Role,
			Name:       "default",
			AccessTime: time.Now(),
		}, &result)
		if err != nil {
			_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500003, "could not create default project", err))
			return
		}

		_ = WriteResponse(w, http.StatusOK, payload)
	case "reset":
		if !code.User.IsActive {
			code.User.IsActive = true
		}
		code.User.Password = code.Data

		payload = NewResponsePayload(
			"successfully confirmed changing password User",
			nil,
		)

		code.IsUsed = true
		err = h.codeRepo.UpdateWithAssociations(code)
		if err != nil {
			_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500002, "could not confirm registration", err))
			return
		}

		_ = WriteResponse(w, http.StatusOK, payload)
	}

	return
}
