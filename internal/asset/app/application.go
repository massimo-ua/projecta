package app

import (
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

// Application exposes all CQRS Command and Query handlers for the asset bounded context.
type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateAsset                CreateAssetHandler
	CreateAssetFromInvestments CreateAssetFromInvestmentsHandler
	UpdateAsset                UpdateAssetHandler
	RemoveAsset                RemoveAssetHandler
	LinkChildAsset             LinkChildAssetHandler
	UnlinkChildAsset           UnlinkChildAssetHandler
	LinkChildrenAssets         LinkChildrenAssetsHandler
	AssignInvestments          AssignInvestmentsHandler
	GroupAssets                GroupAssetsHandler
	RescheduleAsset            RescheduleAssetHandler
}

type Queries struct {
	GetAsset             GetAssetHandler
	GetAssetScheduleView GetAssetScheduleViewHandler
	ListAssets           ListAssetsHandler
	GetCashFlowForecast  GetCashFlowForecastHandler
}

// NewApplication wires all asset CQRS handlers and fails fast on missing required dependencies.
func NewApplication(
	db core.DbConnection,
	repo asset.Repository,
	owners OwnerReader,
	projects ProjectReader,
	investments InvestmentSource,
	converter asset.CurrencyConverter,
	logger core.Logger,
) *Application {
	if db == nil {
		panic("missing db connection")
	}
	if repo == nil {
		panic("missing asset repository")
	}
	if owners == nil {
		panic("missing owner reader")
	}
	if projects == nil {
		panic("missing project reader")
	}

	linkChild := NewLinkChildAssetHandler(repo, projects, logger)
	getAsset := NewGetAssetHandler(repo, projects, investments, converter, logger)

	return &Application{
		Commands: Commands{
			CreateAsset:                NewCreateAssetHandler(db, repo, owners, projects, investments, converter, logger),
			CreateAssetFromInvestments: NewCreateAssetFromInvestmentsHandler(db, repo, owners, projects, investments, converter, logger),
			UpdateAsset:                NewUpdateAssetHandler(repo, projects, converter, logger),
			RemoveAsset:                NewRemoveAssetHandler(db, repo, projects, logger),
			LinkChildAsset:             linkChild,
			UnlinkChildAsset:           NewUnlinkChildAssetHandler(repo, projects, logger),
			LinkChildrenAssets:         NewLinkChildrenAssetsHandler(db, linkChild, logger),
			AssignInvestments:          NewAssignInvestmentsHandler(repo, projects, investments, logger),
			GroupAssets:                NewGroupAssetsHandler(db, repo, owners, projects, logger),
			RescheduleAsset:            NewRescheduleAssetHandler(repo, projects, logger),
		},
		Queries: Queries{
			GetAsset:             getAsset,
			GetAssetScheduleView: NewGetAssetScheduleViewHandler(getAsset, repo, logger),
			ListAssets:           NewListAssetsHandler(repo, projects, investments, converter, logger),
			GetCashFlowForecast:  NewGetCashFlowForecastHandler(repo, projects, investments, converter, logger),
		},
	}
}
