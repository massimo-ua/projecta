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
	token := command.Token

	if command.IdentityProvider == LOCAL {
		hash, err := s.hasher.Hash(command.Token)
		if err != nil {
			return err
		}

		token = hash
	}

	credentials, err := NewCredentials(command.IdentityProvider, command.Login, token)

	if err != nil {
		return exceptions.NewValidationException(customerRegistrationFailedError.Error(), err)
	}

	person, err := NewPerson(uuid.Nil, command.FirstName, command.LastName, []Credentials{credentials})

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
