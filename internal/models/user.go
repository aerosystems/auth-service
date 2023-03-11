package models

import "time"

type User struct {
	ID        int       `json:"id" gorm:"<-"`
	Email     string    `json:"email" gorm:"<-"`
	Password  string    `json:"-" gorm:"<-"`
	Role      string    `json:"role" gorm:"<-"`
	CreatedAt time.Time `json:"created_at" gorm:"<-"`
	UpdatedAt time.Time `json:"updated_at" gorm:"<-"`
	IsActive  bool      `json:"is_active" gorm:"<-"`
	GoogleID  string    `json:"google_id" gorm:"<-"`
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
