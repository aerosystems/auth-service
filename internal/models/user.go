package models

import (
	"github.com/google/uuid"
	"time"
)

type KindRole string

const (
	Customer KindRole = "customer"
	Staff    KindRole = "staff"
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

type UserRepository interface {
	GetById(Id int) (*User, error)
	GetByUserId(UserId int) (*User, error)
	GetByEmail(Email string) (*User, error)
	GetByGoogleId(GoogleId string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(user *User) error
}
