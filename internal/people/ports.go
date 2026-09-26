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
	Refresh(ctx context.Context, tokenRing *core.TokenRing) (*core.AuthResponse, error)
}

type Repository interface {
	FindByID(ctx context.Context, personID uuid.UUID) (*Person, error)
	FindAll(ctx context.Context, pagination core.Pagination) ([]*Person, int, error)
	Register(ctx context.Context, person *Person) error
	FindCredentials(ctx context.Context, provider IdentityProvider, registrationID string) (uuid.UUID, string, error)
	SaveRoles(ctx context.Context, personID uuid.UUID, roles []Role) error
}
