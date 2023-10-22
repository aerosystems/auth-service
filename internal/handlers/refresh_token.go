package handlers

import (
	"errors"
	"net/http"
)

type RefreshTokenRequestBody struct {
	RefreshToken string `json:"refreshToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// RefreshToken godoc
// @Summary refresh a pair of JWT tokens
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param login body handlers.RefreshTokenRequestBody true "raw request body, should contain Refresh Token"
// @Success 200 {object} Response{data=handlers.TokensResponseBody}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /v1/token/refresh [post]
func (h *BaseHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var requestPayload RefreshTokenRequestBody
	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusUnprocessableEntity, NewErrorPayload(422001, "could not read request body", err))
		return
	}
	if requestPayload.RefreshToken == "" {
		err := errors.New("refresh Token does not exists or empty")
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(422013, err.Error(), err))
		return
	}
	// validate & parse refresh token claims
	refreshTokenClaims, err := h.tokenService.DecodeRefreshToken(requestPayload.RefreshToken)
	if err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(422015, "could not validate refresh token claims", err))
		return
	}
	// create pair JWT tokens
	ts, err := h.tokenService.CreateToken(refreshTokenClaims.UserId, refreshTokenClaims.UserRole)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500003, "could not to create a pair of JWT Tokens", err))
		return
	}
	// add refresh token UUID to cache
	err = h.tokenService.CreateCacheKey(refreshTokenClaims.UserId, ts)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500004, "could not create Refresh Token", err))
		return
	}
	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}
	_ = WriteResponse(w, http.StatusOK, NewResponsePayload("tokens successfully refreshed", tokens))
	return
}
