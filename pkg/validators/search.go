package validators

import (
	"errors"
	"regexp"
)

func ValidateQuery(query string) error {
	if len(query) == 0 {
		return errors.New("username should be of 1 characters long")
	}
	if len(query) > 40 {
		return errors.New("username length should be maximum 40 characters long")
	}
	done, err := regexp.MatchString("([a-zA-Z0-9])+", query)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("username should contain only lower, upper case latin letters and digits")
	}
	return nil
}
