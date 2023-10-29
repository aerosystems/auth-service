package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type UserRequestBody struct {
	Email    string `json:"email" example:"example@gmail.com"`
	Password string `json:"password" example:"P@ssw0rd" validate:"gte=8"`
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
// @Param registration body SuccessResponse true "raw request body"
// @Success 201 {object} Response
// @Failure 400 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/sign-up [post]
func (h *BaseHandler) SignUp(c echo.Context) error {
	requestPayload := new(UserRequestBody)
	if err := c.Bind(&requestPayload); err != nil {
		return h.ErrorResponse(c, http.StatusUnprocessableEntity, "could not read request body", err)
	}
	if err := c.Validate(requestPayload); err != nil {
		return err
	}
	if err := h.userService.Register(requestPayload.Email, requestPayload.Password, c.RealIP()); err != nil {
		return h.ErrorResponse(c, http.StatusInternalServerError, "could not register user", err)
	}
	return h.SuccessResponse(c, http.StatusCreated, "user was successfully registered", nil)
}
