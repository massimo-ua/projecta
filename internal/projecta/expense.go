package projecta

import (
	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"time"
)

type Expense struct {
	ID          uuid.UUID
	Project     *Project
	Owner       *Owner
	Type        *CostType
	Description string
	Amount      *money.Money
	Date        time.Time
}

func NewExpense(id uuid.UUID, project *Project, person *Owner, costType *CostType, description string, amount *money.Money, date time.Time) *Expense {
	return &Expense{
		ID:          id,
		Project:     project,
		Owner:       person,
		Type:        costType,
		Description: description,
		Amount:      amount,
		Date:        date,
	}
}

type ExpenseCollection = core.PaginatedCollection[*Expense]

func NewExpenseCollection(total int) *ExpenseCollection {
	return core.NewPaginatedCollection[*Expense](total)
}
