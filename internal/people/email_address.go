package people

import (
	"errors"
	"net/mail"
)

type EmailAddress = string

func NewEmailAddress(emailAddress string) (EmailAddress, error) {
	email, err := mail.ParseAddress(emailAddress)

	if err != nil {
		return "", errors.New("invalid email address")
	}

	return email.String(), nil
}
