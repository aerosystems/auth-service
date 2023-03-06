package repository

import (
	"os"
	"strconv"
	"time"

	"github.com/aerosystems/auth-service/internal/helpers"
	"github.com/aerosystems/auth-service/internal/models"
	"gorm.io/gorm"
)

type codeRepo struct {
	db *gorm.DB
}

func NewCodeRepo(db *gorm.DB) *codeRepo {
	return &codeRepo{
		db: db,
	}
}

func (r *codeRepo) FindAll() (*[]models.Code, error) {
	var codes []models.Code
	r.db.Find(&codes)
	return &codes, nil
}

func (r *codeRepo) FindByID(ID int) (*models.Code, error) {
	var code models.Code
	result := r.db.Find(&code, ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &code, nil
}

func (r *codeRepo) Create(code *models.Code) error {
	result := r.db.Create(&code)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *codeRepo) Update(code *models.Code) error {
	result := r.db.Save(&code)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *codeRepo) Delete(code *models.Code) error {
	result := r.db.Delete(&code)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *codeRepo) GetByCode(Code int) (*models.Code, error) {
	var code models.Code
	result := r.db.Where("code = ?", Code).First(&code)
	if result.Error != nil {
		return nil, result.Error
	}
	return &code, nil
}

func (r *codeRepo) GetLastIsActiveCode(UserID int, Action string) (*models.Code, error) {
	var code models.Code
	result := r.db.Where("user_id = ? AND action = ?", UserID, Action).First(&code)
	if result.Error != nil {
		return nil, result.Error
	}
	return &code, nil
}

func (r *codeRepo) ExtendExpiration(code *models.Code) error {
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
func (r *codeRepo) NewCode(UserID int, Action string, Data string) (*models.Code, error) {
	codeExpMinutes, err := strconv.Atoi(os.Getenv("CODE_EXP_MINUTES"))
	if err != nil {
		return nil, err
	}

	code := models.Code{
		Code:      helpers.GenCode(),
		UserID:    UserID,
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
