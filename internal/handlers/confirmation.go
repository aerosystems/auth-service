package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aerosystems/nix-trainee-5-6-7-8/internal/helpers"
	"github.com/labstack/echo/v4"
)

type CodeRequestBoby struct {
	Code int `json:"code" xml:"code" example:"123456"`
}

// Confirmation godoc
// @Summary confirm registration/reset password with 6-digit code from email/sms
// @Tags auth
// @Accept  json
// @Accept  xml
// @Produce application/json
// @Produce application/xml
// @Param code body handlers.CodeRequestBoby true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /users/confirmation [post]
func (h *BaseHandler) Confirmation(c echo.Context) error {
	var requestPayload CodeRequestBoby

	if err := c.Bind(&requestPayload); err != nil {
		return WriteResponse(c, http.StatusBadRequest, NewErrorPayload(err))
	}

	if err := helpers.ValidateCode(requestPayload.Code); err != nil {
		return WriteResponse(c, http.StatusBadRequest, NewErrorPayload(err))
	}

	code, err := h.codeRepo.GetByCode(requestPayload.Code)
	if err != nil {
		err = errors.New("code is not found")
		return WriteResponse(c, http.StatusNotFound, NewErrorPayload(err))
	}
	if code.ExpireAt.Before(time.Now()) {
		err := errors.New("code is expired")
		return WriteResponse(c, http.StatusNotFound, NewErrorPayload(err))
	}
	if code.IsUsed {
		err := errors.New("code was used")
		return WriteResponse(c, http.StatusNotFound, NewErrorPayload(err))
	}

	user, err := h.userRepo.FindByID(code.UserID)
	if err != nil {
		return WriteResponse(c, http.StatusNotFound, NewErrorPayload(err))
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
		return WriteResponse(c, http.StatusBadRequest, NewErrorPayload(err))
	}

	code.IsUsed = true
	err = h.codeRepo.Update(code)
	if err != nil {
		return WriteResponse(c, http.StatusBadRequest, NewErrorPayload(err))
	}

	return WriteResponse(c, http.StatusAccepted, payload)
}
