package projecta

import (
    "context"
    "github.com/google/uuid"
    "gitlab.com/massimo-ua/projecta/internal/core"
    "gitlab.com/massimo-ua/projecta/internal/exceptions"
    "time"
)

const (
    FailedToCreateExpense = "failed to create expense"
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
    //TODO implement me
    panic("implement me")
}

func NewExpenseService(
    expenses ExpenseRepository,
    categories CategoryRepository,
    types TypeRepository,
    projects ProjectRepository,
    people PeopleService,
) *ExpenseServiceImpl {
    return &ExpenseServiceImpl{
        expenses:   expenses,
        categories: categories,
        types:      types,
        projects:   projects,
        people:     people,
    }
}

func (s *ExpenseServiceImpl) Create(ctx context.Context, command CreateExpenseCommand) (*Expense, error) {
    personID := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)

    if personID == uuid.Nil {
        return nil, exceptions.NewInternalException(FailedToCreateExpense, core.FailedToIdentifyRequester)
    }

    owner, err := s.people.FindOwner(ctx, personID)

    category, err := s.categories.FindOne(ctx, CategoryFilter{CategoryID: command.CategoryID, ProjectID: command.ProjectID})

    if err != nil {
        return nil, exceptions.NewValidationException(FailedToCreateExpense, err)
    }

    costType, err := s.types.FindOne(ctx, TypeFilter{TypeID: command.TypeID, ProjectID: command.ProjectID})

    if err != nil {
        return nil, exceptions.NewValidationException(FailedToCreateExpense, err)
    }

    project, err := s.projects.FindOne(ctx, ProjectFilter{ProjectID: command.ProjectID})

    if err != nil {
        return nil, exceptions.NewValidationException(FailedToCreateExpense, err)
    }

    if err != nil {
        return nil, err
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
        category,
        command.Description,
        command.Amount,
        expenseDate,
    )

    if err != nil {
        return nil, exceptions.NewValidationException(FailedToCreateExpense, err)
    }

    err = s.expenses.Save(ctx, expense)

    if err != nil {
        return nil, err
    }

    return expense, nil
}
