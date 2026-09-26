package people

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"
)

type DisplayName struct {
	value string
}

func NewDisplayName(value string) (DisplayName, error) {
	trimmed := strings.TrimSpace(value)
	if utf8.RuneCountInString(trimmed) > 255 {
		return DisplayName{}, errors.New("display name cannot exceed 255 characters")
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

func (d *DisplayName) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	dn, err := NewDisplayName(raw)
	if err != nil {
		return err
	}
	*d = dn
	return nil
}

func (d DisplayName) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.value)
}
