package models

import (
	"gorm.io/gorm"
	"time"
)

type KindCode string

const (
	Registration  KindCode = "registration"
	ResetPassword KindCode = "resetPassword"
)

type Code struct {
	gorm.Model
	Id        int       `json:"-" gorm:"primaryKey;unique;autoIncrement"`
	Code      string    `json:"code"`
	UserId    int       `json:"-"`
	User      User      `json:"-" gorm:"foreignKey:UserId"`
	Action    KindCode  `json:"-"`
	Data      string    `json:"-"`
	IsUsed    bool      `json:"-"`
	ExpireAt  time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type CodeRepository interface {
	GetById(Id int) (*Code, error)
	GetByCode(value string) (*Code, error)
	GetLastIsActiveCode(UserId int, Action string) (*Code, error)
	Create(code *Code) error
	Update(code *Code) error
	UpdateWithAssociations(code *Code) error
	ExtendExpiration(code *Code) error
	Delete(code *Code) error
}
