package validators

import (
	"errors"
)

func ValidateCode(code int) error {
	count := 0
	for code > 0 {
		code = code / 10
		count++
		if count > 6 {
			return errors.New("code must contain 6 digits")
		}
	}
	if count != 6 {
		return errors.New("code must contain 6 digits")
	}
	return nil
}
