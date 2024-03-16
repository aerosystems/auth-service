package handlers

import (
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/labstack/echo/v4"
	"net/http"
)

type TokenHandler struct {
	*BaseHandler
	tokenUsecase TokenUsecase
}

func NewTokenHandler(baseHandler *BaseHandler, tokenUsecase TokenUsecase) *TokenHandler {
	return &TokenHandler{
		BaseHandler:  baseHandler,
		tokenUsecase: tokenUsecase,
	}
}

type TokensResponseBody struct {
	AccessToken  string `json:"accessToken" validate:"required,jwt" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
	RefreshToken string `json:"refreshToken" validate:"required,jwt" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

type RefreshTokenRequestBody struct {
	RefreshToken string `json:"refreshToken" validate:"required,jwt" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// ValidateToken godoc
// @Summary validate token
// @Tags api-gateway-special
// @Accept  json
// @Produce application/json
// @Security BearerAuth
// @Success 204 {object} Response
// @Failure 401 {object} Response
// @Router /v1/token/validate [get]
func (th TokenHandler) ValidateToken(c echo.Context) error {
	accessTokenClaims := c.Get("accessTokenClaims").(*models.AccessTokenClaims)
	if _, err := th.tokenUsecase.GetCacheValue(accessTokenClaims.AccessUuid); err != nil {
		return th.ErrorResponse(c, http.StatusUnauthorized, "invalid token", err)
	}
	return th.SuccessResponse(c, http.StatusNoContent, "", nil)
}

// RefreshToken godoc
// @Summary refresh a pair of JWT tokens
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param login body TokensResponseBody true "raw request body, should contain Refresh Token"
// @Success 200 {object} Response{data=TokensResponseBody}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 422 {object} Response
// @Failure 500 {object} Response
// @Router /v1/token/refresh [post]
func (th TokenHandler) RefreshToken(c echo.Context) error {
	var requestPayload RefreshTokenRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return th.ErrorResponse(c, http.StatusUnprocessableEntity, "could not read request body", err)
	}
	refreshTokenClaims, err := th.tokenUsecase.DecodeRefreshToken(requestPayload.RefreshToken)
	if err != nil {
		return th.ErrorResponse(c, http.StatusUnauthorized, "invalid refresh token", err)
	}
	ts, err := th.tokenUsecase.CreateToken(refreshTokenClaims.UserUuid, refreshTokenClaims.UserRole)
	if err != nil {
		return th.ErrorResponse(c, http.StatusInternalServerError, "could not create a pair of JWT tokens", err)
	}
	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}
	return th.SuccessResponse(c, http.StatusOK, "tokens were successfully refreshed", tokens)
}
