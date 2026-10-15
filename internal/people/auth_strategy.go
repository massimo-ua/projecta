package people

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

// IdentityProviderStrategy defines the polymorphic contract for authenticating credentials.
type IdentityProviderStrategy interface {
	Provider() IdentityProvider
	Authenticate(ctx context.Context, credentials Credentials) (uuid.UUID, error)
}

// LocalAuthStrategy authenticates users against stored password hashes.
type LocalAuthStrategy struct {
	repo   Repository
	hasher core.Hasher
}

// NewLocalAuthStrategy creates a new instance of LocalAuthStrategy.
func NewLocalAuthStrategy(repo Repository, hasher core.Hasher) *LocalAuthStrategy {
	return &LocalAuthStrategy{
		repo:   repo,
		hasher: hasher,
	}
}

func (s *LocalAuthStrategy) Provider() IdentityProvider {
	return LOCAL
}

func (s *LocalAuthStrategy) Authenticate(ctx context.Context, credentials Credentials) (uuid.UUID, error) {
	personID, hash, err := s.repo.FindCredentials(
		ctx,
		credentials.Provider(),
		credentials.RegistrationID(),
	)
	if err != nil {
		return uuid.Nil, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	if !s.hasher.Compare(credentials.Identifier(), hash) {
		return uuid.Nil, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	return personID, nil
}

// GoogleAuthStrategy authenticates users by validating Google ID tokens.
type GoogleAuthStrategy struct {
	repo   Repository
	google core.ThirdPartyAuth
}

// NewGoogleAuthStrategy creates a new instance of GoogleAuthStrategy.
func NewGoogleAuthStrategy(repo Repository, google core.ThirdPartyAuth) *GoogleAuthStrategy {
	return &GoogleAuthStrategy{
		repo:   repo,
		google: google,
	}
}

func (s *GoogleAuthStrategy) Provider() IdentityProvider {
	return GOOGLE
}

func (s *GoogleAuthStrategy) Authenticate(ctx context.Context, credentials Credentials) (uuid.UUID, error) {
	claims, err := s.google.ValidateToken(credentials.Identifier())
	if err != nil {
		log.Printf("[GOOGLE AUTH ERROR] Failed to validate Google token: %v", err)
		return uuid.Nil, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	log.Printf("[GOOGLE AUTH] Successfully validated Google token for Sub ID: %s (DisplayName: %s)", claims.Sub, claims.DisplayName)

	personID, _, err := s.repo.FindCredentials(ctx, GOOGLE, claims.Sub)
	if err != nil {
		log.Printf("[GOOGLE AUTH ERROR] No user found for Google Sub ID: %s. To map this user, insert credentials with provider='GOOGLE' and registration_id='%s'", claims.Sub, claims.Sub)
		return uuid.Nil, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	return personID, nil
}
