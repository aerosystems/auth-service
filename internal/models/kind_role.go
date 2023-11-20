package models

type KindRole string

const (
	Customer KindRole = "customer"
	Staff    KindRole = "staff"
)

func (k KindRole) String() string {
	return string(k)
}
