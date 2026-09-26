package projecta

import (
	"encoding/json"
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

func (p Participant) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.displayName)
}

func (p *Participant) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	participant, err := NewParticipant(raw)
	if err != nil {
		return err
	}
	*p = participant
	return nil
}
