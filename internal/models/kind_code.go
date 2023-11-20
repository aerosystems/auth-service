package models

type KindCode string

const (
	Registration  KindCode = "registration"
	ResetPassword KindCode = "resetPassword"
)

func (k KindCode) String() string {
	return string(k)
}
