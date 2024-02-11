package models

import (
	"gorm.io/gorm"
	"time"
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

type KindCode string

const (
	Registration  KindCode = "registration"
	ResetPassword KindCode = "resetPassword"
)

func (k KindCode) String() string {
	return string(k)
}
