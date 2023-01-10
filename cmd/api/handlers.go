package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/aerosystems/auth-service/data"
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

	}

	err = app.createAuth(user.ID, ts)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)

	}

	// log request
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
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
			payload = jsonResponse{
				Error:   false,
				Message: fmt.Sprintf("Updated user with Id: %d", user.ID),
				Data:    user,
			}
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
	payload = jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Registered user with Id: %d", newUserId),
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

	payload := jsonResponse{
		Error: false,
		// Message: fmt.Sprintf("Succesfuly confirmed registration user with Id: %d", user.ID),
		// Data:    user,
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
