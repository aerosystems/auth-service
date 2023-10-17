package handlers

import (
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/aerosystems/auth-service/pkg/normalizers"
	"github.com/aerosystems/auth-service/pkg/validators"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"net/rpc"
)

type RegistrationRequestBody struct {
	Email    string `json:"email" example:"example@gmail.com"`
	Password string `json:"password" example:"P@ssw0rd"`
}

type MailRPCPayload struct {
	To      string
	Subject string
	Body    string
}

type InspectRPCPayload struct {
	Domain   string
	ClientIp string
}

// Register godoc
// @Summary registration user by credentials
// @Description Password should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param registration body handlers.RegistrationRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/user/register [post]
func (h *BaseHandler) Register(w http.ResponseWriter, r *http.Request) {
	var requestPayload RegistrationRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422001, "Could not read request body", err))
		return
	}

	addr, err := validators.ValidateEmail(requestPayload.Email)
	if err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422005, "Email does not valid", err))
		return
	}

	email := normalizers.NormalizeEmail(addr)

	if err := validators.ValidatePassword(requestPayload.Password); err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422006, "Password does not valid", err))
		return
	}

	// checking email in blacklist via RPC
	if checkmailClientRPC, err := rpc.Dial("tcp", "checkmail-service:5001"); err == nil {
		var result string
		if err := checkmailClientRPC.Call(
			"CheckmailServer.Inspect",
			InspectRPCPayload{
				Domain:   email,
				ClientIp: r.RemoteAddr,
			},
			&result); err != nil {
			_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400101, "Email address does not valid", err))
			return
		}

		if result == "blacklist" {
			err := fmt.Errorf("email address contains in blacklist")
			_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400102, "Email address contains in Blacklist", err))
			return
		}
	} else {
		h.log.Error(err)
	}

	user, err := h.userRepo.FindByEmail(email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(404007, "Could not find User", err))
		return
	}

	if user != nil {
		if user.IsActive {
			err := fmt.Errorf("user with claim Email %s already exists", email)
			_ = WriteResponse(w, http.StatusConflict, NewErrorPayload(409011, "User already exists", err))
			return
		} else {
			// updating password for inactive user
			err := h.userRepo.ResetPassword(user, requestPayload.Password)
			if err != nil {
				_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500005, "Could not reset User password", err))
				return
			}

			code, _ := h.codeRepo.GetLastIsActiveCode(user.ID, "registration")

			if code == nil {
				// generating confirmation code
				_, err = h.codeRepo.NewCode(*user, "registration", "")
				if err != nil {
					_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500008, "Could not create new Code", err))
					return
				}
			} else {
				// extend expiration code and return previous active code
				if err = h.codeRepo.ExtendExpiration(code); err != nil {
					_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500012, "Could not extend expiration date Code", err))
					return
				}
			}

			// sending confirmation code via RPC
			mailClientRPC, err := rpc.Dial("tcp", "mail-service:5001")
			if err != nil {
				_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500007, "Could not send email", err))
				return
			}
			var result string
			if err := mailClientRPC.Call("MailServer.SendEmail",
				MailRPCPayload{
					To:      user.Email,
					Subject: "Confirm your email🗯",
					Body:    fmt.Sprintf("Your confirmation code is %s", code.Code),
				}, &result); err != nil {
				_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500008, "Could not send email", err))
				return
			}

			_ = WriteResponse(w, http.StatusOK, NewResponsePayload(fmt.Sprintf("User with Email %s was updated successfully", requestPayload.Email), nil))
			return
		}
	}

	// hashing password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestPayload.Password), 12)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500009, "Could not create Password", err))
		return
	}

	// creating new inactive user
	newUser := models.User{
		Email:    email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	err = h.userRepo.Create(&newUser)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500010, "Could not create new User", err))
		return
	}

	// generating confirmation code
	code, err := h.codeRepo.NewCode(newUser, "registration", "")
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500008, "Could not gen new Code", err))
		return
	}

	// sending confirmation code via RPC
	mailClientRPC, err := rpc.Dial("tcp", "mail-service:5001")
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500007, "Could not send email", err))
		return
	}
	var result string
	err = mailClientRPC.Call("MailServer.SendEmail", MailRPCPayload{
		To:      newUser.Email,
		Subject: "Confirm your email🗯",
		Body:    fmt.Sprintf("Your confirmation code is %s", code.Code),
	}, &result)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500008, "Could not send email", err))
		return
	}

	_ = WriteResponse(w, http.StatusOK, NewResponsePayload(fmt.Sprintf("User with Email %s was registered successfully", requestPayload.Email), nil))
	return
}
