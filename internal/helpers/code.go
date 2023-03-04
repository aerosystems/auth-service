package helpers

import (
	"errors"
	"math/rand"
	"time"
)

func GenCode() int {
	rand.Seed(time.Now().UnixNano())
	min := 100000
	max := 999999
	return rand.Intn(max-min+1) + min
}

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
