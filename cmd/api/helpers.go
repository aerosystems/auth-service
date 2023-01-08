package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/twinj/uuid"
)

// jsonResponse is the type used for sending JSON around
type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// TokenDetails is the structure which holds data with JWT tokens
type tokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

// createToken returns JWT Token
func (app *Config) createToken(userid int) (*tokenDetails, error) {
	td := &tokenDetails{}

	access_exp_minutes, err := strconv.Atoi(os.Getenv("ACCESS_EXP_MINUTES"))
	if err != nil {
		return nil, err
	}

	refresh_exp_minutes, err := strconv.Atoi(os.Getenv("REFRESH_EXP_MINUTES"))
	if err != nil {
		return nil, err
	}

	td.AtExpires = time.Now().Add(time.Minute * time.Duration(access_exp_minutes)).Unix()
	td.AccessUuid = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(time.Minute * time.Duration(refresh_exp_minutes)).Unix()
	td.RefreshUuid = uuid.NewV4().String()

	atClaims := jwt.MapClaims{}
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userid
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}
	return td, nil
}

// createAuth: function that will be used to save the JWTs metadata
func (app *Config) createAuth(userid int, td *tokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0) //converting Unix to UTC(to Time object)
	now := time.Now()

	errAccess := app.Cache.Set(td.AccessUuid, strconv.Itoa(int(userid)), at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}
	errRefresh := app.Cache.Set(td.RefreshUuid, strconv.Itoa(int(userid)), rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}
	return nil
}

// readJSON tries to read the body of a request and converts it into JSON
func (app *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // one megabyte
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only a single json value")
	}

	return nil
}

// writeJSON takes a response status code and arbitrary data and writes a json response to the client
func (app *Config) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

// errorJSON takes an error, and optionally a response status code, and generates and sends
// a json error response
func (app *Config) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, statusCode, payload)
}

func (app *Config) validateEmail(data string) (string, error) {
	email, err := mail.ParseAddress(data)
	if err != nil {
		return "", err
	}

	return email.Address, nil
}

func (app *Config) normalizeEmail(data string) string {
	addr := strings.ToLower(data)

	arrAddr := strings.Split(addr, "@")
	username := arrAddr[0]
	domain := arrAddr[1]

	googleDomains := []string{"gmail.com", "googlemail.com"}

	//checking google mail aliases
	if Contains(googleDomains, domain) {
		//removing all dots from username mail
		username = strings.ReplaceAll(username, ".", "")
		//removing all characters after +
		if strings.Contains(username, "+") {
			res := strings.Split(username, "+")
			username = res[0]
		}
		addr = username + "@gmail.com"
	}

	return addr
}

func (app *Config) validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password should be of 8 characters long")
	}
	done, err := regexp.MatchString("([a-z])+", password)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("password should contain atleast one lower case character")
	}
	done, err = regexp.MatchString("([A-Z])+", password)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("password should contain atleast one upper case character")
	}
	done, err = regexp.MatchString("([0-9])+", password)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("password should contain atleast one digit")
	}

	done, err = regexp.MatchString("([!@#$%^&*.?-])+", password)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("password should contain atleast one special character")
	}
	return nil
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
