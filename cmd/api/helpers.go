package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

// jsonResponse is the type used for sending JSON around
type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// TokenDetails is the structure which holds data with JWT tokens
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   uuid.UUID
	RefreshUuid  uuid.UUID
	AtExpires    int64
	RtExpires    int64
}

type AccessTokenClaims struct {
	AccessUUID string `json:"access_uuid"`
	UserID     int    `json:"user_id"`
	Exp        int    `json:"exp"`
	jwt.StandardClaims
}

type RefreshTokenClaims struct {
	RefreshUUID string `json:"refresh_uuid"`
	UserID      int    `json:"user_id"`
	Exp         int    `json:"exp"`
	jwt.StandardClaims
}

type AccessTokenCache struct {
	UserID      int    `json:"user_id"`
	RefreshUUID string `json:"refresh_uuid"`
}

// createToken returns JWT Token
func (app *Config) createToken(userid int) (*TokenDetails, error) {
	td := &TokenDetails{}

	accessExpMinutes, err := strconv.Atoi(os.Getenv("ACCESS_EXP_MINUTES"))
	if err != nil {
		return nil, err
	}

	refreshExpMinutes, err := strconv.Atoi(os.Getenv("REFRESH_EXP_MINUTES"))
	if err != nil {
		return nil, err
	}

	td.AtExpires = time.Now().Add(time.Minute * time.Duration(accessExpMinutes)).Unix()
	td.AccessUuid = uuid.New()

	td.RtExpires = time.Now().Add(time.Minute * time.Duration(refreshExpMinutes)).Unix()
	td.RefreshUuid = uuid.New()

	atClaims := jwt.MapClaims{}
	atClaims["access_uuid"] = td.AccessUuid.String()
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

func (app *Config) decodeRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func (app *Config) decodeAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func (app *Config) dropCacheTokens(accessTokenClaims AccessTokenClaims) error {
	cacheJSON, _ := app.GetCacheValue(accessTokenClaims.AccessUUID)
	accessTokenCache := new(AccessTokenCache)
	err := json.Unmarshal([]byte(*cacheJSON), accessTokenCache)
	if err != nil {
		return err
	}
	// drop refresh token from Redis cache
	err = app.dropCacheKey(accessTokenCache.RefreshUUID)
	if err != nil {
		return err
	}
	// drop access token from Redis cache
	err = app.dropCacheKey(accessTokenClaims.AccessUUID)
	if err != nil {
		return err
	}

	return nil
}

// dropCacheKey: function that will be used to drop the JWTs metadata from Redis
func (app *Config) dropCacheKey(UUID string) error {
	err := app.Cache.Del(UUID).Err()
	if err != nil {
		return err
	}
	return nil
}

// createCacheKey: function that will be used to save the JWTs metadata in Redis
func (app *Config) createCacheKey(userID int, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0) //converting Unix to UTC(to Time object)
	now := time.Now()

	cacheJSON, err := json.Marshal(AccessTokenCache{
		UserID:      userID,
		RefreshUUID: td.RefreshUuid.String(),
	})
	if err != nil {
		return err
	}

	err = app.Cache.Set(td.AccessUuid.String(), cacheJSON, at.Sub(now)).Err()
	if err != nil {
		return err
	}
	err = app.Cache.Set(td.RefreshUuid.String(), strconv.Itoa(userID), rt.Sub(now)).Err()
	if err != nil {
		return err
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

	googleDomains := strings.Split(os.Getenv("GOOGLEMAIL_DOMAINS"), ",")

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

func (app *Config) validateRole(role string) error {
	trustRoles := strings.Split(os.Getenv("TRUST_ROLES"), ",")
	if !Contains(trustRoles, role) {
		return errors.New("role exists in trusted roles")
	}
	return nil
}

func (app *Config) validateCode(code int) error {
	count := 0
	for code > 0 {
		code = code / 10
		count++
		if count > 6 {
			return errors.New("code must contain 6 digits")
		}
	}
	if count != 6 {
		return errors.New("code must contain 6 digits")
	}
	return nil
}

func (app *Config) VerifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (app *Config) GetAccessTokenFromHeader(r *http.Request) (*string, error) {
	headers := r.Header
	_, ok := headers["Authorization"]
	if !ok {
		return nil, errors.New("request must contain Authorization Header")
	}

	rawData := strings.Split(r.Header.Get("Authorization"), "Bearer ")
	if len(rawData) != 2 {
		return nil, errors.New("authorization Header must contain Bearer format token")
	}
	accessToken := rawData[1]
	return &accessToken, nil

}

func (app *Config) GetCacheValue(UUID string) (*string, error) {
	value, err := app.Cache.Get(UUID).Result()
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
