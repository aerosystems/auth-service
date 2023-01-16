package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aerosystems/auth-service/data"
	"golang.org/x/crypto/bcrypt"
)

// Authenticate accepts a json payload and attempts to authenticate a user
func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validating email
	addr, err := app.validateEmail(requestPayload.Email)
	if err != nil {
		err = errors.New("email is not valid")
		_ = app.errorJSON(w, err, http.StatusBadRequest)
	}

	// nomalizing email
	email := app.normalizeEmail(addr)

	// validating password
	err = app.validatePassword(requestPayload.Password)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate against database
	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		_ = app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		_ = app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	// create pair of JWT tokens
	ts, err := app.createToken(user.ID)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// add refresh token UUID to cache
	err = app.createCacheKey(user.ID, ts)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// log request
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	tokens := map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": ts.RefreshToken,
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", requestPayload.Email),
		Data:    tokens,
	}

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) Registration(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	addr, err := app.validateEmail(requestPayload.Email)
	if err != nil {
		err = errors.New("email is not valid")
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	email := app.normalizeEmail(addr)

	// Minimum of one small case letter
	// Minimum of one upper case letter
	// Minimum of one digit
	// Minimum of one special character
	// Minimum 8 characters length
	err = app.validatePassword(requestPayload.Password)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate user role
	err = app.validateRole(requestPayload.Role)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	var payload jsonResponse

	//checking if email is existing
	user, _ := app.Models.User.GetByEmail(email)
	if user != nil {
		if user.Active {
			err = errors.New("email already exists")
			_ = app.errorJSON(w, err, http.StatusBadRequest)
			return
		} else {
			// updating password for inactive user
			err := user.ResetPassword(requestPayload.Password)
			if err != nil {
				_ = app.errorJSON(w, err, http.StatusBadRequest)
				return
			}
			// updating other data for inactive user
			user.Role = requestPayload.Role
			err = user.Update()
			if err != nil {
				_ = app.errorJSON(w, err, http.StatusBadRequest)
				return
			}

			code, _ := app.Models.Code.GetLastActiveCode(user.ID, "registration")
			var XXXXXX int

			if code == nil {
				// generating confirmation code
				XXXXXX, err = app.Models.Code.CreateCode(user.ID, "registration", "")
				if err != nil {
					_ = app.errorJSON(w, err, http.StatusBadRequest)
					return
				}
			} else {
				// extend expiration code and return previous active code
				code.ExtendExpiration()
				XXXXXX = code.Code
			}

			payload = jsonResponse{
				Error:   false,
				Message: fmt.Sprintf("Updated user with Id: %d. Confirmation code: %d", user.ID, XXXXXX),
				Data:    user,
			}
			_ = app.writeJSON(w, http.StatusAccepted, payload)
			return
		}
	}

	// creating new inactive user
	newUser := data.User{
		Email:    email,
		Password: requestPayload.Password,
		Role:     requestPayload.Role,
	}
	newUserId, err := app.Models.User.Insert(newUser)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	newUser.ID = newUserId
	// generating confirmation code
	XXXXXX, err := app.Models.Code.CreateCode(newUserId, "registration", "")
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	payload = jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Registered user with Id: %d. Confirmation code: %d", newUserId, XXXXXX),
		Data:    newUser,
	}

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) Confirmation(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Code int `json:"code"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	err = app.validateCode(requestPayload.Code)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	code, err := app.Models.Code.GetByCode(requestPayload.Code)
	if err != nil {
		_ = app.errorJSON(w, errors.New("code is not found"), http.StatusNotFound)
		return
	}
	if code.Expiration.Before(time.Now()) {
		_ = app.errorJSON(w, errors.New("code is expired"), http.StatusNotFound)
		return
	}

	user, err := app.Models.User.GetOne(code.UserID)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	user.Active = true
	user.Update()

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Succesfuly confirmed registration user with Id: %d", user.ID),
		Data:    user,
	}

	_ = app.writeJSON(w, http.StatusAccepted, payload)

}

func (app *Config) Refresh(w http.ResponseWriter, r *http.Request) {
	// recieve AccessToken Claims from context middleware
	accessTokenClaims, ok := r.Context().Value(contextKey("accessTokenClaims")).(*AccessTokenClaims)
	if !ok {
		_ = app.errorJSON(w, errors.New("token is untracked"), http.StatusUnauthorized)
		return
	}

	// getting Refresh Token from Redis cache
	cacheJSON, _ := app.GetCacheValue(accessTokenClaims.AccessUUID)
	accessTokenCache := new(AccessTokenCache)
	err := json.Unmarshal([]byte(*cacheJSON), accessTokenCache)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	cacheRefreshTokenUUID := accessTokenCache.RefreshUUID

	var requestPayload struct {
		RefreshToken string `json:"refresh_token"`
	}

	err = app.readJSON(w, r, &requestPayload)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate & parse refresh token claims
	refreshTokenClaims, err := app.decodeRefreshToken(requestPayload.RefreshToken)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}
	requestRefreshTokenUUID := refreshTokenClaims.RefreshUUID

	// drop Access & Refresh Tokens from Redis Cache
	_ = app.dropCacheTokens(*accessTokenClaims)

	// compare RefreshToken UUID from Redis cache & Request body
	if requestRefreshTokenUUID != cacheRefreshTokenUUID {
		// drop request RefreshToken UUID from cache
		_ = app.dropCacheKey(requestRefreshTokenUUID)
		_ = app.errorJSON(w, errors.New("hmmm... refresh token in request body does not match refresh token which publish access token. is it scam?"))
		return
	}

	// create pair of JWT tokens
	ts, err := app.createToken(refreshTokenClaims.UserID)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// add refresh token UUID to cache
	err = app.createCacheKey(refreshTokenClaims.UserID, ts)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	tokens := map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": ts.RefreshToken,
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %d", refreshTokenClaims.UserID),
		Data:    tokens,
	}

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) Logout(w http.ResponseWriter, r *http.Request) {
	accessTokenClaims, ok := r.Context().Value(contextKey("accessTokenClaims")).(*AccessTokenClaims)
	if !ok {
		_ = app.errorJSON(w, errors.New("token is untracked"), http.StatusUnauthorized)
		return
	}

	err := app.dropCacheTokens(*accessTokenClaims)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("User %s successfully logged out", accessTokenClaims.AccessUUID),
		Data:    accessTokenClaims,
	}

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) Reset(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validating email
	addr, err := app.validateEmail(requestPayload.Email)
	if err != nil {
		err = errors.New("email is not valid")
		_ = app.errorJSON(w, err, http.StatusBadRequest)
	}

	// nomalizing email
	email := app.normalizeEmail(addr)

	// validating password
	err = app.validatePassword(requestPayload.Password)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate against database
	user, err := app.Models.User.GetByEmail(email)
	if err != nil {
		_ = app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestPayload.Password), 12)
	if err != nil {
		_ = app.errorJSON(w, errors.New("error creating password hash"))
		return
	}

	// generating confirmation code
	XXXXXX, err := app.Models.Code.CreateCode(user.ID, "reset", string(hashedPassword))
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	_ = XXXXXX

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("User %d initialize changing password", user.ID),
		Data:    user,
	}

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
	entry.Name = name
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceURL := "http://log-service/api/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
