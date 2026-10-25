package project

import (
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
)

type CreateProjectCommand struct {
	PersonID     core.PersonID
	Name         string
	Description  string
	MainCurrency string
	StartDate    time.Time
	EndDate      *time.Time
	TagBudgets   []TagBudgetInput
}

type UpdateProjectCommand struct {
	ProjectID    core.ProjectID
	PersonID     core.PersonID
	Name         string
	Description  string
	MainCurrency string
	StartDate    *time.Time
	EndDate      *time.Time
	TagBudgets   *[]TagBudgetInput
}

type UpdateProjectBudgetCommand struct {
	ProjectID  core.ProjectID
	PersonID   core.PersonID
	TagBudgets []TagBudgetInput
}

type RemoveProjectCommand struct {
	ProjectID core.ProjectID
	PersonID  core.PersonID
}
