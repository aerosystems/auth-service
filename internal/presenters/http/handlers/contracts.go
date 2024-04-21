package handlers

import "github.com/aerosystems/auth-service/internal/models"

type TokenUsecase interface {
	GetAccessSecret() string
	CreateToken(userUuid string, userRole string) (*models.TokenDetails, error)
	DecodeRefreshToken(tokenString string) (*models.RefreshTokenClaims, error)
	DecodeAccessToken(tokenString string) (*models.AccessTokenClaims, error)
	DropCacheTokens(accessUuid string) error
	DropCacheKey(Uuid string) error
	GetCacheValue(Uuid string) (*string, error)
}

type AuthUsecase interface {
	RegisterCustomer(email, password, clientIp string) error
	Confirm(code *models.Code) error
	ResetPassword(email, password string) error
	CheckPassword(user *models.User, password string) (bool, error)
	GetActiveUserByEmail(email string) (*models.User, error)
	GetUserByUuid(uuid string) (*models.User, error)
	GetCode(code string) (*models.Code, error)
}
