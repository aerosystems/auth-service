package models

import "time"

type Code struct {
	Id        uint      `json:"id" gorm:"primaryKey;unique;autoIncrement"`
	Code      string    `json:"code"`
	UserID    int       `json:"userId"`
	User      User      `json:"user" gorm:"foreignKey:UserID"` // Relation to User [Belongs To Association]
	Action    string    `json:"action"`
	Data      string    `json:"data"`
	IsUsed    bool      `json:"isUsed"`
	ExpireAt  time.Time `json:"expireAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CodeRepository interface {
	GetById(Id uint) (*Code, error)
	GetByCode(value string) (*Code, error)
	GetLastIsActiveCode(UserId uint, Action string) (*Code, error)
	Create(code *Code) error
	Update(code *Code) error
	UpdateWithAssociations(code *Code) error
	ExtendExpiration(code *Code) error
	Delete(code *Code) error
}
