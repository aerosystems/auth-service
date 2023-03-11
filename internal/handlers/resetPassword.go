package handlers

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/helpers"
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
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /reset-password [post]
func (h *BaseHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var requestPayload ResetPasswordRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	addr, err := helpers.ValidateEmail(requestPayload.Email)
	if err != nil {
		err = errors.New("email is not valid")
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
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
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	user, _ := h.userRepo.FindByEmail(email)
	if user == nil {
		err = errors.New("invalid credentials")
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	// updating password for inactive user
	err = h.userRepo.ResetPassword(user, requestPayload.Password)
	if err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	code, _ := h.codeRepo.GetLastIsActiveCode(user.ID, "registration")

	if code == nil {
		// generating confirmation code
		_, err = h.codeRepo.NewCode(user.ID, "registration", "")
		if err != nil {
			_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
			return
		}
	} else {
		// extend expiration code and return previous active code
		_ = h.codeRepo.ExtendExpiration(code)
		_ = code.Code
		// TODO Send confirmation code
	}

	payload := Response{
		Error:   false,
		Message: fmt.Sprintf("Updated user with Id: %d", user.ID),
		Data:    nil,
	}

	_ = WriteResponse(w, http.StatusAccepted, payload)
	return
}
