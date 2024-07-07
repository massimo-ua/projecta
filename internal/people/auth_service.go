package people

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type AuthServiceImpl struct {
	peopleRepository Repository
	tokenProvider    core.AuthTokenProvider
	hasher           core.Hasher
}

func NewAuthService(
	peopleRepository Repository,
	tokenProvider core.AuthTokenProvider,
	hasher core.Hasher,
) AuthService {
	return &AuthServiceImpl{
		peopleRepository: peopleRepository,
		tokenProvider:    tokenProvider,
		hasher:           hasher,
	}
}

func (s *AuthServiceImpl) Login(ctx context.Context, credentials Credentials) (*core.AuthResponse, error) {
	personID, hash, err := s.peopleRepository.FindCredentials(
		ctx,
		credentials.Provider(),
		credentials.RegistrationID())

	if err != nil {
		return nil, errors.Join(loginFailedError, err)
	}

	if !s.hasher.Compare(credentials.Identifier(), hash) {
		return nil, errors.Join(loginFailedError, err)
	}

	customer, err := s.peopleRepository.FindByID(ctx, personID)

	if err != nil {
		return nil, errors.Join(loginFailedError, err)
	}

	authResponse, err := s.tokenProvider.GenerateTokenRing(core.AuthTokenPayload{
		Sub:         personID.String(),
		DisplayName: customer.FullName(),
		Roles:       nil,
	})

	if err != nil {
		return nil, errors.Join(loginFailedError, err)
	}

	return authResponse, nil
}

func (s *AuthServiceImpl) Refresh(ctx context.Context, tokenRing *core.TokenRing) (*core.AuthResponse, error) {
	claims, err := s.tokenProvider.DecodeToken(tokenRing.AccessToken())

	if err != nil {
		return nil, errors.Join(core.RefreshTokenIsInvalid, err)
	}

	tokenID, err := uuid.Parse(claims.ID)

	if err != nil {
		return nil, errors.Join(core.RefreshTokenIsInvalid, err)
	}

	if ok := s.tokenProvider.ValidateRefreshToken(tokenID, tokenRing.RefreshToken()); !ok {
		return nil, errors.Join(core.RefreshTokenIsInvalid, err)
	}

	personID, err := uuid.Parse(claims.Sub)

	if err != nil {
		return nil, errors.Join(core.RefreshTokenIsInvalid, err)
	}

	person, err := s.peopleRepository.FindByID(ctx, personID)

	if err != nil {
		return nil, errors.Join(core.RefreshTokenIsInvalid, err)
	}

	authResponse, err := s.tokenProvider.GenerateTokenRing(core.AuthTokenPayload{
		Sub:         person.ID.String(),
		DisplayName: person.FullName(),
		Roles:       nil,
	})

	if err != nil {
		return nil, errors.Join(core.RefreshTokenIsInvalid, err)
	}

	return authResponse, nil
}
