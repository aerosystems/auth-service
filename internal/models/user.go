package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	Id           int       `json:"-" gorm:"primaryKey;unique;autoIncrement"`
	Uuid         uuid.UUID `json:"uuid" gorm:"unique"`
	Email        string    `json:"email" gorm:"unique"`
	PasswordHash string    `json:"-"`
	Role         KindRole  `json:"role"`
	IsActive     bool      `json:"-"`
	GoogleId     string    `json:"-"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}
