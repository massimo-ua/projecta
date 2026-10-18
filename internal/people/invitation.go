package people

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type InvitationStatus string

const (
	InvitationStatusPending   InvitationStatus = "PENDING"
	InvitationStatusCompleted InvitationStatus = "COMPLETED"
)

type Invitation struct {
	id          uuid.UUID
	email       string
	codeHash    string
	expiresAt   time.Time
	createdBy   core.PersonID
	personID    *core.PersonID
	status      InvitationStatus
	createdAt   time.Time
	completedAt *time.Time
}

type InvitationWithCode struct {
	Invitation *Invitation
	Code       string
}

func (i *Invitation) ID() uuid.UUID {
	return i.id
}

func (i *Invitation) Email() string {
	return i.email
}

func (i *Invitation) CodeHash() string {
	return i.codeHash
}

func (i *Invitation) ExpiresAt() time.Time {
	return i.expiresAt
}

func (i *Invitation) CreatedBy() core.PersonID {
	return i.createdBy
}

func (i *Invitation) PersonID() *core.PersonID {
	return i.personID
}

func (i *Invitation) Status() InvitationStatus {
	return i.status
}

func (i *Invitation) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Invitation) CompletedAt() *time.Time {
	return i.completedAt
}

func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.expiresAt)
}

func (i *Invitation) IsPending() bool {
	return i.status == InvitationStatusPending && !i.IsExpired()
}

func (i *Invitation) IsCompleted() bool {
	return i.status == InvitationStatusCompleted
}

func (i *Invitation) ValidateForRedemption(email string) (core.PersonID, error) {
	if i.IsExpired() {
		return core.NilPersonID, exceptions.NewValidationException("invitation has expired", nil)
	}
	if i.IsCompleted() {
		return core.NilPersonID, exceptions.NewValidationException("invitation has already been completed", nil)
	}
	if email != "" && !strings.EqualFold(email, i.email) {
		return core.NilPersonID, exceptions.NewValidationException("google account email does not match invitation email", nil)
	}
	if i.personID == nil {
		return core.NilPersonID, exceptions.NewInternalException("invitation has no associated user record", nil)
	}
	return *i.personID, nil
}

func (i *Invitation) Complete() error {
	if i.status == InvitationStatusCompleted {
		return errors.New("invitation is already completed")
	}
	if i.IsExpired() {
		return errors.New("invitation has expired")
	}
	now := time.Now()
	i.status = InvitationStatusCompleted
	i.completedAt = &now
	return nil
}

func NewInvitation(
	id uuid.UUID,
	email string,
	codeHash string,
	expiresAt time.Time,
	createdBy core.PersonID,
	personID *core.PersonID,
	status InvitationStatus,
	createdAt time.Time,
	completedAt *time.Time,
) (*Invitation, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	if codeHash == "" {
		return nil, errors.New("code hash is required")
	}
	if createdBy.IsNil() {
		return nil, errors.New("created by is required")
	}

	invID := id
	if invID == uuid.Nil {
		invID = uuid.New()
	}

	cAt := createdAt
	if cAt.IsZero() {
		cAt = time.Now()
	}

	s := status
	if s == "" {
		s = InvitationStatusPending
	}

	return &Invitation{
		id:          invID,
		email:       email,
		codeHash:    codeHash,
		expiresAt:   expiresAt,
		createdBy:   createdBy,
		personID:    personID,
		status:      s,
		createdAt:   cAt,
		completedAt: completedAt,
	}, nil
}

func ReconstituteInvitation(
	id uuid.UUID,
	email string,
	codeHash string,
	expiresAt time.Time,
	createdBy core.PersonID,
	personID *core.PersonID,
	status InvitationStatus,
	createdAt time.Time,
	completedAt *time.Time,
) *Invitation {
	return &Invitation{
		id:          id,
		email:       email,
		codeHash:    codeHash,
		expiresAt:   expiresAt,
		createdBy:   createdBy,
		personID:    personID,
		status:      status,
		createdAt:   createdAt,
		completedAt: completedAt,
	}
}
