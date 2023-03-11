package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/aerosystems/auth-service/internal/helpers"
	"github.com/aerosystems/auth-service/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type RegistrationRequestBody struct {
	Email    string `json:"email" example:"example@gmail.com"`
	Password string `json:"password" example:"P@ssw0rd"`
	Role     string `json:"role" example:"startup"`
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
// @Produce application/json l
// @Param registration body handlers.RegistrationRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /register [post]
func (h *BaseHandler) Register(w http.ResponseWriter, r *http.Request) {
	var requestPayload RegistrationRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	if err := helpers.ValidateRole(requestPayload.Role); err != nil {
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

	var payload Response

	//checking if email is existing
	user, _ := h.userRepo.FindByEmail(email)
	if user != nil {
		if user.IsActive {
			err = errors.New("email already exists")
			_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
			return
		} else {
			// updating password for inactive user
			err := h.userRepo.ResetPassword(user, requestPayload.Password)
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
	}

	// hashing password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestPayload.Password), 12)
	if err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
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
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	// generating confirmation code
	code, err := h.codeRepo.NewCode(newUser.ID, "registration", "")
	if err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
		return
	}

	payload = Response{
		Error:   false,
		Message: fmt.Sprintf("Registered user with Id: %d. Confirmation code: %d", newUser.ID, code.Code),
		Data:    nil,
	}

	_ = WriteResponse(w, http.StatusOK, payload)
	return
}
