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
	google           core.ThirdPartyAuth
}

func NewAuthService(
	peopleRepository Repository,
	tokenProvider core.AuthTokenProvider,
	hasher core.Hasher,
	google core.ThirdPartyAuth,
) AuthService {
	return &AuthServiceImpl{
		peopleRepository: peopleRepository,
		tokenProvider:    tokenProvider,
		hasher:           hasher,
		google:           google,
	}
}

func (s *AuthServiceImpl) Login(ctx context.Context, credentials Credentials) (*core.AuthResponse, error) {
	switch credentials.Provider() {
	case LOCAL:
		return s.loginWithLocal(ctx, credentials)
	case GOOGLE:
		return s.loginWithGoogle(ctx, credentials.Identifier())
	default:
		return nil, errors.New("unsupported identity provider")
	}
}

func (s *AuthServiceImpl) loginWithLocal(ctx context.Context, credentials Credentials) (*core.AuthResponse, error) {
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

	return s.authorizePerson(ctx, personID)
}

func (s *AuthServiceImpl) authorizePerson(ctx context.Context, personID uuid.UUID) (*core.AuthResponse, error) {
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

func (s *AuthServiceImpl) loginWithGoogle(ctx context.Context, token string) (*core.AuthResponse, error) {
	claims, err := s.google.ValidateToken(token)

	if err != nil {
		return nil, errors.Join(loginFailedError, err)
	}

	personID, _, err := s.peopleRepository.FindCredentials(ctx, GOOGLE, claims.Sub)

	if err != nil {
		return nil, errors.Join(loginFailedError, err)
	}

	return s.authorizePerson(ctx, personID)
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
		Sub:         person.ID().String(),
		DisplayName: person.FullName(),
		Roles:       nil,
	})

	if err != nil {
		return nil, errors.Join(core.RefreshTokenIsInvalid, err)
	}

	return authResponse, nil
}
