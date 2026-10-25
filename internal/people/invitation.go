package people

import (
	"strings"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type InvitationStatus string

const (
	InvitationStatusPending   InvitationStatus = "PENDING"
	InvitationStatusCompleted InvitationStatus = "COMPLETED"
)

type Invitation struct {
	id          core.InvitationID
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

func (i *Invitation) ID() core.InvitationID {
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

func copyPersonIDPtr(p *core.PersonID) *core.PersonID {
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}

func copyTimePtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	cp := *t
	return &cp
}

func (i *Invitation) CreatedBy() core.PersonID {
	return i.createdBy
}

func (i *Invitation) PersonID() *core.PersonID {
	return copyPersonIDPtr(i.personID)
}

func (i *Invitation) Status() InvitationStatus {
	return i.status
}

func (i *Invitation) CreatedAt() time.Time {
	return i.createdAt
}

func (i *Invitation) CompletedAt() *time.Time {
	return copyTimePtr(i.completedAt)
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
		return exceptions.NewValidationException("invitation is already completed", nil)
	}
	if i.IsExpired() {
		return exceptions.NewValidationException("invitation has expired", nil)
	}
	now := time.Now()
	i.status = InvitationStatusCompleted
	i.completedAt = &now
	return nil
}

func NewInvitation(
	id core.InvitationID,
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
		return nil, exceptions.NewValidationException("email is required", nil)
	}
	if codeHash == "" {
		return nil, exceptions.NewValidationException("code hash is required", nil)
	}
	if createdBy.IsNil() {
		return nil, exceptions.NewValidationException("created by is required", nil)
	}

	invID := id
	if invID.IsNil() {
		invID = core.NewInvitationID()
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
		personID:    copyPersonIDPtr(personID),
		status:      s,
		createdAt:   cAt,
		completedAt: copyTimePtr(completedAt),
	}, nil
}

func ReconstituteInvitation(
	id core.InvitationID,
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
		personID:    copyPersonIDPtr(personID),
		status:      status,
		createdAt:   createdAt,
		completedAt: copyTimePtr(completedAt),
	}
}
