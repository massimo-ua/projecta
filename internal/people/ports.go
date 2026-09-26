package people

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type UserService interface {
	Register(ctx context.Context, command RegisterCommand) error
	FindByID(ctx context.Context, personID uuid.UUID) (*Person, error)
	FindAll(ctx context.Context, pagination core.Pagination) ([]*Person, int, error)
	AssignRoles(ctx context.Context, command AssignRolesCommand) error
}

type AuthService interface {
	Login(ctx context.Context, credentials Credentials) (*core.AuthResponse, error)
	LoginWithInvitation(ctx context.Context, token string, provider IdentityProvider, invitationCode string) (*core.AuthResponse, error)
	Refresh(ctx context.Context, tokenRing *core.TokenRing) (*core.AuthResponse, error)
}

type Repository interface {
	FindByID(ctx context.Context, personID uuid.UUID) (*Person, error)
	FindAll(ctx context.Context, pagination core.Pagination) ([]*Person, int, error)
	Register(ctx context.Context, person *Person) error
	FindCredentials(ctx context.Context, provider IdentityProvider, registrationID string) (uuid.UUID, string, error)
	SaveRoles(ctx context.Context, personID uuid.UUID, roles []Role) error
	DeletePerson(ctx context.Context, personID uuid.UUID) error
	UpdateProfile(ctx context.Context, personID uuid.UUID, firstName string, lastName string, displayName string) error
	SaveCredentials(ctx context.Context, personID uuid.UUID, cred Credentials) error
}

type InvitationRepository interface {
	Create(ctx context.Context, invitation *Invitation) error
	FindByID(ctx context.Context, id uuid.UUID) (*Invitation, error)
	FindByCodeHash(ctx context.Context, codeHash string) (*Invitation, error)
	FindByCreator(ctx context.Context, creatorID uuid.UUID) ([]*Invitation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Complete(ctx context.Context, id uuid.UUID) error
	EmailExists(ctx context.Context, email string) (bool, error)
}

type InvitationService interface {
	Create(ctx context.Context, adminID uuid.UUID, email string) (*InvitationWithCode, error)
	FindAllByCreator(ctx context.Context, adminID uuid.UUID) ([]*Invitation, error)
	Delete(ctx context.Context, adminID uuid.UUID, invitationID uuid.UUID) error
	FindByCode(ctx context.Context, code string) (*Invitation, error)
}
