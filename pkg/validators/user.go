package validators

import (
	"errors"
	"regexp"
)

func ValidateUsername(username string) error {
	if len(username) < 4 {
		return errors.New("username should be of 4 characters long")
	}
	if len(username) > 40 {
		return errors.New("username length should be maximum 40 characters long")
	}
	done, err := regexp.MatchString("([a-zA-Z0-9])+", username)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("username should contain only lower, upper case latin letters and digits")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password should be of 8 characters long")
	}
	if len(password) > 40 {
		return errors.New("password length should be maximum 40 characters long")
	}
	done, err := regexp.MatchString("([a-z])+", password)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("password should contain at least one lower case character")
	}
	done, err = regexp.MatchString("([A-Z])+", password)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("password should contain at least one upper case character")
	}
	done, err = regexp.MatchString("([0-9])+", password)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("password should contain at least one digit")
	}

	done, err = regexp.MatchString("([!@#$%^&*.?-])+", password)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("password should contain at least one special character")
	}
	return nil
}
