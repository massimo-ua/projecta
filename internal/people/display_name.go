package people

import (
	"strings"
	"unicode/utf8"

	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type DisplayName struct {
	value string
}

func NewDisplayName(value string) (DisplayName, error) {
	trimmed := strings.TrimSpace(value)
	if utf8.RuneCountInString(trimmed) > 255 {
		return DisplayName{}, exceptions.NewValidationException("display name cannot exceed 255 characters", nil)
	}
	return DisplayName{value: trimmed}, nil
}

func (d DisplayName) String() string {
	return d.value
}

func (d DisplayName) IsEmpty() bool {
	return d.value == ""
}

func (d DisplayName) Equals(other DisplayName) bool {
	return d.value == other.value
}

