package repository

import (
	"errors"
	"github.com/aerosystems/auth-service/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"strconv"
	"time"
)

type CodeRepo struct {
	db *gorm.DB
}

func NewCodeRepo(db *gorm.DB) *CodeRepo {
	return &CodeRepo{
		db: db,
	}
}

func (r *CodeRepo) GetById(Id int) (*models.Code, error) {
	var code models.Code
	result := r.db.Find(&code, Id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &code, nil
}

func (r *CodeRepo) Create(code *models.Code) error {
	result := r.db.Create(&code)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *CodeRepo) Update(code *models.Code) error {
	result := r.db.Save(&code)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *CodeRepo) UpdateWithAssociations(code *models.Code) error {
	result := r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&code)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *CodeRepo) Delete(code *models.Code) error {
	result := r.db.Delete(&code)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *CodeRepo) GetByCode(value string) (*models.Code, error) {
	var code models.Code
	result := r.db.Preload(clause.Associations).Where("code = ?", value).Find(&code)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &code, nil
}

func (r *CodeRepo) GetLastIsActiveCode(UserId int, Action string) (*models.Code, error) {
	var code models.Code
	result := r.db.Where("user_id = ? AND action = ? AND is_used = ?", UserId, Action, false).First(&code)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &code, nil
}

func (r *CodeRepo) ExtendExpiration(code *models.Code) error {
	codeExpMinutes, err := strconv.Atoi(os.Getenv("CODE_EXP_MINUTES"))
	if err != nil {
		return err
	}
	code.ExpireAt = time.Now().Add(time.Minute * time.Duration(codeExpMinutes))
	result := r.db.Save(&code)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
