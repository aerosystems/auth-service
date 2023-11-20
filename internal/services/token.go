package services

import (
	"encoding/json"
	"github.com/go-redis/redis/v7"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"time"
)

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
	AccessUuid string `json:"accessUuid"`
	UserUuid   string `json:"userUuid"`
	UserRole   string `json:"userRole"`
	Exp        int    `json:"exp"`
	jwt.StandardClaims
}

type RefreshTokenClaims struct {
	RefreshUuid string `json:"refreshUuid"`
	UserUuid    string `json:"userUuid"`
	UserRole    string `json:"userRole"`
	Exp         int    `json:"exp"`
	jwt.StandardClaims
}

type AccessTokenCache struct {
	UserUuid    string `json:"userUuid"`
	RefreshUuid string `json:"refreshUuid"`
}

type TokenService interface {
	GetAccessSecret() string
	CreateToken(userUuid string, userRole string) (*TokenDetails, error)
	DecodeRefreshToken(tokenString string) (*RefreshTokenClaims, error)
	DecodeAccessToken(tokenString string) (*AccessTokenClaims, error)
	DropCacheTokens(accessTokenClaims AccessTokenClaims) error
	DropCacheKey(Uuid string) error
	GetCacheValue(Uuid string) (*string, error)
}

type TokenServiceImpl struct {
	cache             *redis.Client
	accessSecret      string
	refreshSecret     string
	accessExpMinutes  int
	refreshExpMinutes int
}

func NewTokenServiceImpl(cache *redis.Client, accessSecret string, refreshSecret string, accessExpMinutes int, refreshExpMinutes int) *TokenServiceImpl {
	return &TokenServiceImpl{
		cache:             cache,
		accessSecret:      accessSecret,
		refreshSecret:     refreshSecret,
		accessExpMinutes:  accessExpMinutes,
		refreshExpMinutes: refreshExpMinutes,
	}
}

func (r *TokenServiceImpl) GetAccessSecret() string {
	return r.accessSecret
}

// DropCacheKey function that will be used to drop the JWTs metadata from Redis
func (r *TokenServiceImpl) DropCacheKey(Uuid string) error {
	err := r.cache.Del(Uuid).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetCacheValue function that will be used to get the JWTs metadata from Redis
func (r *TokenServiceImpl) GetCacheValue(Uuid string) (*string, error) {
	value, err := r.cache.Get(Uuid).Result()
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// CreateToken returns JWT Token
func (r *TokenServiceImpl) CreateToken(userUuid string, userRole string) (*TokenDetails, error) {
	td := &TokenDetails{}
	var err error

	td.AtExpires = time.Now().Add(time.Minute * time.Duration(r.accessExpMinutes)).Unix()
	td.AccessUuid = uuid.New()

	td.RtExpires = time.Now().Add(time.Minute * time.Duration(r.refreshExpMinutes)).Unix()
	td.RefreshUuid = uuid.New()

	atClaims := jwt.MapClaims{}
	atClaims["accessUuid"] = td.AccessUuid.String()
	atClaims["userUuid"] = userUuid
	atClaims["userRole"] = userRole
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(r.accessSecret))
	if err != nil {
		return nil, err
	}
	rtClaims := jwt.MapClaims{}
	rtClaims["refreshUuid"] = td.RefreshUuid.String()
	rtClaims["userUuid"] = userUuid
	rtClaims["userRole"] = userRole
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(r.refreshSecret))
	if err != nil {
		return nil, err
	}
	// add a refresh token Uuid to cache
	if err = r.createCacheKey(userUuid, td); err != nil {
		return nil, err
	}
	return td, nil
}

func (r *TokenServiceImpl) DecodeRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.refreshSecret), nil
	})
	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func (r *TokenServiceImpl) DecodeAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.accessSecret), nil
	})
	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func (r *TokenServiceImpl) DropCacheTokens(accessTokenClaims AccessTokenClaims) error {
	cacheJSON, _ := r.GetCacheValue(accessTokenClaims.AccessUuid)
	accessTokenCache := new(AccessTokenCache)
	err := json.Unmarshal([]byte(*cacheJSON), accessTokenCache)
	if err != nil {
		return err
	}
	// drop refresh token from Redis cache
	err = r.DropCacheKey(accessTokenCache.RefreshUuid)
	if err != nil {
		return err
	}
	// drop access token from Redis cache
	err = r.DropCacheKey(accessTokenClaims.AccessUuid)
	if err != nil {
		return err
	}
	return nil
}

// createCacheKey function that will be used to save the JWTs metadata in Redis
func (r *TokenServiceImpl) createCacheKey(userUuid string, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0) //converting Unix to UTC(to Time object)
	now := time.Now()
	cacheJSON, err := json.Marshal(AccessTokenCache{
		UserUuid:    userUuid,
		RefreshUuid: td.RefreshUuid.String(),
	})
	if err != nil {
		return err
	}
	if err := r.cache.Set(td.AccessUuid.String(), cacheJSON, at.Sub(now)).Err(); err != nil {
		return err
	}
	if err := r.cache.Set(td.RefreshUuid.String(), userUuid, rt.Sub(now)).Err(); err != nil {
		return err
	}
	return nil
}
