package data

import (
	"database/sql"
	"time"
)

const dbTimeout = time.Second * 3

var db *sql.DB

// New is the function used to create an instance of the data package. It returns the type
// Model, which embeds all the types we want to be available to our application.
func New(dbPool *sql.DB) Models {
	db = dbPool

	return Models{
		User: User{},
		Code: Code{},
	}
}

// Models is the type for this package. Note that any model that is included as a member
// in this type is available to us throughout the application, anywhere that the
// app variable is used, provided that the model is also added in the New function.
type Models struct {
	User User
	Code Code
}

// User is the structure which holds one user from the database.
type User struct {
	ID       int       `json:"id"`
	Email    string    `json:"email"`
	Password string    `json:"-"`
	Role     string    `json:"role"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
	Active   bool      `json:"active"`
}

type Code struct {
	ID         int       `json:"id"`
	Code       int       `json:"code"`
	UserID     int       `json:"user_id"`
	Created    time.Time `json:"created"`
	Expiration time.Time `json:"expiration"`
	Action     string    `json:"action"`
	Data       string    `json:"data"`
}
