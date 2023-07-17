package handlers

import (
	"fmt"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/aerosystems/auth-service/pkg/normalizers"
	"github.com/aerosystems/auth-service/pkg/validators"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
)

type RegistrationRequestBody struct {
	Email    string `json:"email" example:"example@gmail.com"`
	Password string `json:"password" example:"P@ssw0rd"`
	Role     string `json:"role" example:"startup"`
}

type RPCMailPayload struct {
	To      string
	Subject string
	Body    string
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
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422001, "could not read request body", err))
		return
	}

	if err := validators.ValidateRole(requestPayload.Role); err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422010, "Role does not valid", err))
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

	var payload Response

	user, err := h.userRepo.FindByEmail(email)
	if err != nil && err != gorm.ErrRecordNotFound {
		_ = WriteResponse(w, http.StatusNotFound, NewErrorPayload(404007, "could not find User", err))
		return
	}

	if user != nil {
		if user.IsActive {
			err := fmt.Errorf("user with claim Email %s already exists", email)
			_ = WriteResponse(w, http.StatusConflict, NewErrorPayload(409011, "user already exists", err))
			return
		} else {
			// updating password for inactive user
			err := h.userRepo.ResetPassword(user, requestPayload.Password)
			if err != nil {
				_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500005, "could not reset User password", err))
				return
			}

			code, _ := h.codeRepo.GetLastIsActiveCode(user.ID, "registration")

			if code == nil {
				// generating confirmation code
				_, err = h.codeRepo.NewCode(*user, "registration", "")
				if err != nil {
					_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500008, "could not create new Code", err))
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
				fmt.Sprintf("User with Email %s was updated successfully", requestPayload.Email),
				nil,
			)

			_ = WriteResponse(w, http.StatusOK, payload)
			return
		}
	}

	// hashing password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestPayload.Password), 12)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500009, "could not create Password", err))
		return
	}

	// creating new inactive user
	newUser := models.User{
		Email:    email,
		Password: string(hashedPassword),
		Role:     requestPayload.Role,
	}

	err = h.userRepo.Create(&newUser)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500010, "could not create new User", err))
		return
	}

	// generating confirmation code
	code, err := h.codeRepo.NewCode(newUser, "registration", "")
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500008, "could not gen new Code", err))
		return
	}

	// TODO Uncomment when mail service will be ready
	_ = code.Code

	// sending confirmation code via RPC
	//var result string
	//err = h.mailClientRPC.Call("MailServer.SendEmail", RPCMailPayload{
	//	To:      newUser.Email,
	//	Subject: "Confirm your email🗯",
	//	Body:    fmt.Sprintf("Your confirmation code is %d", code.Code),
	//}, &result)
	//if err != nil {
	//	_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500008, "could not send email", err))
	//	return
	//}

	payload = *NewResponsePayload(
		fmt.Sprintf("User with Email %s was registered successfully", requestPayload.Email),
		nil,
	)

	_ = WriteResponse(w, http.StatusOK, payload)
	return
}
