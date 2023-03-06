package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aerosystems/auth-service/internal/helpers"
	"net/http"

	"github.com/aerosystems/auth-service/internal/models"
)

type RefreshTokenRequestBody struct {
	RefreshToken string `json:"refresh_token" xml:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// RefreshToken godoc
// @Summary refresh pair JWT tokens
// @Tags auth
// @Accept  json
// @Accept  xml
// @Produce application/json
// @Produce application/xml
// @Param login body handlers.RefreshTokenRequestBody true "raw request body, should contain Refresh Token"
// @Param Authorization header string true "should contain Access Token, with the Bearer started"
// @Success 200 {object} Response{data=TokensResponseBody}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /tokens/refresh [post]
func (h *BaseHandler) RefreshToken(w http.ResponseWriter, r *http.Request) error {
	// receive AccessToken Claims from context middleware
	accessTokenClaims, ok := r.Context().Value(helpers.ContextKey("accessTokenClaims")).(*models.AccessTokenClaims)
	if !ok {
		err := errors.New("token is untracked")
		return WriteResponse(w, http.StatusUnauthorized, NewErrorPayload(err))
	}

	// getting Refresh Token from Redis cache
	cacheJSON, _ := h.tokensRepo.GetCacheValue(accessTokenClaims.AccessUUID)
	accessTokenCache := new(models.AccessTokenCache)
	err := json.Unmarshal([]byte(*cacheJSON), accessTokenCache)
	if err != nil {
		return WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
	}
	cacheRefreshTokenUUID := accessTokenCache.RefreshUUID

	var requestPayload RefreshTokenRequestBody

	if err := ReadRequest(w, r, &requestPayload); err != nil {
		return WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
	}

	// validate & parse refresh token claims
	refreshTokenClaims, err := h.tokensRepo.DecodeRefreshToken(requestPayload.RefreshToken)
	if err != nil {
		return WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
	}
	requestRefreshTokenUUID := refreshTokenClaims.RefreshUUID

	// drop Access & Refresh Tokens from Redis Cache
	_ = h.tokensRepo.DropCacheTokens(*accessTokenClaims)

	// compare RefreshToken UUID from Redis cache & Request body
	if requestRefreshTokenUUID != cacheRefreshTokenUUID {
		// drop request RefreshToken UUID from cache
		_ = h.tokensRepo.DropCacheKey(requestRefreshTokenUUID)
		err := errors.New("hmmm... refresh token in request body does not match refresh token which publish access token")
		return WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
	}

	// create pair JWT tokens
	ts, err := h.tokensRepo.CreateToken(refreshTokenClaims.UserID)
	if err != nil {
		return WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
	}

	// add refresh token UUID to cache
	err = h.tokensRepo.CreateCacheKey(refreshTokenClaims.UserID, ts)
	if err != nil {
		return WriteResponse(w, http.StatusBadRequest, NewErrorPayload(err))
	}

	tokens := TokensResponseBody{
		AccessToken:  ts.AccessToken,
		RefreshToken: ts.RefreshToken,
	}

	payload := Response{
		Error:   false,
		Message: fmt.Sprintf("Tokens successfuly refreshed for User %d", refreshTokenClaims.UserID),
		Data:    tokens,
	}

	return WriteResponse(w, http.StatusAccepted, payload)
}
