package repository

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/aerosystems/nix-trainee-5-6-7-8/internal/models"
	"github.com/go-redis/redis/v7"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type tokensRepo struct {
	cache *redis.Client
}

func NewTokensRepo(cache *redis.Client) *tokensRepo {
	return &tokensRepo{
		cache: cache,
	}
}

// DropCacheKey: function that will be used to drop the JWTs metadata from Redis
func (r *tokensRepo) DropCacheKey(UUID string) error {
	err := r.cache.Del(UUID).Err()
	if err != nil {
		return err
	}
	return nil
}

// createCacheKey: function that will be used to save the JWTs metadata in Redis
func (r *tokensRepo) CreateCacheKey(userID int, td *models.TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0) //converting Unix to UTC(to Time object)
	now := time.Now()

	cacheJSON, err := json.Marshal(models.AccessTokenCache{
		UserID:      userID,
		RefreshUUID: td.RefreshUuid.String(),
	})
	if err != nil {
		return err
	}

	err = r.cache.Set(td.AccessUuid.String(), cacheJSON, at.Sub(now)).Err()
	if err != nil {
		return err
	}
	err = r.cache.Set(td.RefreshUuid.String(), strconv.Itoa(userID), rt.Sub(now)).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *tokensRepo) GetCacheValue(UUID string) (*string, error) {
	value, err := r.cache.Get(UUID).Result()
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// CreateToken returns JWT Token
func (r *tokensRepo) CreateToken(userid int) (*models.TokenDetails, error) {
	td := &models.TokenDetails{}

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

func (r *tokensRepo) DecodeRefreshToken(tokenString string) (*models.RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})

	if claims, ok := token.Claims.(*models.RefreshTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func (r *tokensRepo) DecodeAccessToken(tokenString string) (*models.AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	if claims, ok := token.Claims.(*models.AccessTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func (r *tokensRepo) DropCacheTokens(accessTokenClaims models.AccessTokenClaims) error {
	cacheJSON, _ := r.GetCacheValue(accessTokenClaims.AccessUUID)
	accessTokenCache := new(models.AccessTokenCache)
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
