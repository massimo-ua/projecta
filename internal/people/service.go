package people

import (
    "context"
    "errors"
    "github.com/google/uuid"
    "gitlab.com/massimo-ua/projecta/internal/core"
    "gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type ServiceImpl struct {
    peopleRepository Repository
    hasher           core.Hasher
}

func (s *ServiceImpl) FindByID(ctx context.Context, personID uuid.UUID) (*Person, error) {
    return s.peopleRepository.FindByID(ctx, personID)
}

var loginFailedError = errors.New("failed to login")
var customerRegistrationFailedError = errors.New("failed to register customer")

func NewCustomerService(
    repo Repository,
    hasher core.Hasher) UserService {
    return &ServiceImpl{
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

    txCtx, err := s.peopleRepository.TxCtx(ctx)

    if err != nil {
        return errors.Join(customerRegistrationFailedError, err)
    }

    if err := s.peopleRepository.Register(txCtx, person); err != nil {
        return err
    }

    return nil
}
