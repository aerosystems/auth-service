package models

import (
	"time"
)

type Code struct {
	Id        int
	Code      string
	UserId    int
	User      User
	Action    KindCode
	Data      string
	IsUsed    bool
	ExpireAt  time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type KindCode struct {
	slug string
}

var (
	UnknownCode       = KindCode{"unknown"}
	RegistrationCode  = KindCode{"registration"}
	ResetPasswordCode = KindCode{"resetPassword"}
)

func (k KindCode) String() string {
	return k.slug
}

func CodeFromString(s string) KindCode {
	switch s {
	case "registration":
		return RegistrationCode
	case "resetPassword":
		return ResetPasswordCode
	default:
		return UnknownCode
	}
}
