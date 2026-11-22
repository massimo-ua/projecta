package people

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

const maxRefreshTokenAge = 7 * 24 * time.Hour

type AuthServiceImpl struct {
	db               core.DbConnection
	peopleRepository Repository
	tokenProvider    core.AuthTokenProvider
	hasher           core.Hasher
	invitationRepo   InvitationRepository
	strategies       map[IdentityProvider]IdentityProviderStrategy
}

func NewAuthService(
	db core.DbConnection,
	peopleRepository Repository,
	tokenProvider core.AuthTokenProvider,
	hasher core.Hasher,
	google core.ThirdPartyAuth,
	invitationRepo InvitationRepository,
	strategies ...IdentityProviderStrategy,
) AuthService {
	strategyMap := make(map[IdentityProvider]IdentityProviderStrategy)
	if len(strategies) > 0 {
		for _, strat := range strategies {
			strategyMap[strat.Provider()] = strat
		}
	} else {
		strategyMap[LOCAL] = NewLocalAuthStrategy(peopleRepository, hasher)
		if google != nil {
			strategyMap[GOOGLE] = NewGoogleAuthStrategy(peopleRepository, google)
		}
	}

	return &AuthServiceImpl{
		db:               db,
		peopleRepository: peopleRepository,
		tokenProvider:    tokenProvider,
		hasher:           hasher,
		invitationRepo:   invitationRepo,
		strategies:       strategyMap,
	}
}

func (s *AuthServiceImpl) Login(ctx context.Context, credentials Credentials) (*core.AuthResponse, error) {
	strategy, ok := s.strategies[credentials.Provider()]
	if !ok {
		return nil, exceptions.NewValidationException("unsupported identity provider", nil)
	}

	personID, err := strategy.Authenticate(ctx, credentials)
	if err != nil {
		return nil, err
	}

	return s.authorizePerson(ctx, personID, "login failed", loginFailedError)
}

func (s *AuthServiceImpl) LoginWithInvitation(
	ctx context.Context,
	token string,
	provider IdentityProvider,
	invitationCode string,
) (*core.AuthResponse, error) {
	invCode, err := NewInvitationCode(invitationCode)
	if err != nil {
		return nil, err
	}

	strategy, ok := s.strategies[provider]
	if !ok {
		return nil, exceptions.NewValidationException("unsupported identity provider for invitation", nil)
	}

	claims, err := strategy.ResolveInvitationClaims(ctx, token)
	if err != nil {
		return nil, err
	}

	if s.invitationRepo == nil {
		return nil, exceptions.NewInternalException("invitation repository not configured", nil)
	}

	redeemFn := func(txCtx context.Context) (any, error) {
		invitation, err := s.invitationRepo.FindByCodeHash(txCtx, invCode.Hash())
		if err != nil {
			return nil, exceptions.NewNotFoundException("invitation not found", err)
		}

		personID, err := invitation.ValidateForRedemption(claims.Email)
		if err != nil {
			return nil, err
		}

		person, err := s.peopleRepository.FindByID(txCtx, personID)
		if err != nil {
			return nil, exceptions.NewInternalException("failed to find invited user", err)
		}

		if err := person.AcceptInvitation(claims, provider); err != nil {
			return nil, err
		}

		if err := invitation.Complete(); err != nil {
			return nil, exceptions.NewValidationException(err.Error(), err)
		}

		if err := s.peopleRepository.Save(txCtx, person); err != nil {
			return nil, exceptions.NewInternalException("failed to update invited user", err)
		}

		if err := s.invitationRepo.Save(txCtx, invitation); err != nil {
			return nil, exceptions.NewInternalException("failed to update invitation", err)
		}

		return person, nil
	}

	var result any
	if s.db != nil {
		result, err = s.db.Tx(ctx, redeemFn)
	} else {
		result, err = redeemFn(ctx)
	}
	if err != nil {
		return nil, err
	}

	person, ok := result.(*Person)
	if !ok || person == nil {
		return nil, exceptions.NewInternalException("failed to redeem invitation", nil)
	}

	return s.authorizePerson(ctx, person.ID(), "login failed", loginFailedError)
}

func (s *AuthServiceImpl) issueTokens(person *Person) (*core.AuthResponse, error) {
	authResponse, err := s.tokenProvider.GenerateTokenRing(core.AuthTokenPayload{
		Sub:         person.ID().String(),
		DisplayName: person.DisplayName(),
		Roles:       person.RoleStrings(),
	})
	if err != nil {
		return nil, exceptions.NewInternalException("failed to generate tokens", errors.Join(core.AuthTokenGenerationFailed, err))
	}

	return authResponse, nil
}

func (s *AuthServiceImpl) authorizePerson(
	ctx context.Context,
	personID core.PersonID,
	unauthorizedMsg string,
	unauthorizedErr error,
) (*core.AuthResponse, error) {
	person, err := s.peopleRepository.FindByID(ctx, personID)
	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return nil, exceptions.NewUnauthorizedException(unauthorizedMsg, errors.Join(unauthorizedErr, err))
		}
		return nil, exceptions.NewInternalException("failed to load user", err)
	}
	if person == nil {
		return nil, exceptions.NewUnauthorizedException(unauthorizedMsg, unauthorizedErr)
	}

	return s.issueTokens(person)
}

func (s *AuthServiceImpl) decodeRefreshClaims(tokenRing *core.TokenRing) (*core.AuthTokenClaims, error) {
	if tokenRing == nil {
		return nil, exceptions.NewUnauthorizedException("refresh token is invalid", core.RefreshTokenIsInvalid)
	}

	claims, err := s.tokenProvider.DecodeToken(tokenRing.AccessToken())
	if err != nil || claims == nil {
		return nil, exceptions.NewUnauthorizedException("refresh token is invalid", errors.Join(core.RefreshTokenIsInvalid, err))
	}

	if claims.IssuedAt <= 0 {
		return nil, exceptions.NewUnauthorizedException("refresh token is invalid", core.RefreshTokenIsInvalid)
	}

	if time.Since(time.Unix(claims.IssuedAt, 0)) > maxRefreshTokenAge {
		return nil, exceptions.NewUnauthorizedException("refresh token has expired", core.RefreshTokenIsInvalid)
	}

	return claims, nil
}

func (s *AuthServiceImpl) validateRefreshToken(tokenRing *core.TokenRing) (core.PersonID, error) {
	claims, err := s.decodeRefreshClaims(tokenRing)
	if err != nil {
		return core.NilPersonID, err
	}

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil || tokenID == uuid.Nil {
		return core.NilPersonID, exceptions.NewUnauthorizedException("refresh token is invalid", errors.Join(core.RefreshTokenIsInvalid, err))
	}

	if !s.tokenProvider.ValidateRefreshToken(tokenID, tokenRing.RefreshToken()) {
		return core.NilPersonID, exceptions.NewUnauthorizedException("refresh token is invalid", core.RefreshTokenIsInvalid)
	}

	personID, err := core.ParsePersonID(claims.Sub)
	if err != nil || personID.IsNil() {
		return core.NilPersonID, exceptions.NewUnauthorizedException("refresh token is invalid", errors.Join(core.RefreshTokenIsInvalid, err))
	}

	return personID, nil
}

func (s *AuthServiceImpl) Refresh(ctx context.Context, tokenRing *core.TokenRing) (*core.AuthResponse, error) {
	personID, err := s.validateRefreshToken(tokenRing)
	if err != nil {
		return nil, err
	}

	return s.authorizePerson(ctx, personID, "refresh token is invalid", core.RefreshTokenIsInvalid)
}
