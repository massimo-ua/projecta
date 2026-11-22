package app

import (
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/investment"
)

// Application exposes all CQRS Command and Query handlers for the investment bounded context.
type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateInvestment         CreateInvestmentHandler
	CreateBatchInvestments   CreateBatchInvestmentsHandler
	UpdateInvestment         UpdateInvestmentHandler
	AssignInvestmentsToAsset AssignInvestmentsToAssetHandler
	RemoveInvestment         RemoveInvestmentHandler
}

type Queries struct {
	GetInvestment   GetInvestmentHandler
	ListInvestments ListInvestmentsHandler
	ListProjectTags ListProjectTagsHandler
}

// NewApplication wires all investment CQRS handlers and fails fast on missing required dependencies.
func NewApplication(
	db core.DbConnection,
	repo investment.Repository,
	tagsReadModel TagsReadModel,
	assets AssetReader,
	projects ProjectReader,
	contributors ContributorReader,
	converter investment.CurrencyConverter,
	logger core.Logger,
) *Application {
	if db == nil {
		panic("missing db connection")
	}
	if repo == nil {
		panic("missing investment repository")
	}
	if tagsReadModel == nil {
		tagsReadModel = repo
	}
	if assets == nil {
		panic("missing asset reader")
	}
	if projects == nil {
		panic("missing project reader")
	}
	if contributors == nil {
		panic("missing contributor reader")
	}

	return &Application{
		Commands: Commands{
			CreateInvestment:         NewCreateInvestmentHandler(db, repo, assets, projects, contributors, converter, logger),
			CreateBatchInvestments:   NewCreateBatchInvestmentsHandler(db, repo, assets, projects, contributors, converter, logger),
			UpdateInvestment:         NewUpdateInvestmentHandler(repo, assets, projects, converter, logger),
			AssignInvestmentsToAsset: NewAssignInvestmentsToAssetHandler(db, repo, assets, projects, logger),
			RemoveInvestment:         NewRemoveInvestmentHandler(db, repo, projects, logger),
		},
		Queries: Queries{
			GetInvestment:   NewGetInvestmentHandler(repo, projects, converter, logger),
			ListInvestments: NewListInvestmentsHandler(repo, projects, converter, logger),
			ListProjectTags: NewListProjectTagsHandler(tagsReadModel, projects, logger),
		},
	}
}
