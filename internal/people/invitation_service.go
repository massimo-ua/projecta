package people

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type InvitationServiceImpl struct {
	db             core.DbConnection
	invitationRepo InvitationRepository
	peopleRepo     Repository
}

func NewInvitationService(
	db core.DbConnection,
	invitationRepo InvitationRepository,
	peopleRepo Repository,
) InvitationService {
	return &InvitationServiceImpl{
		db:             db,
		invitationRepo: invitationRepo,
		peopleRepo:     peopleRepo,
	}
}

func HashInvitationCode(code string) string {
	invCode, err := NewInvitationCode(code)
	if err != nil {
		return ""
	}
	return invCode.Hash()
}

func (s *InvitationServiceImpl) Create(ctx context.Context, adminID uuid.UUID, email string) (*InvitationWithCode, error) {
	if adminID == uuid.Nil {
		return nil, exceptions.NewValidationException("admin id is required", nil)
	}

	emailAddr, err := NewEmailAddress(email)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid email address", err)
	}
	cleanEmail := emailAddr.String()

	exists, err := s.invitationRepo.EmailExists(ctx, cleanEmail)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to verify email", err)
	}
	if exists {
		return nil, exceptions.NewValidationException("user with this email already exists", errors.New("user already exists"))
	}

	invCode, err := GenerateInvitationCode()
	if err != nil {
		return nil, err
	}
	codeHash := invCode.Hash()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	personID := uuid.New()
	precreatedPerson, err := NewPerson(personID, "Invited", "User", cleanEmail, nil, RoleUser)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to prepare user record", err)
	}

	invitation, err := NewInvitation(
		uuid.Nil,
		cleanEmail,
		codeHash,
		expiresAt,
		adminID,
		&personID,
		InvitationStatusPending,
		time.Now(),
		nil,
	)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to prepare invitation", err)
	}

	_, err = s.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		if err := s.peopleRepo.Register(txCtx, precreatedPerson); err != nil {
			return nil, err
		}

		if err := s.invitationRepo.Create(txCtx, invitation); err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return nil, exceptions.NewInternalException("failed to create invitation", err)
	}

	return &InvitationWithCode{
		Invitation: invitation,
		Code:       invCode.String(),
	}, nil
}

func (s *InvitationServiceImpl) FindAllByCreator(ctx context.Context, adminID uuid.UUID) ([]*Invitation, error) {
	if adminID == uuid.Nil {
		return nil, exceptions.NewValidationException("admin id is required", nil)
	}
	return s.invitationRepo.FindByCreator(ctx, adminID)
}

func (s *InvitationServiceImpl) Delete(ctx context.Context, adminID uuid.UUID, invitationID uuid.UUID) error {
	inv, err := s.invitationRepo.FindByID(ctx, invitationID)
	if err != nil {
		return err
	}

	if inv.CreatedBy() != adminID {
		return exceptions.NewForbiddenException("not authorized to delete this invitation", nil)
	}

	_, err = s.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		if err := s.invitationRepo.Delete(txCtx, invitationID); err != nil {
			return nil, err
		}

		if inv.Status() == InvitationStatusPending && inv.PersonID() != nil {
			if err := s.peopleRepo.DeletePerson(txCtx, *inv.PersonID()); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	if err != nil {
		return exceptions.NewInternalException("failed to delete invitation", err)
	}

	return nil
}

func (s *InvitationServiceImpl) FindByCode(ctx context.Context, code string) (*Invitation, error) {
	invCode, err := NewInvitationCode(code)
	if err != nil {
		return nil, err
	}

	inv, err := s.invitationRepo.FindByCodeHash(ctx, invCode.Hash())
	if err != nil {
		return nil, err
	}

	if inv.IsExpired() {
		return nil, exceptions.NewValidationException("invitation has expired", nil)
	}

	if inv.IsCompleted() {
		return nil, exceptions.NewValidationException("invitation has already been completed", nil)
	}

	return inv, nil
}
