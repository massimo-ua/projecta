package projecta

import (
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type CategoryFilter struct {
	Name       string
	ProjectID  uuid.UUID
	CategoryID uuid.UUID
}

type ProjectFilter struct {
	Name      string
	ProjectID uuid.UUID
}

type ProjectCollectionFilter struct {
	core.Pagination
	Name string
}

type TypeFilter struct {
	Name       string
	TypeID     uuid.UUID
	ProjectID  uuid.UUID
	CategoryID uuid.UUID
}

type ExpenseFilter struct {
	ExpenseID uuid.UUID
	ProjectID uuid.UUID
	TypeID    uuid.UUID
}

type TypeCollectionFilter struct {
	core.Pagination
	ProjectID  uuid.UUID
	CategoryID uuid.UUID
	Name       string
}

type CategoryCollectionFilter struct {
	core.Pagination
	ProjectID uuid.UUID
	Name      string
}

type ExpenseCollectionFilter struct {
	core.Pagination
	core.Sorting
	ProjectID  uuid.UUID
	CategoryID uuid.UUID
	TypeID     uuid.UUID
}
