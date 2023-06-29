package models

import "time"

type Code struct {
	ID        int       `json:"id" gorm:"primaryKey;unique;autoIncrement"`
	Code      int       `json:"code"`
	User      User      `json:"userId" gorm:"foreignKey:UserID"`
	CreatedAt time.Time `json:"createdAt"`
	ExpireAt  time.Time `json:"expireAt"`
	Action    string    `json:"action"`
	Data      string    `json:"data"`
	IsUsed    bool      `json:"isUsed"`
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
	NewCode(User User, Action string, Data string) (*Code, error)
}
