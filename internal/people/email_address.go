package people

import (
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"net/mail"
)

type EmailAddress struct {
	address string
}

func (e EmailAddress) String() string {
	return e.address
}

func (e EmailAddress) Equals(other EmailAddress) bool {
	return e.address == other.address
}

func NewEmailAddress(address string) (EmailAddress, error) {
	email, err := mail.ParseAddress(address)
	if err != nil {
		return EmailAddress{}, exceptions.NewValidationException("invalid email address", nil)
	}
	return EmailAddress{address: email.Address}, nil
}
