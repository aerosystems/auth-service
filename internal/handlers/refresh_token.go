package handlers

import (
	"net/http"
)

type RefreshTokenRequestBody struct {
	RefreshToken string `json:"refresh_token" xml:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// RefreshToken godoc
// @Summary refresh a pair of JWT tokens
// @Tags auth
// @Accept  json
// @Produce application/json
// @Param login body handlers.RefreshTokenRequestBody true "raw request body, should contain Refresh Token"
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Success 200 {object} Response{data=TokensResponseBody}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /token/refresh [post]
func (h *BaseHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var requestPayload RefreshTokenRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400001, "request payload is incorrect", err))
		return
	}

	// validate & parse refresh token claims
	refreshTokenClaims, err := h.tokenService.DecodeRefreshToken(requestPayload.RefreshToken)
	if err != nil {
		_ = WriteResponse(w, http.StatusBadRequest, NewErrorPayload(400015, "could not validate & parse Refresh Token claims", err))
		return
	}

	// create pair JWT tokens
	ts, err := h.tokenService.CreateToken(refreshTokenClaims.UserID)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500003, "could not to create a pair of JWT Tokens", err))
		return
	}

	// add refresh token UUID to cache
	err = h.tokenService.CreateCacheKey(refreshTokenClaims.UserID, ts)
	if err != nil {
		_ = WriteResponse(w, http.StatusInternalServerError, NewErrorPayload(500004, "could not to add a Refresh Token UUID to cache", err))
		return
	}

	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}

	payload := NewResponsePayload("tokens successfully refreshed", tokens)

	_ = WriteResponse(w, http.StatusOK, payload)
	return
}
