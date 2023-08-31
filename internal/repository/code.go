package repository

import (
	"os"
	"strconv"
	"time"

	"github.com/aerosystems/auth-service/internal/helpers"
	"github.com/aerosystems/auth-service/internal/models"
	"gorm.io/gorm"
)

type CodeRepo struct {
	db *gorm.DB
}

func NewCodeRepo(db *gorm.DB) *CodeRepo {
	return &CodeRepo{
		db: db,
	}
}

func (r *CodeRepo) FindAll() (*[]models.Code, error) {
	var codes []models.Code
	r.db.Find(&codes)
	return &codes, nil
}

func (r *CodeRepo) FindByID(ID int) (*models.Code, error) {
	var code models.Code
	result := r.db.Find(&code, ID)
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
	result := r.db.Preload("User").Where("code = ?", value).First(&code)
	if result.Error != nil {
		return nil, result.Error
	}
	return &code, nil
}

func (r *CodeRepo) GetLastIsActiveCode(UserID int, Action string) (*models.Code, error) {
	var code models.Code
	result := r.db.Where("user_id = ? AND action = ?", UserID, Action).First(&code)
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

// NewCode CreateCode generation new code
func (r *CodeRepo) NewCode(User models.User, Action string, Data string) (*models.Code, error) {
	codeExpMinutes, err := strconv.Atoi(os.Getenv("CODE_EXP_MINUTES"))
	if err != nil {
		return nil, err
	}

	code := models.Code{
		Code:      helpers.GenCode(),
		User:      User,
		CreatedAt: time.Now(),
		ExpireAt:  time.Now().Add(time.Minute * time.Duration(codeExpMinutes)),
		Action:    Action,
		Data:      Data,
		IsUsed:    false,
	}

	result := r.db.Create(&code)
	if result.Error != nil {
		return nil, result.Error
	}
	return &code, nil
}
