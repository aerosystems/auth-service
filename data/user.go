package data

import (
	"context"
	"errors"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// GetAll returns a slice of all users, sorted by last name
func (u *User) GetAll() ([]*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, email, password, role, created, updated, active
				FROM users
				ORDER BY last_name`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.Created,
			&user.Updated,
			&user.Active,
		)
		if err != nil {
			log.Println("Error scanning", err)
			return nil, err
		}

		users = append(users, &user)
	}

	return users, nil
}

// GetByEmail returns one user by email
func (u *User) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, email, password, role, created, updated, active
				FROM users
				WHERE email = $1`

	var user User
	row := db.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Created,
		&user.Updated,
		&user.Active,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetOne returns one user by id
func (u *User) GetOne(id int) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, email, password, role, created, updated, active
				FROM users
				WHERE id = $1`

	var user User
	row := db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Created,
		&user.Updated,
		&user.Active,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update updates one user in the database, using the information
// stored in the receiver u
func (u *User) Update() error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `UPDATE users SET
		email = $1,
		role = $2,
		updated = $3,
		active = $4
		WHERE id = $5
	`

	_, err := db.ExecContext(ctx, stmt,
		u.Email,
		u.Role,
		time.Now(),
		u.Active,
		u.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// Delete deletes one user from the database, by User.ID
func (u *User) Delete() error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM users
				WHERE id = $1`

	_, err := db.ExecContext(ctx, stmt, u.ID)
	if err != nil {
		return err
	}

	return nil
}

// DeleteByID deletes one user from the database, by ID
func (u *User) DeleteByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `DELETE FROM users
				WHERE id = $1`

	_, err := db.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	return nil
}

// Insert inserts a new user into the database, and returns the ID of the newly inserted row
func (u *User) Insert(user User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return 0, err
	}

	var newID int
	stmt := `INSERT INTO users (email, password, role, created, updated, active)
				VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err = db.QueryRowContext(ctx, stmt,
		user.Email,
		hashedPassword,
		user.Role,
		time.Now(),
		time.Now(),
		user.Active,
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// ResetPassword is the method we will use to change a user's password.
func (u *User) ResetPassword(password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `UPDATE users
				SET password = $1
				WHERE id = $2`
	_, err = db.ExecContext(ctx, stmt, hashedPassword, u.ID)
	if err != nil {
		return err
	}

	return nil
}

// PasswordMatches uses Go's bcrypt package to compare a user supplied password
// with the hash we have stored for a given user in the database. If the password
// and hash match, we return true; otherwise, we return false.
func (u *User) PasswordMatches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			// invalid password
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}
