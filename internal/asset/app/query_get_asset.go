package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type GetAsset struct {
	ID          core.AssetID
	ProjectID   core.ProjectID
	Name        string
	RequesterID core.PersonID
}

type GetAssetHandler struct {
	repo        asset.Repository
	projects    ProjectReader
	investments InvestmentSource
	converter   asset.CurrencyConverter
	logger      core.Logger
}

func NewGetAssetHandler(
	repo asset.Repository,
	projects ProjectReader,
	investments InvestmentSource,
	converter asset.CurrencyConverter,
	logger core.Logger,
) GetAssetHandler {
	if repo == nil {
		panic("missing asset repository")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return GetAssetHandler{
		repo:        repo,
		projects:    projects,
		investments: investments,
		converter:   converter,
		logger:      logger,
	}
}

func (h GetAssetHandler) Handle(ctx context.Context, q GetAsset) (res *asset.Asset, err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "query", "GetAsset", start, err)
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

	anAsset, err := h.repo.FindOne(ctx, asset.Filter{
		ID:        q.ID,
		ProjectID: q.ProjectID,
		OwnerID:   q.RequesterID,
		Name:      q.Name,
	})
	if err != nil {
		return nil, preserveOrWrapError(err, failedToFindAsset)
	}
	if anAsset == nil {
		return nil, exceptions.NewNotFoundException("asset not found", nil)
	}

	projID := q.ProjectID
	if projID.IsNil() {
		projID = anAsset.ProjectID()
		if !projID.IsNil() {
			proj, err := h.projects.FindProject(ctx, projID, q.RequesterID)
			if err != nil {
				return nil, preserveOrWrapError(err, failedToFindAsset)
			}
			defaultHomeCurrency = proj.MainCurrency
		}
	}
	if defaultHomeCurrency == "" {
		defaultHomeCurrency = "UAH"
	}
	if err := reconcileAndEvaluateAssets(ctx, h.repo, h.projects, h.investments, h.converter, q.RequesterID, projID, defaultHomeCurrency, []*asset.Asset{anAsset}); err != nil {
		return nil, preserveOrWrapError(err, failedToFindAsset)
	}
	return anAsset, nil
}

type AssetScheduleView struct {
	Asset    *asset.Asset
	Warnings []string
}

type GetAssetScheduleView struct {
	AssetID     core.AssetID
	ProjectID   core.ProjectID
	RequesterID core.PersonID
}

type GetAssetScheduleViewHandler struct {
	getAsset GetAssetHandler
	repo     asset.Repository
	logger   core.Logger
}

func NewGetAssetScheduleViewHandler(getAsset GetAssetHandler, repo asset.Repository, logger core.Logger) GetAssetScheduleViewHandler {
	if repo == nil {
		panic("missing asset repository")
	}
	return GetAssetScheduleViewHandler{getAsset: getAsset, repo: repo, logger: logger}
}

func (h GetAssetScheduleViewHandler) Handle(ctx context.Context, q GetAssetScheduleView) (res AssetScheduleView, err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "query", "GetAssetScheduleView", start, err)
	}()

	target, err := h.getAsset.Handle(ctx, GetAsset{
		ID:          q.AssetID,
		ProjectID:   q.ProjectID,
		RequesterID: q.RequesterID,
	})
	if err != nil {
		return AssetScheduleView{}, err
	}

	children := make([]*asset.Asset, 0, len(target.Children()))
	for _, ch := range target.Children() {
		if ca, err := h.repo.FindOne(ctx, asset.Filter{ID: ch.ChildID, ProjectID: target.ProjectID(), OwnerID: q.RequesterID}); err == nil && ca != nil {
			children = append(children, ca)
		}
	}
	parents := make([]*asset.Asset, 0, len(target.Parents()))
	for _, p := range target.Parents() {
		if pa, err := h.repo.FindOne(ctx, asset.Filter{ID: p.ParentID, ProjectID: target.ProjectID(), OwnerID: q.RequesterID}); err == nil && pa != nil {
			parents = append(parents, pa)
		}
	}

	return AssetScheduleView{
		Asset:    target,
		Warnings: asset.ComputeScheduleWarnings(target, children, parents),
	}, nil
}
