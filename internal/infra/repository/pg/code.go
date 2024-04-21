package pg

import (
	"errors"
	"github.com/aerosystems/auth-service/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type CodeRepo struct {
	db             *gorm.DB
	codeExpMinutes int
}

func NewCodeRepo(db *gorm.DB, codeExpMinutes int) *CodeRepo {
	return &CodeRepo{
		db:             db,
		codeExpMinutes: codeExpMinutes,
	}
}

type Code struct {
	Id        int       `gorm:"primaryKey;unique;autoIncrement"`
	Code      string    `gorm:"<-"`
	UserId    int       `gorm:"<-"`
	User      User      `gorm:"foreignKey:UserId"`
	Action    string    `gorm:"<-"`
	Data      string    `gorm:"<-"`
	IsUsed    bool      `gorm:"<-"`
	ExpireAt  time.Time `gorm:"<-"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (c *Code) ToModel() *models.Code {
	return &models.Code{
		Id:        c.Id,
		Code:      c.Code,
		UserId:    c.UserId,
		User:      *c.User.ToModel(),
		Action:    models.CodeFromString(c.Action),
		Data:      c.Data,
		IsUsed:    c.IsUsed,
		ExpireAt:  c.ExpireAt,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func ModelToCodePg(code *models.Code) *Code {
	return &Code{
		Id:        code.Id,
		Code:      code.Code,
		UserId:    code.UserId,
		Action:    code.Action.String(),
		Data:      code.Data,
		IsUsed:    code.IsUsed,
		ExpireAt:  code.ExpireAt,
		CreatedAt: code.CreatedAt,
		UpdatedAt: code.UpdatedAt,
	}
}

func (r *CodeRepo) GetById(Id int) (*models.Code, error) {
	var code Code
	result := r.db.Find(&code, Id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return code.ToModel(), nil
}

func (r *CodeRepo) Create(code *models.Code) error {
	code.ExpireAt = time.Now().Add(time.Minute * time.Duration(r.codeExpMinutes))
	codePg := ModelToCodePg(code)
	result := r.db.Create(&codePg)
	if result.Error != nil {
		return result.Error
	}
	code = codePg.ToModel()
	return nil
}

func (r *CodeRepo) UpdateWithAssociations(code *models.Code) error {
	codePg := ModelToCodePg(code)
	result := r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&codePg)
	if result.Error != nil {
		return result.Error
	}
	code = codePg.ToModel()
	return nil
}

func (r *CodeRepo) GetByCode(value string) (*models.Code, error) {
	var codePg Code
	result := r.db.Preload(clause.Associations).Where("code = ?", value).Find(&codePg)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return codePg.ToModel(), nil
}

func (r *CodeRepo) GetLastIsActiveCode(UserId int, Action string) (*models.Code, error) {
	var codePg Code
	result := r.db.Where("user_id = ? AND action = ? AND is_used = ?", UserId, Action, false).First(&codePg)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return codePg.ToModel(), nil
}

func (r *CodeRepo) Update(code *models.Code) error {
	codePg := ModelToCodePg(code)
	result := r.db.Save(&codePg)
	if result.Error != nil {
		return result.Error
	}
	code = codePg.ToModel()
	return nil
}
