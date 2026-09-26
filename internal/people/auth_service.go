package people

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type AuthServiceImpl struct {
	db               core.DbConnection
	peopleRepository Repository
	tokenProvider    core.AuthTokenProvider
	hasher           core.Hasher
	google           core.ThirdPartyAuth
	invitationRepo   InvitationRepository
}

func NewAuthService(
	db core.DbConnection,
	peopleRepository Repository,
	tokenProvider core.AuthTokenProvider,
	hasher core.Hasher,
	google core.ThirdPartyAuth,
	invitationRepo InvitationRepository,
) AuthService {
	return &AuthServiceImpl{
		db:               db,
		peopleRepository: peopleRepository,
		tokenProvider:    tokenProvider,
		hasher:           hasher,
		google:           google,
		invitationRepo:   invitationRepo,
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

func (s *AuthServiceImpl) LoginWithInvitation(
	ctx context.Context,
	token string,
	provider IdentityProvider,
	invitationCode string,
) (*core.AuthResponse, error) {
	if provider != GOOGLE {
		return nil, exceptions.NewValidationException("unsupported identity provider for invitation", nil)
	}
	if invitationCode == "" {
		return nil, exceptions.NewValidationException("invitation code is required", nil)
	}

	claims, err := s.google.ValidateToken(token)
	if err != nil {
		log.Printf("[GOOGLE AUTH ERROR] Failed to validate Google token: %v", err)
		return nil, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	if s.invitationRepo == nil {
		return nil, exceptions.NewInternalException("invitation repository not configured", nil)
	}

	codeHash := HashInvitationCode(invitationCode)
	invitation, err := s.invitationRepo.FindByCodeHash(ctx, codeHash)
	if err != nil {
		return nil, exceptions.NewNotFoundException("invitation not found", err)
	}

	if invitation.IsExpired() {
		return nil, exceptions.NewValidationException("invitation has expired", nil)
	}

	if invitation.IsCompleted() {
		return nil, exceptions.NewValidationException("invitation has already been completed", nil)
	}

	if claims.Email != "" && !strings.EqualFold(claims.Email, invitation.Email()) {
		return nil, exceptions.NewValidationException("google account email does not match invitation email", nil)
	}

	if invitation.PersonID() == nil {
		return nil, exceptions.NewInternalException("invitation has no associated user record", nil)
	}

	personID := *invitation.PersonID()
	person, err := s.peopleRepository.FindByID(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to find invited user", err)
	}

	firstName := claims.FirstName
	lastName := claims.LastName
	displayName := claims.DisplayName

	if displayName == "" && claims.Email != "" {
		displayName = claims.Email
	}
	if firstName == "" {
		parts := strings.Fields(displayName)
		if len(parts) >= 2 {
			firstName = parts[0]
			lastName = strings.Join(parts[1:], " ")
		} else if len(parts) == 1 {
			firstName = parts[0]
			lastName = "User"
		} else {
			firstName = "Google"
			lastName = "User"
		}
	}
	if lastName == "" {
		lastName = "User"
	}
	if len(firstName) < 2 {
		firstName = firstName + " "
	}
	if len(lastName) < 2 {
		lastName = lastName + " "
	}

	if err := person.UpdateProfile(firstName, lastName, displayName); err != nil {
		return nil, exceptions.NewValidationException("invalid user profile data", err)
	}

	cred, err := NewCredentials(GOOGLE, claims.Sub, claims.Sub)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to prepare credentials", err)
	}

	updateFn := func(txCtx context.Context) (any, error) {
		if err := s.peopleRepository.UpdateProfile(txCtx, person.ID(), person.FirstName(), person.LastName(), person.DisplayName()); err != nil {
			return nil, err
		}

		if err := s.peopleRepository.SaveCredentials(txCtx, person.ID(), cred); err != nil {
			return nil, err
		}

		if err := s.peopleRepository.SaveRoles(txCtx, person.ID(), []Role{RoleUser}); err != nil {
			return nil, err
		}

		if err := s.invitationRepo.Complete(txCtx, invitation.ID()); err != nil {
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

func (s *AuthServiceImpl) loginWithLocal(ctx context.Context, credentials Credentials) (*core.AuthResponse, error) {
	personID, hash, err := s.peopleRepository.FindCredentials(
		ctx,
		credentials.Provider(),
		credentials.RegistrationID())

	if err != nil {
		return nil, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	if !s.hasher.Compare(credentials.Identifier(), hash) {
		return nil, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	return s.authorizePerson(ctx, personID)
}

func (s *AuthServiceImpl) authorizePerson(ctx context.Context, personID uuid.UUID) (*core.AuthResponse, error) {
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

func (s *AuthServiceImpl) loginWithGoogle(ctx context.Context, token string) (*core.AuthResponse, error) {
	claims, err := s.google.ValidateToken(token)

	if err != nil {
		log.Printf("[GOOGLE AUTH ERROR] Failed to validate Google token: %v", err)
		return nil, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
	}

	log.Printf("[GOOGLE AUTH] Successfully validated Google token for Sub ID: %s (DisplayName: %s)", claims.Sub, claims.DisplayName)

	personID, _, err := s.peopleRepository.FindCredentials(ctx, GOOGLE, claims.Sub)

	if err != nil {
		log.Printf("[GOOGLE AUTH ERROR] No user found for Google Sub ID: %s. To map this user, insert credentials with provider='GOOGLE' and registration_id='%s'", claims.Sub, claims.Sub)
		return nil, exceptions.NewUnauthorizedException("login failed", errors.Join(loginFailedError, err))
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
		Roles:       person.RoleStrings(),
	})

	if err != nil {
		return nil, errors.Join(core.RefreshTokenIsInvalid, err)
	}

	return authResponse, nil
}
