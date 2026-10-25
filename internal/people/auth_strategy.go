package people

import (
	"context"
	"errors"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

// CredentialsReader defines the narrow read contract required by authentication strategies.
type CredentialsReader interface {
	FindCredentials(ctx context.Context, provider IdentityProvider, registrationID string) (core.PersonID, string, error)
}

// IdentityProviderStrategy defines the polymorphic contract for authenticating credentials.
type IdentityProviderStrategy interface {
	Provider() IdentityProvider
	Authenticate(ctx context.Context, credentials Credentials) (core.PersonID, error)
	ResolveInvitationClaims(ctx context.Context, token string) (*core.AuthTokenClaims, error)
}

// LocalAuthStrategy authenticates users against stored password hashes.
type LocalAuthStrategy struct {
	repo   CredentialsReader
	hasher core.Hasher
}

// NewLocalAuthStrategy creates a new instance of LocalAuthStrategy.
func NewLocalAuthStrategy(repo CredentialsReader, hasher core.Hasher) *LocalAuthStrategy {
	return &LocalAuthStrategy{
		repo:   repo,
		hasher: hasher,
	}
}

func (s *LocalAuthStrategy) Provider() IdentityProvider {
	return LOCAL
}

func (s *LocalAuthStrategy) Authenticate(ctx context.Context, credentials Credentials) (core.PersonID, error) {
	personID, hash, err := s.repo.FindCredentials(
		ctx,
		credentials.Provider(),
		credentials.RegistrationID(),
	)
	if err != nil {
		return core.NilPersonID, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	if !s.hasher.Compare(credentials.Identifier(), hash) {
		return core.NilPersonID, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	return personID, nil
}

func (s *LocalAuthStrategy) ResolveInvitationClaims(ctx context.Context, token string) (*core.AuthTokenClaims, error) {
	return nil, exceptions.NewValidationException("unsupported identity provider for invitation", nil)
}

// GoogleAuthStrategy authenticates users by validating Google ID tokens.
type GoogleAuthStrategy struct {
	repo   CredentialsReader
	google core.ThirdPartyAuth
}

// NewGoogleAuthStrategy creates a new instance of GoogleAuthStrategy.
func NewGoogleAuthStrategy(repo CredentialsReader, google core.ThirdPartyAuth) *GoogleAuthStrategy {
	return &GoogleAuthStrategy{
		repo:   repo,
		google: google,
	}
}

func (s *GoogleAuthStrategy) Provider() IdentityProvider {
	return GOOGLE
}

func (s *GoogleAuthStrategy) Authenticate(ctx context.Context, credentials Credentials) (core.PersonID, error) {
	claims, err := s.google.ValidateToken(credentials.Identifier())
	if err != nil {
		return core.NilPersonID, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	personID, _, err := s.repo.FindCredentials(ctx, GOOGLE, claims.Sub)
	if err != nil {
		return core.NilPersonID, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	return personID, nil
}

func (s *GoogleAuthStrategy) ResolveInvitationClaims(ctx context.Context, token string) (*core.AuthTokenClaims, error) {
	claims, err := s.google.ValidateToken(token)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}
	return claims, nil
}
