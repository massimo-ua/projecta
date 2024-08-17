package projecta

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

const (
	FailedToCreatePayment = "failed to create payment"
	FailedToFindPayment   = "failed to find payment"
)

type PaymentServiceImpl struct {
	payments   PaymentRepository
	categories CategoryRepository
	types      TypeRepository
	projects   ProjectRepository
	people     PeopleService
}

func (s *PaymentServiceImpl) Update(ctx context.Context, command UpdatePaymentCommand) error {
	//TODO implement me
	panic("implement me")
}

func (s *PaymentServiceImpl) Remove(ctx context.Context, command RemovePaymentCommand) error {
	e, err := s.payments.FindOne(ctx, PaymentFilter{
		PaymentID: command.ID,
		ProjectID: command.ProjectID,
	})

	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return exceptions.NewNotFoundException(FailedToFindPayment, err)
		}

		return exceptions.NewInternalException(FailedToFindPayment, err)
	}

	return s.payments.Remove(ctx, e)
}

func NewPaymentService(
	payments PaymentRepository,
	types TypeRepository,
	projects ProjectRepository,
	people PeopleService,
) *PaymentServiceImpl {
	return &PaymentServiceImpl{
		payments: payments,
		types:    types,
		projects: projects,
		people:   people,
	}
}

func (s *PaymentServiceImpl) Create(ctx context.Context, command CreatePaymentCommand) (*Payment, error) {
	personID := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)

	if personID == uuid.Nil {
		return nil, exceptions.NewInternalException(FailedToCreatePayment, core.FailedToIdentifyRequester)
	}

	owner, err := s.people.FindOwner(ctx, personID)

	costType, err := s.types.FindOne(ctx, TypeFilter{TypeID: command.TypeID, ProjectID: command.ProjectID})

	if err != nil {
		return nil, exceptions.NewValidationException(FailedToCreatePayment, err)
	}

	project, err := s.projects.FindOne(ctx, ProjectFilter{ProjectID: command.ProjectID})

	if err != nil {
		return nil, exceptions.NewValidationException(FailedToCreatePayment, err)
	}

	paymentDate := core.DateOrNow(command.PaymentDate)

	payment := NewPayment(
		uuid.New(),
		project,
		owner,
		costType,
		command.Description,
		command.Amount,
		paymentDate,
		command.Kind,
	)

	err = s.payments.Save(ctx, payment)

	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *PaymentServiceImpl) Find(ctx context.Context, filter PaymentCollectionFilter) (*PaymentCollection, error) {
	collection, err := s.payments.Find(ctx, filter)

	if err != nil {
		return nil, exceptions.NewInternalException(FailedToFindPayment, err)
	}

	return collection, nil
}
