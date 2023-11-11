package services

import (
	"encoding/json"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/go-redis/redis/v7"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"os"
	"strconv"
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
	AccessUUID string          `json:"accessUuid"`
	UserId     int             `json:"userId"`
	UserRole   models.KindRole `json:"userRole"`
	Exp        int             `json:"exp"`
	jwt.StandardClaims
}

type RefreshTokenClaims struct {
	RefreshUUID string          `json:"refreshUuid"`
	UserId      int             `json:"userId"`
	UserRole    models.KindRole `json:"userRole"`
	Exp         int             `json:"exp"`
	jwt.StandardClaims
}

type AccessTokenCache struct {
	UserId      int    `json:"userId"`
	RefreshUUID string `json:"refreshUuid"`
}

type TokenService interface {
	CreateToken(userId int, userRole models.KindRole) (*TokenDetails, error)
	DecodeRefreshToken(tokenString string) (*RefreshTokenClaims, error)
	DecodeAccessToken(tokenString string) (*AccessTokenClaims, error)
	DropCacheTokens(accessTokenClaims AccessTokenClaims) error
	DropCacheKey(UUID string) error
	GetCacheValue(UUID string) (*string, error)
}

type TokenServiceImpl struct {
	cache *redis.Client
}

func NewTokenServiceImpl(cache *redis.Client) *TokenServiceImpl {
	return &TokenServiceImpl{
		cache: cache,
	}
}

// DropCacheKey function that will be used to drop the JWTs metadata from Redis
func (r *TokenServiceImpl) DropCacheKey(UUID string) error {
	err := r.cache.Del(UUID).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *TokenServiceImpl) GetCacheValue(UUID string) (*string, error) {
	value, err := r.cache.Get(UUID).Result()
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// CreateToken returns JWT Token
func (r *TokenServiceImpl) CreateToken(userId int, userRole models.KindRole) (*TokenDetails, error) {
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
	atClaims["accessUuid"] = td.AccessUuid.String()
	atClaims["userId"] = userId
	atClaims["userRole"] = userRole
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}
	rtClaims := jwt.MapClaims{}
	rtClaims["refreshUuid"] = td.RefreshUuid.String()
	rtClaims["userId"] = userId
	rtClaims["userRole"] = userRole
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}
	// add a refresh token UUID to cache
	if err = r.createCacheKey(userId, td); err != nil {
		return nil, err
	}
	return td, nil
}

func (r *TokenServiceImpl) DecodeRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func (r *TokenServiceImpl) DecodeAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func (r *TokenServiceImpl) DropCacheTokens(accessTokenClaims AccessTokenClaims) error {
	cacheJSON, _ := r.GetCacheValue(accessTokenClaims.AccessUUID)
	accessTokenCache := new(AccessTokenCache)
	err := json.Unmarshal([]byte(*cacheJSON), accessTokenCache)
	if err != nil {
		return err
	}
	// drop refresh token from Redis cache
	err = r.DropCacheKey(accessTokenCache.RefreshUUID)
	if err != nil {
		return err
	}
	// drop access token from Redis cache
	err = r.DropCacheKey(accessTokenClaims.AccessUUID)
	if err != nil {
		return err
	}
	return nil
}

// createCacheKey function that will be used to save the JWTs metadata in Redis
func (r *TokenServiceImpl) createCacheKey(userId int, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0) //converting Unix to UTC(to Time object)
	now := time.Now()
	cacheJSON, err := json.Marshal(AccessTokenCache{
		UserId:      userId,
		RefreshUUID: td.RefreshUuid.String(),
	})
	if err != nil {
		return err
	}
	if err := r.cache.Set(td.AccessUuid.String(), cacheJSON, at.Sub(now)).Err(); err != nil {
		return err
	}
	if err := r.cache.Set(td.RefreshUuid.String(), strconv.Itoa(userId), rt.Sub(now)).Err(); err != nil {
		return err
	}
	return nil
}
