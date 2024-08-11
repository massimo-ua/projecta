package projecta

import (
	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"time"
)

type ExpenseKind string

const (
	DownPayment           ExpenseKind = "DOWN_PAYMENT"
	UponCompletionPayment ExpenseKind = "UPON_COMPLETION"
	CreditPayment         ExpenseKind = "CREDIT_PAYMENT"
)

type Expense struct {
	ID           uuid.UUID
	Project      *Project
	Owner        *Owner
	Type         *CostType
	Description  string
	Amount       *money.Money
	Date         time.Time
	Kind         ExpenseKind
	Compensation *Expense
}

func ToExpenseKind(kind string) (ExpenseKind, error) {
	switch true {
	case kind == DownPayment.String():
		return DownPayment, nil
	case kind == UponCompletionPayment.String():
		return UponCompletionPayment, nil
	case kind == CreditPayment.String():
		return CreditPayment, nil
	default:
		return "", exceptions.NewValidationException("invalid expense kind", nil)
	}
}

func (e ExpenseKind) String() string {
	return string(e)
}

func NewExpense(
	id uuid.UUID,
	project *Project,
	person *Owner,
	costType *CostType,
	description string,
	amount *money.Money,
	date time.Time,
	kind ExpenseKind) *Expense {
	return &Expense{
		ID:          id,
		Project:     project,
		Owner:       person,
		Type:        costType,
		Description: description,
		Amount:      amount,
		Date:        date,
		Kind:        kind,
	}
}

type ExpenseCollection = core.PaginatedCollection[*Expense]

func NewExpenseCollection(total int) *ExpenseCollection {
	return core.NewPaginatedCollection[*Expense](total)
}

func (e *Expense) SetCompensation(compensation *Expense) {
	e.Compensation = compensation
}
