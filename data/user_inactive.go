package data

import (
	"context"
	"errors"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// GetAll returns a slice of all users, sorted by last name
func (u *UserInactive) GetAll() ([]*UserInactive, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, password, role, created_at, updated_at
	from users_inactive order by last_name`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*UserInactive

	for rows.Next() {
		var user UserInactive
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
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
func (u *UserInactive) GetByEmail(email string) (*UserInactive, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email, password, role, created_at, updated_at from users_inactive where email = $1`

	var user UserInactive
	row := db.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetOne returns one user by id
func (u *UserInactive) GetOne(id int) (*UserInactive, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `select id, email,password, role, created_at, updated_at from users_inactive where id = $1`

	var user UserInactive
	row := db.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update updates one user in the database, using the information
// stored in the receiver u
func (u *UserInactive) Update() error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `update users_inactive set email = $1, role = $2, updated_at = $3 where id = $4`

	_, err := db.ExecContext(ctx, stmt,
		u.Email,
		u.Role,
		time.Now(),
		u.ID,
	)

	if err != nil {

		return errors.New(stmt)
	}

	return nil
}

// ResetPassword is the method we will use to change a user's password.
func (u *UserInactive) ResetPassword(password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `update users_inactive set password = $1 where id = $2`
	_, err = db.ExecContext(ctx, stmt, hashedPassword, u.ID)
	if err != nil {
		return err
	}

	return nil
}

// Insert inserts a new user into the database, and returns the ID of the newly inserted row
func (u *UserInactive) Insert(user User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return 0, err
	}

	var newID int
	stmt := `insert into users_inactive (email, password, role, created_at, updated_at)
		values ($1, $2, $3, $4, $5) returning id`

	err = db.QueryRowContext(ctx, stmt,
		user.Email,
		hashedPassword,
		user.Role,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}
