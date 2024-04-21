package usecases

import (
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetByUuid(Uuid uuid.UUID) (*models.User, error)
	GetByEmail(Email string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(user *models.User) error
}

type CodeRepository interface {
	GetById(Id int) (*models.Code, error)
	GetByCode(value string) (*models.Code, error)
	GetLastIsActiveCode(UserId int, Action string) (*models.Code, error)
	Create(code *models.Code) error
	UpdateWithAssociations(code *models.Code) error
	Update(code *models.Code) error
}

type CheckmailAdapter interface {
	IsTrustEmail(email, clientIp string) (bool, error)
}

type MailAdapter interface {
	SendEmail(to, subject, body string) error
}

type CustomerAdapter interface {
	CreateCustomer() (uuid.UUID, error)
}
