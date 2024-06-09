package projecta

import "github.com/google/uuid"

type CategoryFilter struct {
    Name       string
    ProjectID  uuid.UUID
    CategoryID uuid.UUID
}

type ProjectFilter struct {
    Name      string
    ProjectID uuid.UUID
}

type TypeFilter struct {
    Name      string
    TypeID    uuid.UUID
    ProjectID uuid.UUID
}

type ExpenseFilter struct {
    ExpenseID  uuid.UUID
    ProjectID  uuid.UUID
    CategoryID uuid.UUID
    TypeID     uuid.UUID
}
