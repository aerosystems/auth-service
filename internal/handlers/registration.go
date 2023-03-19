package handlers

import (
	"fmt"
	"gorm.io/gorm"
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
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /register [post]
func (h *BaseHandler) Register(w http.ResponseWriter, r *http.Request) {
	var requestPayload RegistrationRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400001, "request payload is incorrect", err))
		return
	}

	if err := helpers.ValidateRole(requestPayload.Role); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400010, "claim Role does not valid", err))
		return
	}

	addr, err := helpers.ValidateEmail(requestPayload.Email)
	if err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400005, "claim Email does not valid", err))
		return
	}

	email := helpers.NormalizeEmail(addr)

	// Minimum of one small case letter
	// Minimum of one upper case letter
	// Minimum of one digit
	// Minimum of one special character
	// Minimum 8 characters length
	if err := helpers.ValidatePassword(requestPayload.Password); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400006, "claim Password does not valid", err))
		return
	}

	var payload Response

	//checking if email is existing
	user, err := h.userRepo.FindByEmail(email)
	if err != nil && err != gorm.ErrRecordNotFound {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500007, "could not get User from storage by Email", err))
		return
	}
	if user != nil {
		if user.IsActive {
			err := fmt.Errorf("user with claim Email %s already exists", email)
			_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400011, "user with claim Email already exists", err))
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
				fmt.Sprintf("updated User with Email: %s", requestPayload.Email),
				nil,
			)

			_ = WriteResponse(w, http.StatusAccepted, payload)
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
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500010, "could not to create new User", err))
		return
	}

	// generating confirmation code
	code, err := h.codeRepo.NewCode(newUser.ID, "registration", "")
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500008, "could not gen new Code", err))
		return
	}

	// TODO Send confirmation code
	_ = code.Code

	payload = *NewResponsePayload(
		fmt.Sprintf("registered new User with Email %s", requestPayload.Email),
		nil,
	)

	_ = WriteResponse(w, http.StatusOK, payload)
	return
}
