package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type ListAssets struct {
	core.Pagination
	core.Sorting
	RequesterID core.PersonID
	ProjectID   core.ProjectID
	TypeID      uuid.UUID
	Name        string
}

type ListAssetsHandler struct {
	repo        asset.Repository
	projects    ProjectReader
	investments InvestmentSource
	converter   asset.CurrencyConverter
	logger      core.Logger
}

func NewListAssetsHandler(
	repo asset.Repository,
	projects ProjectReader,
	investments InvestmentSource,
	converter asset.CurrencyConverter,
	logger core.Logger,
) ListAssetsHandler {
	if repo == nil {
		panic("missing asset repository")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return ListAssetsHandler{
		repo:        repo,
		projects:    projects,
		investments: investments,
		converter:   converter,
		logger:      logger,
	}
}

func (h ListAssetsHandler) Handle(ctx context.Context, q ListAssets) (res *asset.Collection, err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "query", "ListAssets", start, err)
	}()

	if q.RequesterID.IsNil() {
		return nil, exceptions.NewUnauthorizedException(failedToFindAsset, nil)
	}

	var defaultHomeCurrency string
	if !q.ProjectID.IsNil() {
		proj, err := h.projects.FindProject(ctx, q.ProjectID, q.RequesterID)
		if err != nil {
			return nil, preserveOrWrapError(err, failedToFindAsset)
		}
		defaultHomeCurrency = proj.MainCurrency
	}

	collection, err := h.repo.Find(ctx, asset.CollectionFilter{
		Pagination: q.Pagination,
		Sorting:    q.Sorting,
		ProjectID:  q.ProjectID,
		TypeID:     q.TypeID,
		OwnerID:    q.RequesterID,
		Name:       q.Name,
	})
	if err != nil {
		return nil, preserveOrWrapError(err, failedToFindAsset)
	}

	if collection != nil {
		if err := reconcileAndEvaluateAssets(ctx, h.repo, h.projects, h.investments, h.converter, q.RequesterID, q.ProjectID, defaultHomeCurrency, collection.Elements()); err != nil {
			return nil, preserveOrWrapError(err, failedToFindAsset)
		}
	}

	return collection, nil
}
