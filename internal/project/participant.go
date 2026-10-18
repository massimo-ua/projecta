package project

import (
	"strings"
	"unicode/utf8"

	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type Participant struct {
	displayName string
}

func NewParticipant(displayName string) (Participant, error) {
	trimmed := strings.TrimSpace(displayName)
	if trimmed == "" {
		return Participant{}, exceptions.NewValidationException("participant display name cannot be empty", nil)
	}
	if utf8.RuneCountInString(trimmed) > 255 {
		return Participant{}, exceptions.NewValidationException("participant display name cannot exceed 255 characters", nil)
	}
	return Participant{displayName: trimmed}, nil
}

func (p Participant) DisplayName() string {
	return p.displayName
}

func (p Participant) String() string {
	return p.displayName
}

func (p Participant) Equals(other Participant) bool {
	return p.displayName == other.displayName
}

