package data

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// GetByCode returns one code by code
func (c *Code) GetByCode(XXXXXX int) (*Code, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, code, user_id, created, expiration, action, data
				FROM codes
				WHERE code = $1`

	var code Code
	row := db.QueryRowContext(ctx, query, XXXXXX)

	err := row.Scan(
		&code.ID,
		&code.Code,
		&code.UserID,
		&code.Created,
		&code.Expiration,
		&code.Action,
		&code.Data,
	)

	if err != nil {
		return nil, err
	}

	return &code, nil
}

func (c *Code) GetLastActiveCode(userID int, action string) (*Code, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, code, user_id, created, expiration, action, data 
				FROM codes
				WHERE user_id = $1
				AND action = $2
				AND expiration > NOW()
				ORDER BY created DESC
				LIMIT 11`

	var code Code
	row := db.QueryRowContext(ctx, query, userID, action)

	err := row.Scan(
		&code.ID,
		&code.Code,
		&code.UserID,
		&code.Created,
		&code.Expiration,
		&code.Action,
		&code.Data,
	)

	if err != nil {
		return nil, err
	}

	return &code, nil
}

func (c *Code) ExtendExpiration() error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	codeExpMinutes, err := strconv.Atoi(os.Getenv("CODE_EXP_MINUTES"))
	if err != nil {
		return err
	}

	stmt := `UPDATE codes
				SET expiration = $1
				WHERE id = $2`

	_, err = db.ExecContext(ctx, stmt,
		codeExpMinutes,
		c.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// Insert inserts a new code into the database, and returns the ID of the newly inserted row
func (c *Code) Insert() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var newID int
	stmt := `INSERT INTO codes (code, user_id, created, expiration, action, data)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err := db.QueryRowContext(ctx, stmt,
		&c.Code,
		&c.UserID,
		&c.Created,
		&c.Expiration,
		&c.Action,
		&c.Data,
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// CreateCode generation new code
func (c *Code) CreateCode(userID int, action string, data string) (int, error) {
	codeExpMinutes, err := strconv.Atoi(os.Getenv("CODE_EXP_MINUTES"))
	if err != nil {
		return 0, err
	}

	rand.Seed(time.Now().UnixNano())
	min := 100000
	max := 999999
	XXXXXX := rand.Intn(max-min+1) + min
	code := Code{
		Code:       XXXXXX,
		UserID:     userID,
		Created:    time.Now(),
		Expiration: time.Now().Add(time.Minute * time.Duration(codeExpMinutes)),
		Action:     action,
		Data:       data,
	}

	_, err = code.Insert()
	if err != nil {
		return 0, err
	}

	return XXXXXX, nil
}
