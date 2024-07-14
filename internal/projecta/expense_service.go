package projecta

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"time"
)

const (
	FailedToCreateExpense = "failed to create expense"
	FailedToFindExpenses  = "failed to find expenses"
)

type ExpenseServiceImpl struct {
	expenses   ExpenseRepository
	categories CategoryRepository
	types      TypeRepository
	projects   ProjectRepository
	people     PeopleService
}

func (s *ExpenseServiceImpl) Update(ctx context.Context, command UpdateExpenseCommand) error {
	//TODO implement me
	panic("implement me")
}

func (s *ExpenseServiceImpl) Remove(ctx context.Context, command RemoveExpenseCommand) error {
	e, err := s.expenses.FindOne(ctx, ExpenseFilter{
		ExpenseID: command.ID,
		ProjectID: command.ProjectID,
	})

	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return exceptions.NewNotFoundException(FailedToFindExpenses, err)
		}

		return exceptions.NewInternalException(FailedToFindExpenses, err)
	}

	return s.expenses.Remove(ctx, e)
}

func NewExpenseService(
	expenses ExpenseRepository,
	types TypeRepository,
	projects ProjectRepository,
	people PeopleService,
) *ExpenseServiceImpl {
	return &ExpenseServiceImpl{
		expenses: expenses,
		types:    types,
		projects: projects,
		people:   people,
	}
}

func (s *ExpenseServiceImpl) Create(ctx context.Context, command CreateExpenseCommand) (*Expense, error) {
	personID := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)

	if personID == uuid.Nil {
		return nil, exceptions.NewInternalException(FailedToCreateExpense, core.FailedToIdentifyRequester)
	}

	owner, err := s.people.FindOwner(ctx, personID)

	costType, err := s.types.FindOne(ctx, TypeFilter{TypeID: command.TypeID, ProjectID: command.ProjectID})

	if err != nil {
		return nil, exceptions.NewValidationException(FailedToCreateExpense, err)
	}

	project, err := s.projects.FindOne(ctx, ProjectFilter{ProjectID: command.ProjectID})

	if err != nil {
		return nil, exceptions.NewValidationException(FailedToCreateExpense, err)
	}

	var expenseDate time.Time

	if command.ExpenseDate.IsZero() {
		expenseDate = time.Now()
	} else {
		expenseDate = command.ExpenseDate
	}

	expense := NewExpense(
		uuid.New(),
		project,
		owner,
		costType,
		command.Description,
		command.Amount,
		expenseDate,
	)

	err = s.expenses.Save(ctx, expense)

	if err != nil {
		return nil, err
	}

	return expense, nil
}

func (s *ExpenseServiceImpl) Find(ctx context.Context, filter ExpenseCollectionFilter) ([]*Expense, error) {
	expenses, err := s.expenses.Find(ctx, filter)

	if err != nil {
		return nil, exceptions.NewInternalException(FailedToFindExpenses, err)
	}

	return expenses, nil
}
