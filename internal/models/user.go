package models

import "time"

type User struct {
	ID        int       `json:"id" gorm:"primaryKey;unique;autoIncrement"`
	Email     string    `json:"email" gorm:"unique"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	IsActive  bool      `json:"-"`
	GoogleID  string    `json:"-"`
}

type UserRepository interface {
	FindAll() (*[]User, error)
	FindByID(ID int) (*User, error)
	FindByEmail(Email string) (*User, error)
	FindByGoogleID(GoogleID string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(user *User) error
	ResetPassword(user *User, password string) error
	PasswordMatches(user *User, plainText string) (bool, error)
}
