package people

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type ServiceImpl struct {
	db               core.DbConnection
	peopleRepository Repository
	hasher           core.Hasher
}

func (s *ServiceImpl) FindByID(ctx context.Context, personID uuid.UUID) (*Person, error) {
	return s.peopleRepository.FindByID(ctx, personID)
}

var loginFailedError = errors.New("failed to login")
var customerRegistrationFailedError = errors.New("failed to register customer")

func NewCustomerService(
	db core.DbConnection,
	repo Repository,
	hasher core.Hasher) UserService {
	return &ServiceImpl{
		db:               db,
		peopleRepository: repo,
		hasher:           hasher,
	}
}

func (s *ServiceImpl) Register(ctx context.Context, command RegisterCommand) error {
	displayName := ""
	token := command.Token

	login := command.Login

	if command.IdentityProvider == LOCAL {
		email, err := NewEmailAddress(command.Login)

		if err != nil {
			return err
		}

		login = email.String()

		hash, err := s.hasher.Hash(command.Token)
		if err != nil {
			return err
		}

		token = hash
	}

	credentials, err := NewCredentials(command.IdentityProvider, login, token)

	if err != nil {
		return exceptions.NewValidationException(customerRegistrationFailedError.Error(), err)
	}

	person, err := NewPerson(uuid.Nil, command.FirstName, command.LastName, displayName, []Credentials{credentials}, RoleUser)

	if err != nil {
		return exceptions.NewValidationException(customerRegistrationFailedError.Error(), err)
	}

	_, err = s.db.Tx(ctx, func(ctx context.Context) (any, error) {
		if err = s.peopleRepository.Register(ctx, person); err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return exceptions.NewInternalException(customerRegistrationFailedError.Error(), err)
	}

	return nil
}

func (s *ServiceImpl) FindAll(ctx context.Context, pagination core.Pagination) ([]*Person, int, error) {
	return s.peopleRepository.FindAll(ctx, pagination)
}

func (s *ServiceImpl) AssignRoles(ctx context.Context, command AssignRolesCommand) error {
	person, err := s.peopleRepository.FindByID(ctx, command.PersonID)
	if err != nil {
		return err
	}

	if err := person.AssignRoles(command.Roles); err != nil {
		return exceptions.NewValidationException("invalid roles", err)
	}

	_, err = s.db.Tx(ctx, func(ctx context.Context) (any, error) {
		if err := s.peopleRepository.SaveRoles(ctx, person.ID(), person.Roles()); err != nil {
			return nil, err
		}
		return nil, nil
	})

	if err != nil {
		return exceptions.NewInternalException("failed to assign roles", err)
	}

	return nil
}

func (s *ServiceImpl) UpdateDisplayName(ctx context.Context, command UpdateDisplayNameCommand) (*Person, error) {
	person, err := s.peopleRepository.FindByID(ctx, command.PersonID)
	if err != nil {
		return nil, err
	}

	person.UpdateDisplayName(command.DisplayName)

	_, err = s.db.Tx(ctx, func(ctx context.Context) (any, error) {
		if err := s.peopleRepository.UpdateProfile(ctx, person.ID(), person.FirstName(), person.LastName(), person.displayName.String()); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return nil, exceptions.NewInternalException("failed to update display name", err)
	}

	return person, nil
}

