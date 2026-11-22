package people

import (
	"context"
	"errors"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type ServiceImpl struct {
	db               core.DbConnection
	peopleRepository Repository
	hasher           core.Hasher
}

func (s *ServiceImpl) FindByID(ctx context.Context, personID core.PersonID) (*Person, error) {
	return s.peopleRepository.FindByID(ctx, personID)
}

var loginFailedError = errors.New("failed to login")
var userRegistrationFailedError = errors.New("failed to register user")

func NewUserService(
	db core.DbConnection,
	repo Repository,
	hasher core.Hasher) UserService {
	if db == nil {
		panic("missing db connection")
	}
	if repo == nil {
		panic("missing people repository")
	}
	if hasher == nil {
		panic("missing password hasher")
	}
	return &ServiceImpl{
		db:               db,
		peopleRepository: repo,
		hasher:           hasher,
	}
}

func (s *ServiceImpl) Register(ctx context.Context, command RegisterCommand) error {
	if command.IdentityProvider != LOCAL {
		return exceptions.NewValidationException("unsupported identity provider for registration", nil)
	}

	if len(command.Token) < 8 || len([]byte(command.Token)) > 72 {
		return exceptions.NewValidationException("password must be between 8 and 72 bytes", nil)
	}

	email, err := NewEmailAddress(command.Login)
	if err != nil {
		return err
	}

	login := email.String()

	if existingID, _, findErr := s.peopleRepository.FindCredentials(ctx, LOCAL, login); findErr == nil && !existingID.IsNil() {
		return exceptions.NewValidationException("user with this email already exists", nil)
	}

	hash, err := s.hasher.Hash(command.Token)
	if err != nil {
		return err
	}

	displayName := ""
	credentials, err := NewCredentials(command.IdentityProvider, login, hash)

	if err != nil {
		return exceptions.NewValidationException(userRegistrationFailedError.Error(), err)
	}

	person, err := NewPerson(core.NilPersonID, command.FirstName, command.LastName, displayName, []Credentials{credentials}, RoleUser)

	if err != nil {
		return exceptions.NewValidationException(userRegistrationFailedError.Error(), err)
	}

	_, err = s.db.Tx(ctx, func(ctx context.Context) (any, error) {
		if err = s.peopleRepository.Save(ctx, person); err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return exceptions.NewInternalException(userRegistrationFailedError.Error(), err)
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
		if err := s.peopleRepository.Save(ctx, person); err != nil {
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
		if err := s.peopleRepository.Save(ctx, person); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return nil, exceptions.NewInternalException("failed to update display name", err)
	}

	return person, nil
}

