package pg

import (
	"errors"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

type User struct {
	Id           int       `gorm:"primaryKey;unique;autoIncrement"`
	Uuid         string    `gorm:"unique"`
	Email        string    `gorm:"unique"`
	PasswordHash string    `gorm:"<-"`
	Role         string    `gorm:"<-"`
	IsActive     bool      `gorm:"<-"`
	GoogleId     string    `gorm:"unique"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (u *User) ToModel() *models.User {
	return &models.User{
		Id:           u.Id,
		Uuid:         uuid.MustParse(u.Uuid),
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         models.RoleFromString(u.Role),
		IsActive:     u.IsActive,
		GoogleId:     u.GoogleId,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func ModelToUserPg(user *models.User) *User {
	return &User{
		Id:           user.Id,
		Uuid:         user.Uuid.String(),
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         user.Role.String(),
		IsActive:     user.IsActive,
		GoogleId:     user.GoogleId,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func (r *UserRepo) GetByEmail(Email string) (*models.User, error) {
	var user User
	result := r.db.Where("email = ?", Email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return user.ToModel(), nil
}

func (r *UserRepo) GetByUuid(Uuid uuid.UUID) (*models.User, error) {
	var user User
	result := r.db.Where("uuid = ?", Uuid.String()).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return user.ToModel(), nil
}

func (r *UserRepo) Create(user *models.User) error {
	userPg := ModelToUserPg(user)
	result := r.db.Create(&userPg)
	if result.Error != nil {
		return result.Error
	}
	user = userPg.ToModel()
	return nil
}

func (r *UserRepo) Update(user *models.User) error {
	userPg := ModelToUserPg(user)
	result := r.db.Save(&userPg)
	if result.Error != nil {
		return result.Error
	}
	user = userPg.ToModel()
	return nil
}

func (r *UserRepo) Delete(user *models.User) error {
	userPg := ModelToUserPg(user)
	result := r.db.Delete(&userPg)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
