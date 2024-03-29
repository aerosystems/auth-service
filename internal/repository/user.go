package repository

import (
	"errors"
	"github.com/aerosystems/auth-service/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) GetById(Id int) (*models.User, error) {
	var user models.User
	result := r.db.Find(&user, Id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) GetByEmail(Email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", Email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) GetByUserId(UserId int) (*models.User, error) {
	var user models.User
	result := r.db.Where("user_id = ?", UserId).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) GetByUuid(Uuid uuid.UUID) (*models.User, error) {
	var user models.User
	result := r.db.Where("uuid = ?", Uuid.String()).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) GetByGoogleId(GoogleId string) (*models.User, error) {
	var user models.User
	result := r.db.Where("google_id = ?", GoogleId).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) Create(user *models.User) error {
	result := r.db.Create(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepo) Update(user *models.User) error {
	result := r.db.Save(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *UserRepo) Delete(user *models.User) error {
	result := r.db.Delete(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
