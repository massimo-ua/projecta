package projecta

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	p, err := s.payments.FindOne(ctx, PaymentFilter{
		PaymentID: command.ID,
		ProjectID: command.ProjectID,
	})

	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return exceptions.NewNotFoundException(FailedToFindPayment, err)
		}

		return exceptions.NewInternalException(FailedToFindPayment, err)
	}

	costType, err := s.types.FindOne(ctx, TypeFilter{TypeID: command.TypeID, ProjectID: command.ProjectID})

	if err != nil {
		return exceptions.NewValidationException(FailedToFindPayment, err)
	}

	paymentDate := core.DateOrNow(command.PaymentDate)

	p.Type = costType
	p.Description = command.Description
	p.Amount = command.Amount
	p.Date = paymentDate
	p.Kind = command.Kind

	return s.payments.Save(ctx, p)
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

func (s *PaymentServiceImpl) FindOne(ctx context.Context, filter PaymentFilter) (*Payment, error) {
	p, err := s.payments.FindOne(ctx, filter)

	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return nil, exceptions.NewNotFoundException(FailedToFindPayment, err)
		}

		return nil, exceptions.NewInternalException(FailedToFindPayment, err)
	}

	return p, nil
}

func (s *PaymentServiceImpl) CreateBatch(ctx context.Context, commands []CreatePaymentCommand) ([]*Payment, error) {
	if len(commands) == 0 {
		return []*Payment{}, nil
	}

	requesterVal := ctx.Value(core.RequesterIDContextKey)
	if requesterVal == nil {
		return nil, exceptions.NewInternalException(FailedToCreatePayment, core.FailedToIdentifyRequester)
	}
	personID, ok := requesterVal.(uuid.UUID)
	if !ok || personID == uuid.Nil {
		return nil, exceptions.NewInternalException(FailedToCreatePayment, core.FailedToIdentifyRequester)
	}

	owner, err := s.people.FindOwner(ctx, personID)
	if err != nil {
		return nil, exceptions.NewValidationException(FailedToCreatePayment, err)
	}

	projectCache := make(map[uuid.UUID]*Project)
	typeCache := make(map[string]*CostType)
	createdPayments := make([]*Payment, 0, len(commands))

	for _, cmd := range commands {
		proj, ok := projectCache[cmd.ProjectID]
		if !ok {
			p, err := s.projects.FindOne(ctx, ProjectFilter{ProjectID: cmd.ProjectID})
			if err != nil {
				return nil, exceptions.NewValidationException(FailedToCreatePayment, err)
			}
			projectCache[cmd.ProjectID] = p
			proj = p
		}

		typeKey := fmt.Sprintf("%s_%s", cmd.ProjectID, cmd.TypeID)
		cType, ok := typeCache[typeKey]
		if !ok {
			t, err := s.types.FindOne(ctx, TypeFilter{TypeID: cmd.TypeID, ProjectID: cmd.ProjectID})
			if err != nil {
				return nil, exceptions.NewValidationException(FailedToCreatePayment, err)
			}
			typeCache[typeKey] = t
			cType = t
		}

		paymentDate := core.DateOrNow(cmd.PaymentDate)
		payment := NewPayment(
			uuid.New(),
			proj,
			owner,
			cType,
			cmd.Description,
			cmd.Amount,
			paymentDate,
			cmd.Kind,
		)
		createdPayments = append(createdPayments, payment)
	}

	if err := s.payments.SaveBatch(ctx, createdPayments); err != nil {
		return nil, err
	}

	return createdPayments, nil
}

func (s *PaymentServiceImpl) ParseStatement(ctx context.Context, projectID uuid.UUID, fileData []byte) (*StatementParseResult, error) {
	requesterVal := ctx.Value(core.RequesterIDContextKey)
	if requesterVal == nil {
		return nil, exceptions.NewInternalException("failed to parse statement", core.FailedToIdentifyRequester)
	}
	personID, ok := requesterVal.(uuid.UUID)
	if !ok || personID == uuid.Nil {
		return nil, exceptions.NewInternalException("failed to parse statement", core.FailedToIdentifyRequester)
	}

	_, err := s.projects.FindOne(ctx, ProjectFilter{ProjectID: projectID})
	if err != nil {
		return nil, exceptions.NewValidationException("project not found", err)
	}

	statement, err := ParseKredobankStatement(fileData)
	if err != nil {
		return nil, exceptions.NewValidationException("failed to parse statement: "+err.Error(), err)
	}

	existingPayments, err := s.payments.Find(ctx, PaymentCollectionFilter{
		Pagination: core.Pagination{Limit: 10000},
		ProjectID:  projectID,
	})
	if err != nil {
		return nil, exceptions.NewInternalException("failed to check existing payments", err)
	}

	existingMap := make(map[string]bool)
	if existingPayments != nil {
		for _, p := range existingPayments.Elements() {
			key := fmt.Sprintf("%s_%d_%s", p.Date.Format("2006-01-02"), p.Amount.Amount(), strings.TrimSpace(p.Description))
			existingMap[key] = true
		}
	}

	items := make([]StatementParseItem, 0, len(statement.Transactions))
	for _, tx := range statement.Transactions {
		key := fmt.Sprintf("%s_%d_%s", tx.Date.Format("2006-01-02"), tx.AmountMinor, strings.TrimSpace(tx.Description))
		isDup := existingMap[key]
		items = append(items, StatementParseItem{
			DocNum:      tx.DocNum,
			Date:        tx.Date,
			Amount:      tx.AmountMinor,
			Currency:    tx.Currency,
			Description: tx.Description,
			IsDuplicate: isDup,
		})
	}

	return &StatementParseResult{
		Account:      statement.Account,
		Currency:     statement.Currency,
		Period:       statement.Period,
		Transactions: items,
	}, nil
}

