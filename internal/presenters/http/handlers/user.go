package handlers

import (
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

type UserHandler struct {
	*BaseHandler
	tokenUsecase TokenUsecase
	authUsecase  AuthUsecase
}

func NewUserHandler(baseHandler *BaseHandler, tokenUsecase TokenUsecase, userUsecase AuthUsecase) *UserHandler {
	return &UserHandler{
		BaseHandler:  baseHandler,
		tokenUsecase: tokenUsecase,
		authUsecase:  userUsecase,
	}
}

type CodeRequestBody struct {
	Code string `json:"code" validate:"required,numeric,len=6" example:"012345"`
}

type UserRequestBody struct {
	Email    string `json:"email" validate:"required,email" example:"example@gmail.com"`
	Password string `json:"password" validate:"required,customPasswordRule" example:"P@ssw0rd"`
}

type UserResponseBody struct {
	Uuid  string `json:"uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email string `json:"email" example:"example@gmail.com"`
	Role  string `json:"role" example:"customer"`
}

func ModelToResponse(user *models.User) *UserResponseBody {
	return &UserResponseBody{
		Uuid:  user.Uuid.String(),
		Email: user.Email,
		Role:  user.Role.String(),
	}
}

// SignUp godoc
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
// @Param registration body UserRequestBody true "raw request body"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 422 {object} Response
// @Failure 500 {object} Response
// @Router /v1/sign-up [post]
func (uh UserHandler) SignUp(c echo.Context) error {
	var requestPayload UserRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return uh.ErrorResponse(c, http.StatusUnprocessableEntity, "could not read request body", err)
	}
	if err := c.Validate(requestPayload); err != nil {
		return err
	}
	if err := uh.authUsecase.RegisterCustomer(requestPayload.Email, requestPayload.Password, c.RealIP()); err != nil {
		return uh.ErrorResponse(c, http.StatusInternalServerError, "could not register user", err)
	}
	return uh.SuccessResponse(c, http.StatusCreated, "user was successfully registered", nil)
}

// SignIn godoc
// @Summary login user by credentials
// @Description Password should contain:
// @Description - minimum of one small case letter
// @Description - minimum of one upper case letter
// @Description - minimum of one digit
// @Description - minimum of one special character
// @Description - minimum 8 characters length
// @Description Response contain pair JWT tokens
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param login body UserRequestBody true "raw request body"
// @Success 200 {object} Response{data=TokensResponseBody}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Failure 422 {object} Response
// @Failure 500 {object} Response
// @Router /v1/sign-in [post]
func (uh UserHandler) SignIn(c echo.Context) error {
	var requestPayload UserRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return uh.ErrorResponse(c, http.StatusUnprocessableEntity, "could not read request body", err)
	}
	user, err := uh.authUsecase.GetActiveUserByEmail(requestPayload.Email)
	if err != nil {
		return uh.ErrorResponse(c, http.StatusNotFound, "user not found", err)
	}
	if _, err := uh.authUsecase.CheckPassword(user, requestPayload.Password); err != nil {
		return uh.ErrorResponse(c, http.StatusUnauthorized, "invalid credentials", err)
	}
	ts, err := uh.tokenUsecase.CreateToken(user.Uuid.String(), user.Role.String())
	if err != nil {
		return uh.ErrorResponse(c, http.StatusInternalServerError, "could not create a pair of JWT tokens", err)
	}
	return uh.SuccessResponse(c, http.StatusOK, "user was successfully logged in", ModelToResponseTokenDetails(ts))
}

// SignOut godoc
// @Summary logout user
// @Tags auth
// @Accept  json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /v1/sign-out [post]
func (uh UserHandler) SignOut(c echo.Context) error {
	accessTokenClaims := c.Get("accessTokenClaims").(*models.AccessTokenClaims)
	if err := uh.tokenUsecase.DropCacheTokens(accessTokenClaims.AccessUuid); err != nil {
		return uh.ErrorResponse(c, http.StatusInternalServerError, "could not logout user", err)
	}
	return uh.SuccessResponse(c, http.StatusOK, "user was successfully logged out", nil)
}

// GetUser godoc
// @Summary Get user
// @Description Get user
// @Tags users
// @Accept  json
// @Produce application/json
// @Security BearerAuth
// @Success 200 {object} Response{data=models.User}
// @Failure 401 {object} Response
// @Failure 403 {object} Response
// @Failure 500 {object} Response
// @Router /v1/users [get]
func (uh UserHandler) GetUser(c echo.Context) error {
	accessTokenClaims := c.Get("accessTokenClaims").(*models.AccessTokenClaims)
	user, err := uh.authUsecase.GetUserByUuid(accessTokenClaims.UserUuid)
	if err != nil {
		return uh.ErrorResponse(c, http.StatusInternalServerError, "could not get user", err)
	}
	return uh.SuccessResponse(c, http.StatusOK, "user was successfully found", ModelToResponse(user))
}

// Confirm godoc
// @Summary confirm registration/reset password with 6-digit code from email/sms
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param code body CodeRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 422 {object} Response
// @Failure 500 {object} Response
// @Router /v1/confirm [post]
func (uh UserHandler) Confirm(c echo.Context) error {
	var requestPayload CodeRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return uh.ErrorResponse(c, http.StatusUnprocessableEntity, "could not read request body", err)
	}
	if err := c.Validate(requestPayload); err != nil {
		return err
	}
	code, err := uh.authUsecase.GetCode(requestPayload.Code)
	if err != nil {
		return uh.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
	}
	if err := uh.authUsecase.Confirm(code); err != nil {
		return uh.ErrorResponse(c, http.StatusInternalServerError, "could not confirm user", err)
	}
	return uh.SuccessResponse(c, http.StatusOK, "code was successfully confirmed", nil)
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
// @Param registration body UserRequestBody true "raw request body"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 422 {object} Response
// @Failure 500 {object} Response
// @Router /v1/reset-password [post]
func (uh UserHandler) ResetPassword(c echo.Context) error {
	var requestPayload UserRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return uh.ErrorResponse(c, http.StatusUnprocessableEntity, "could not read request body", err)
	}
	if err := uh.authUsecase.ResetPassword(requestPayload.Email, requestPayload.Password); err != nil {
		return uh.ErrorResponse(c, http.StatusInternalServerError, "could not reset password", err)
	}
	return uh.SuccessResponse(c, http.StatusOK, "password was successfully reset", nil)
}
