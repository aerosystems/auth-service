package usecases

import (
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetById(Id int) (*models.User, error)
	GetByUserId(UserId int) (*models.User, error)
	GetByUuid(Uuid uuid.UUID) (*models.User, error)
	GetByEmail(Email string) (*models.User, error)
	GetByGoogleId(GoogleId string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(user *models.User) error
}

type CodeRepository interface {
	GetById(Id int) (*models.Code, error)
	GetByCode(value string) (*models.Code, error)
	GetLastIsActiveCode(UserId int, Action string) (*models.Code, error)
	Create(code *models.Code) error
	Update(code *models.Code) error
	UpdateWithAssociations(code *models.Code) error
	ExtendExpiration(code *models.Code) error
	Delete(code *models.Code) error
}

type CheckmailRepo interface {
	IsTrustEmail(email, clientIp string) (bool, error)
}

type MailRepo interface {
	SendEmail(to, subject, body string) error
}

type CustomerRepo interface {
	CreateCustomer() (uuid.UUID, error)
}
