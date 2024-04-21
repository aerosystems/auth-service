package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	Id           int
	Uuid         uuid.UUID
	Email        string
	PasswordHash string
	Role         KindRole
	IsActive     bool
	GoogleId     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
