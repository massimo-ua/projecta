package projecta

import (
	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"time"
)

type CreateCategoryCommand struct {
	ProjectID   uuid.UUID
	PersonID    uuid.UUID
	Name        string
	Description string
}

type UpdateCategoryCommand struct {
	ProjectID   uuid.UUID
	PersonID    uuid.UUID
	ID          uuid.UUID
	Name        string
	Description string
}

type RemoveCategoryCommand struct {
	ProjectID uuid.UUID
	PersonID  uuid.UUID
	ID        uuid.UUID
}

type CreateProjectCommand struct {
	PersonID    uuid.UUID
	Name        string
	Description string
}

type UpdateProjectCommand struct {
	ProjectID   uuid.UUID
	PersonID    uuid.UUID
	Name        string
	Description string
}

type RemoveProjectCommand struct {
	ProjectID uuid.UUID
	PersonID  uuid.UUID
}

type CreateTypeCommand struct {
	ProjectID   uuid.UUID
	CategoryID  uuid.UUID
	PersonID    uuid.UUID
	Name        string
	Description string
}

type UpdateTypeCommand struct {
	ProjectID   uuid.UUID
	CategoryID  uuid.UUID
	PersonID    uuid.UUID
	ID          uuid.UUID
	Name        string
	Description string
}

type RemoveTypeCommand struct {
	ProjectID uuid.UUID
	ID        uuid.UUID
}

type CreateExpenseCommand struct {
	ProjectID       uuid.UUID
	TypeID          uuid.UUID
	Amount          *money.Money
	Description     string
	ExpenseDate     time.Time
	Kind            ExpenseKind
	FromDownPayment bool
}

type UpdateExpenseCommand struct {
	ProjectID   uuid.UUID
	PersonID    uuid.UUID
	ID          uuid.UUID
	TypeID      uuid.UUID
	Amount      *money.Money
	Description string
	ExpenseDate time.Time
}

type RemoveExpenseCommand struct {
	ProjectID uuid.UUID
	ID        uuid.UUID
}

type RemoveProjectResourceCommand struct {
	ProjectID  uuid.UUID
	ResourceID uuid.UUID
}
