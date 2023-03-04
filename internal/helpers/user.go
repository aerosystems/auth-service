package helpers

import (
	"errors"
	"net/mail"
	"os"
	"regexp"
	"strings"
)

func ValidateEmail(data string) (string, error) {
	email, err := mail.ParseAddress(data)
	if err != nil {
		return "", err
	}

	return email.Address, nil
}

func NormalizeEmail(data string) string {
	addr := strings.ToLower(data)

	arrAddr := strings.Split(addr, "@")
	username := arrAddr[0]
	domain := arrAddr[1]

	googleDomains := strings.Split(os.Getenv("GOOGLEMAIL_DOMAINS"), ",")

	//checking google mail aliases
	if Contains(googleDomains, domain) {
		//removing all dots from username mail
		username = strings.ReplaceAll(username, ".", "")
		//removing all characters after +
		if strings.Contains(username, "+") {
			res := strings.Split(username, "+")
			username = res[0]
		}
		addr = username + "@gmail.com"
	}

	return addr
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password should be of 8 characters long")
	}
	done, err := regexp.MatchString("([a-z])+", password)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("password should contain atleast one lower case character")
	}
	done, err = regexp.MatchString("([A-Z])+", password)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("password should contain atleast one upper case character")
	}
	done, err = regexp.MatchString("([0-9])+", password)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("password should contain atleast one digit")
	}

	done, err = regexp.MatchString("([!@#$%^&*.?-])+", password)
	if err != nil {
		return err
	}
	if !done {
		return errors.New("password should contain atleast one special character")
	}
	return nil
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
