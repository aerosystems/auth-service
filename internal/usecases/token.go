package usecases

import (
	"encoding/json"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/go-redis/redis/v7"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"time"
)

type TokenUsecase struct {
	cache             *redis.Client
	accessSecret      string
	refreshSecret     string
	accessExpMinutes  int
	refreshExpMinutes int
}

func NewTokenUsecase(cache *redis.Client, accessSecret string, refreshSecret string, accessExpMinutes int, refreshExpMinutes int) *TokenUsecase {
	return &TokenUsecase{
		cache:             cache,
		accessSecret:      accessSecret,
		refreshSecret:     refreshSecret,
		accessExpMinutes:  accessExpMinutes,
		refreshExpMinutes: refreshExpMinutes,
	}
}

type AccessTokenCache struct {
	UserUuid    string `json:"userUuid"`
	RefreshUuid string `json:"refreshUuid"`
}

func (r *TokenUsecase) GetAccessSecret() string {
	return r.accessSecret
}

// DropCacheKey function that will be used to drop the JWTs metadata from Redis
func (r *TokenUsecase) DropCacheKey(Uuid string) error {
	err := r.cache.Del(Uuid).Err()
	if err != nil {
		return err
	}
	return nil
}

// GetCacheValue function that will be used to get the JWTs metadata from Redis
func (r *TokenUsecase) GetCacheValue(Uuid string) (*string, error) {
	value, err := r.cache.Get(Uuid).Result()
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// CreateToken returns JWT Token
func (r *TokenUsecase) CreateToken(userUuid string, userRole string) (*models.TokenDetails, error) {
	td := &models.TokenDetails{}
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

func (r *TokenUsecase) DecodeRefreshToken(tokenString string) (*models.RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.refreshSecret), nil
	})
	if claims, ok := token.Claims.(*models.RefreshTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func (r *TokenUsecase) DecodeAccessToken(tokenString string) (*models.AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(r.accessSecret), nil
	})
	if claims, ok := token.Claims.(*models.AccessTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func (r *TokenUsecase) DropCacheTokens(accessTokenClaims models.AccessTokenClaims) error {
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
func (r *TokenUsecase) createCacheKey(userUuid string, td *models.TokenDetails) error {
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
