package project

import (
	"github.com/google/uuid"
)

type CreateProjectCommand struct {
	PersonID    uuid.UUID
	Name        string
	Description string
}

type UpdateProjectCommand struct {
	ProjectID    uuid.UUID
	PersonID     uuid.UUID
	Name         string
	Description  string
	MainCurrency string
}

type RemoveProjectCommand struct {
	ProjectID uuid.UUID
	PersonID  uuid.UUID
}

type RemoveProjectResourceCommand struct {
	ProjectID  uuid.UUID
	ResourceID uuid.UUID
}
