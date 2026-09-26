package people

import (
	"errors"
	"time"

	"github.com/google/uuid"
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
	createdBy   uuid.UUID
	personID    *uuid.UUID
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

func (i *Invitation) CreatedBy() uuid.UUID {
	return i.createdBy
}

func (i *Invitation) PersonID() *uuid.UUID {
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
	createdBy uuid.UUID,
	personID *uuid.UUID,
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
	if createdBy == uuid.Nil {
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
