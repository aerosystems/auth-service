package TokenService

import (
	"encoding/json"
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
	AccessUUID string `json:"accessUuid"`
	UserID     int    `json:"userId"`
	Exp        int    `json:"exp"`
	jwt.StandardClaims
}

type RefreshTokenClaims struct {
	RefreshUUID string `json:"refreshUuid"`
	UserID      int    `json:"userId"`
	Exp         int    `json:"exp"`
	jwt.StandardClaims
}

type AccessTokenCache struct {
	UserID      int    `json:"userId"`
	RefreshUUID string `json:"refreshUuid"`
}

type Service struct {
	cache *redis.Client
}

func NewService(cache *redis.Client) *Service {
	return &Service{
		cache: cache,
	}
}

// DropCacheKey function that will be used to drop the JWTs metadata from Redis
func (r *Service) DropCacheKey(UUID string) error {
	err := r.cache.Del(UUID).Err()
	if err != nil {
		return err
	}
	return nil
}

// CreateCacheKey function that will be used to save the JWTs metadata in Redis
func (r *Service) CreateCacheKey(userID int, td *TokenDetails) error {
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

	if err := r.cache.Set(td.AccessUuid.String(), cacheJSON, at.Sub(now)).Err(); err != nil {
		return err
	}
	if err := r.cache.Set(td.RefreshUuid.String(), strconv.Itoa(userID), rt.Sub(now)).Err(); err != nil {
		return err
	}
	return nil
}

func (r *Service) GetCacheValue(UUID string) (*string, error) {
	value, err := r.cache.Get(UUID).Result()
	if err != nil {
		return nil, err
	}
	return &value, nil
}

// CreateToken returns JWT Token
func (r *Service) CreateToken(userid int) (*TokenDetails, error) {
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
	atClaims["userId"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	rtClaims := jwt.MapClaims{}
	rtClaims["refreshUuid"] = td.RefreshUuid.String()
	rtClaims["userId"] = userid
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}
	return td, nil
}

func (r *Service) DecodeRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func (r *Service) DecodeAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

func (r *Service) DropCacheTokens(accessTokenClaims AccessTokenClaims) error {
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
