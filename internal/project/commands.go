package project

import (
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type CreateProjectCommand struct {
	PersonID     core.PersonID
	Name         string
	Description  string
	MainCurrency string
}

type UpdateProjectCommand struct {
	ProjectID    core.ProjectID
	PersonID     core.PersonID
	Name         string
	Description  string
	MainCurrency string
}

type RemoveProjectCommand struct {
	ProjectID core.ProjectID
	PersonID  core.PersonID
}

type RemoveProjectResourceCommand struct {
	ProjectID  core.ProjectID
	ResourceID uuid.UUID
}
