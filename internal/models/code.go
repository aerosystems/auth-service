package models

import "time"

type Code struct {
	ID        int       `json:"id"  xml:"id"  gorm:"<-"`
	Code      int       `json:"code"  xml:"code"  gorm:"<-"`
	UserID    int       `json:"user_id" xml:"user_id"  gorm:"<-"`
	CreatedAt time.Time `json:"created_at" xml:"created_at"  gorm:"<-"`
	ExpireAt  time.Time `json:"expire_at" xml:"expire_at"  gorm:"<-"`
	Action    string    `json:"action" xml:"action"  gorm:"<-"`
	Data      string    `json:"data" xml:"data"  gorm:"<-"`
	IsUsed    bool      `json:"is_used" xml:"is_used"  gorm:"<-"`
}

type CodeRepository interface {
	FindAll() (*[]Code, error)
	FindByID(ID int) (*Code, error)
	Create(code *Code) error
	Update(code *Code) error
	Delete(code *Code) error
	GetByCode(Code int) (*Code, error)
	GetLastIsActiveCode(UserID int, Action string) (*Code, error)
	ExtendExpiration(code *Code) error
	NewCode(UserID int, Action string, Data string) (*Code, error)
}
