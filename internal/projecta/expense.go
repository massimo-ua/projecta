package projecta

import (
    "github.com/Rhymond/go-money"
    "github.com/google/uuid"
    "time"
)

type Expense struct {
    ID          uuid.UUID
    Project     *Project
    Owner       *Owner
    Type        *CostType
    Category    *CostCategory
    Description string
    Amount      *money.Money
    Date        time.Time
}

func NewExpense(id uuid.UUID, project *Project, person *Owner, costType *CostType, costCategory *CostCategory, description string, amount *money.Money, date time.Time) *Expense {
    return &Expense{
        ID:          id,
        Project:     project,
        Owner:       person,
        Type:        costType,
        Category:    costCategory,
        Description: description,
        Amount:      amount,
        Date:        date,
    }
}
