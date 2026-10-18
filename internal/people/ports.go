package people

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type UserService interface {
	Register(ctx context.Context, command RegisterCommand) error
	FindByID(ctx context.Context, personID core.PersonID) (*Person, error)
	FindAll(ctx context.Context, pagination core.Pagination) ([]*Person, int, error)
	AssignRoles(ctx context.Context, command AssignRolesCommand) error
	UpdateDisplayName(ctx context.Context, command UpdateDisplayNameCommand) (*Person, error)
}

type AuthService interface {
	Login(ctx context.Context, credentials Credentials) (*core.AuthResponse, error)
	LoginWithInvitation(ctx context.Context, token string, provider IdentityProvider, invitationCode string) (*core.AuthResponse, error)
	Refresh(ctx context.Context, tokenRing *core.TokenRing) (*core.AuthResponse, error)
}

type Repository interface {
	FindByID(ctx context.Context, personID core.PersonID) (*Person, error)
	FindAll(ctx context.Context, pagination core.Pagination) ([]*Person, int, error)
	Save(ctx context.Context, person *Person) error
	FindCredentials(ctx context.Context, provider IdentityProvider, registrationID string) (core.PersonID, string, error)
	DeletePerson(ctx context.Context, personID core.PersonID) error
}

type InvitationRepository interface {
	Save(ctx context.Context, invitation *Invitation) error
	FindByID(ctx context.Context, id uuid.UUID) (*Invitation, error)
	FindByCodeHash(ctx context.Context, codeHash string) (*Invitation, error)
	FindByCreator(ctx context.Context, creatorID core.PersonID) ([]*Invitation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	EmailExists(ctx context.Context, email string) (bool, error)
}

type InvitationService interface {
	Create(ctx context.Context, adminID core.PersonID, email string) (*InvitationWithCode, error)
	FindAllByCreator(ctx context.Context, adminID core.PersonID) ([]*Invitation, error)
	Delete(ctx context.Context, adminID core.PersonID, invitationID uuid.UUID) error
	FindByCode(ctx context.Context, code string) (*Invitation, error)
}
