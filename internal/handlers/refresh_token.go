package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// RefreshToken godoc
// @Summary refresh a pair of JWT tokens
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param login body handlers.RefreshTokenRequestBody true "raw request body, should contain Refresh Token"
// @Success 200 {object} Response{data=handlers.TokensResponseBody}
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 401 {object} handlers.ErrorResponse
// @Failure 422 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /v1/token/refresh [post]
func (h *BaseHandler) RefreshToken(c echo.Context) error {
	var requestPayload RefreshTokenRequestBody
	if err := c.Bind(&requestPayload); err != nil {
		return h.ErrorResponse(c, http.StatusUnprocessableEntity, "could not read request body", err)
	}
	refreshTokenClaims, err := h.tokenService.DecodeRefreshToken(requestPayload.RefreshToken)
	if err != nil {
		return h.ErrorResponse(c, http.StatusUnauthorized, "invalid refresh token", err)
	}
	ts, err := h.tokenService.CreateToken(refreshTokenClaims.UserUuid, refreshTokenClaims.UserRole)
	if err != nil {
		return h.ErrorResponse(c, http.StatusInternalServerError, "could not create a pair of JWT tokens", err)
	}
	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}
	return h.SuccessResponse(c, http.StatusOK, "tokens were successfully refreshed", tokens)
}
