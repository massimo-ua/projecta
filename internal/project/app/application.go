package app

import (
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

// Application exposes all CQRS Command and Query handlers for the project bounded context.
type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateProject       CreateProjectHandler
	UpdateProject       UpdateProjectHandler
	UpdateProjectBudget UpdateProjectBudgetHandler
	RemoveProject       RemoveProjectHandler
	AcceptProjectShare  AcceptProjectShareHandler
}

type Queries struct {
	GetProject       GetProjectHandler
	ListProjects     ListProjectsHandler
	GetProjectTotals GetProjectTotalsHandler
}

// NewApplication wires all project CQRS handlers and fails fast on missing required dependencies.
func NewApplication(
	repo project.ProjectRepository,
	ownerReader OwnerReader,
	totalsReadModel TotalsReadModel,
	converter project.CurrencyConverter,
	logger core.Logger,
) *Application {
	if repo == nil {
		panic("missing project repository")
	}
	if ownerReader == nil {
		panic("missing owner reader")
	}

	return &Application{
		Commands: Commands{
			CreateProject:       NewCreateProjectHandler(repo, ownerReader, converter, logger),
			UpdateProject:       NewUpdateProjectHandler(repo, converter, logger),
			UpdateProjectBudget: NewUpdateProjectBudgetHandler(repo, converter, logger),
			RemoveProject:       NewRemoveProjectHandler(repo, logger),
			AcceptProjectShare:  NewAcceptProjectShareHandler(repo, logger),
		},
		Queries: Queries{
			GetProject:       NewGetProjectHandler(repo, converter, logger),
			ListProjects:     NewListProjectsHandler(repo, converter, logger),
			GetProjectTotals: NewGetProjectTotalsHandler(repo, totalsReadModel, logger),
		},
	}
}
