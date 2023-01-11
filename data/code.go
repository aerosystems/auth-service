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

	query := `select id, code, user_id, created, expiration from codes where code = $1`

	var code Code
	row := db.QueryRowContext(ctx, query, XXXXXX)

	err := row.Scan(
		&code.ID,
		&code.Code,
		&code.UserID,
		&code.Created,
		&code.Expiration,
	)

	if err != nil {
		return nil, err
	}

	return &code, nil
}

func (c *Code) GetLastActiveCode(userID int) (*Code, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `SELECT id, code, user_id, created, expiration 
				FROM codes
				WHERE user_id = $1
				AND expiration > NOW()
				ORDER BY created DESC
				LIMIT 11`

	var code Code
	row := db.QueryRowContext(ctx, query, userID)

	err := row.Scan(
		&code.ID,
		&code.Code,
		&code.UserID,
		&code.Created,
		&code.Expiration,
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
	stmt := `insert into codes (code, user_id, created, expiration)
		values ($1, $2, $3, $4) returning id`

	err := db.QueryRowContext(ctx, stmt,
		&c.Code,
		&c.UserID,
		&c.Created,
		&c.Expiration,
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// CreateCode generation new code
func (c *Code) CreateCode(userID int) (int, error) {
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
	}

	_, err = code.Insert()
	if err != nil {
		return 0, err
	}

	return XXXXXX, nil
}
