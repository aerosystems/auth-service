package handlers

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/pkg/normalizers"
	"github.com/aerosystems/auth-service/pkg/validators"
	"gorm.io/gorm"
	"net/http"
	"net/rpc"
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
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/user/reset-password [post]
func (h *BaseHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var requestPayload ResetPasswordRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422001, "could not read request body", err))
		return
	}

	addr, err := validators.ValidateEmail(requestPayload.Email)
	if err != nil {
		err = errors.New("email is not valid")
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422005, "Email does not valid", err))
		return
	}

	email := normalizers.NormalizeEmail(addr)

	err = validators.ValidatePassword(requestPayload.Password)
	if err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422006, "Password does not valid", err))
		return
	}

	user, err := h.userRepo.FindByEmail(email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(404007, "could not find User", err))
		return
	}
	if user == nil {
		err := fmt.Errorf("user with claim Email %s does not exist", email)
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(404014, "user does not exist", err))
		return
	}

	// updating password for inactive user
	err = h.userRepo.ResetPassword(user, requestPayload.Password)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500005, "could not reset User password", err))
		return
	}

	code, err := h.codeRepo.GetLastIsActiveCode(user.ID, "registration")
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500013, "could not find Code", err))
		return
	}

	if code == nil {
		// generating confirmation code
		_, err = h.codeRepo.NewCode(*user, "registration", "")
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

	// sending confirmation code via RPC
	mailClientRPC, err := rpc.Dial("tcp", "mail-service:5001")
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500007, "could not send email", err))
		return
	}
	var result string
	err = mailClientRPC.Call("MailServer.SendEmail", RPCMailPayload{
		To:      user.Email,
		Subject: "Reset your password🗯",
		Body:    fmt.Sprintf("Your confirmation code is %d", code.Code),
	}, &result)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500008, "could not send email", err))
		return
	}

	_ = WriteResponse(w, http.StatusOK, NewResponsePayload(fmt.Sprintf("password was successfully reset for User with Email: %s", requestPayload.Email), nil))
	return
}
