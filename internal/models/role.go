package models

type KindRole struct {
	slug string
}

var (
	UnknownRole  = KindRole{"unknown"}
	CustomerRole = KindRole{"customer"}
	StaffRole    = KindRole{"staff"}
)

func (k KindRole) String() string {
	return k.slug
}

func RoleFromString(s string) KindRole {
	switch s {
	case "customer":
		return CustomerRole
	case "staff":
		return StaffRole
	default:
		return UnknownRole
	}
}
