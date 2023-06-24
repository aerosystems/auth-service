package models

import "time"

type Code struct {
	ID        int       `json:"id" gorm:"primaryKey;unique;autoIncrement"`
	Code      int       `json:"code"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpireAt  time.Time `json:"expire_at"`
	Action    string    `json:"action"`
	Data      string    `json:"data"`
	IsUsed    bool      `json:"is_used"`
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
