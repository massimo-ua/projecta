package people

import (
    "context"
    "github.com/google/uuid"
    "gitlab.com/massimo-ua/projecta/internal/core"
)

type UserService interface {
    Register(ctx context.Context, command RegisterCommand) error
    FindByID(ctx context.Context, personID uuid.UUID) (*Person, error)
}

type AuthService interface {
    Login(ctx context.Context, credentials Credentials) (*core.AuthResponse, error)
    Refresh(ctx context.Context, tokenRing *core.TokenRing) (*core.AuthResponse, error)
}

type Repository interface {
    core.BaseRepository
    FindByID(ctx context.Context, personID uuid.UUID) (*Person, error)
    Register(ctx context.Context, person *Person) error
    FindCredentials(ctx context.Context, provider IdentityProvider, registrationID string) (uuid.UUID, string, error)
}
