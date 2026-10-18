package people

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

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
		return nil, errors.New("unsupported identity provider")
	}

	personID, err := strategy.Authenticate(ctx, credentials)
	if err != nil {
		return nil, err
	}

	return s.authorizePerson(ctx, personID)
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

	invitation, err := s.invitationRepo.FindByCodeHash(ctx, invCode.Hash())
	if err != nil {
		return nil, exceptions.NewNotFoundException("invitation not found", err)
	}

	personID, err := invitation.ValidateForRedemption(claims.Email)
	if err != nil {
		return nil, err
	}

	person, err := s.peopleRepository.FindByID(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to find invited user", err)
	}

	if err := person.AcceptInvitation(claims, provider); err != nil {
		return nil, err
	}

	if err := invitation.Complete(); err != nil {
		return nil, exceptions.NewValidationException(err.Error(), err)
	}

	updateFn := func(txCtx context.Context) (any, error) {
		if err := s.peopleRepository.Save(txCtx, person); err != nil {
			return nil, err
		}

		if err := s.invitationRepo.Save(txCtx, invitation); err != nil {
			return nil, err
		}

		return nil, nil
	}

	if s.db != nil {
		if _, err := s.db.Tx(ctx, updateFn); err != nil {
			return nil, exceptions.NewInternalException("failed to update invited user", err)
		}
	} else {
		if _, err := updateFn(ctx); err != nil {
			return nil, exceptions.NewInternalException("failed to update invited user", err)
		}
	}

	return s.authorizePerson(ctx, person.ID())
}

func (s *AuthServiceImpl) authorizePerson(ctx context.Context, personID core.PersonID) (*core.AuthResponse, error) {
	customer, err := s.peopleRepository.FindByID(ctx, personID)

	if err != nil {
		return nil, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	authResponse, err := s.tokenProvider.GenerateTokenRing(core.AuthTokenPayload{
		Sub:         personID.String(),
		DisplayName: customer.FullName(),
		Roles:       customer.RoleStrings(),
	})

	if err != nil {
		return nil, exceptions.NewInternalException("failed to generate tokens", errors.Join(loginFailedError, err))
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

	personID, err := core.ParsePersonID(claims.Sub)

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
		Roles:       person.RoleStrings(),
	})

	if err != nil {
		return nil, errors.Join(core.RefreshTokenIsInvalid, err)
	}

	return authResponse, nil
}
