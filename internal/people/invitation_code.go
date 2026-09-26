package people

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type InvitationCode struct {
	value string
}

func NewInvitationCode(code string) (InvitationCode, error) {
	clean := strings.ToLower(strings.TrimSpace(code))
	if clean == "" {
		return InvitationCode{}, exceptions.NewValidationException("invitation code is required", nil)
	}
	return InvitationCode{value: clean}, nil
}

func GenerateInvitationCode() (InvitationCode, error) {
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return InvitationCode{}, exceptions.NewInternalException("failed to generate invitation code", err)
	}
	return InvitationCode{value: hex.EncodeToString(rawBytes)}, nil
}

func (c InvitationCode) String() string {
	return c.value
}

func (c InvitationCode) Hash() string {
	if c.value == "" {
		return ""
	}
	h := sha256.Sum256([]byte(c.value))
	return hex.EncodeToString(h[:])
}

func (c InvitationCode) Equals(other InvitationCode) bool {
	return c.value == other.value
}
