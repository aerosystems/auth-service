package handlers

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/helpers"
	"gorm.io/gorm"
	"net/http"
)

type ResetPasswordRequestBody struct {
	Email    string `json:"email" example:"example@gmail.com"`
	Password string `json:"password" example:"P@ssw0rd"`
}

// ResetPassword godoc
// @Summary resetting password
// @Description Password should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param registration body handlers.ResetPasswordRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reset-password [post]
func (h *BaseHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var requestPayload ResetPasswordRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400001, "request payload is incorrect", err))
		return
	}

	addr, err := helpers.ValidateEmail(requestPayload.Email)
	if err != nil {
		err = errors.New("email is not valid")
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400005, "claim Email does not valid", err))
		return
	}

	email := helpers.NormalizeEmail(addr)

	// Minimum of one small case letter
	// Minimum of one upper case letter
	// Minimum of one digit
	// Minimum of one special character
	// Minimum 8 characters length
	err = helpers.ValidatePassword(requestPayload.Password)
	if err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400006, "claim Password does not valid", err))
		return
	}

	user, err := h.userRepo.FindByEmail(email)
	if err != nil && err != gorm.ErrRecordNotFound {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500007, "could not get User from storage by Email", err))
		return
	}
	if user == nil {
		err := fmt.Errorf("user with claim Email %s does not exist", email)
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400014, "user with claim Email does not exist", err))
		return
	}

	// updating password for inactive user
	err = h.userRepo.ResetPassword(user, requestPayload.Password)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500005, "could not reset User password", err))
		return
	}

	code, err := h.codeRepo.GetLastIsActiveCode(user.ID, "registration")
	if err != nil && err != gorm.ErrRecordNotFound {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(5000013, "could not get Code from storage", err))
		return
	}

	if code == nil {
		// generating confirmation code
		_, err = h.codeRepo.NewCode(user.ID, "registration", "")
		if err != nil {
			_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500008, "could not gen new Code", err))
			return
		}
	} else {
		// extend expiration code and return previous active code
		if err = h.codeRepo.ExtendExpiration(code); err != nil {
			_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500012, "could not extend expiration date Code", err))
			return
		}
	}

	// TODO Send confirmation code
	_ = code.Code

	payload := NewResponsePayload(
		fmt.Sprintf("resetted new passwoed for User with Email: %s", requestPayload.Email),
		nil,
	)

	_ = WriteResponse(w, http.StatusAccepted, payload)
	return
}
