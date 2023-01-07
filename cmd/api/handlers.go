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

	// validate against database
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		_ = app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		_ = app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	// log request
	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
	}

	token, err := app.createToken(user.ID)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", requestPayload.Email),
		Data:    token,
	}

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) Registration(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	newUser := data.User{
		Email:    requestPayload.Email,
		Password: requestPayload.Password,
		Active:   1,
		Role:     "startup",
	}

	// validate against database
	newUserId, err := app.Models.User.Insert(newUser)

	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Registered user with Id: %d", newUserId),
		Data:    newUser,
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
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
