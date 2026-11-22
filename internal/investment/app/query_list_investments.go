package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
)

type ListInvestments struct {
	core.Pagination
	core.Sorting
	RequesterID  core.PersonID
	IDs          []core.InvestmentID
	ProjectID    core.ProjectID
	AssetID      core.AssetID
	OwnerID      core.PersonID
	ResourceType investment.ResourceType
	Tag          string
	StartDate    *time.Time
	EndDate      *time.Time
}

type ListInvestmentsHandler struct {
	repo      investment.Repository
	projects  ProjectReader
	converter investment.CurrencyConverter
	logger    core.Logger
}

func NewListInvestmentsHandler(
	repo investment.Repository,
	projects ProjectReader,
	converter investment.CurrencyConverter,
	logger core.Logger,
) ListInvestmentsHandler {
	if repo == nil {
		panic("missing investment repository")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return ListInvestmentsHandler{
		repo:      repo,
		projects:  projects,
		converter: converter,
		logger:    logger,
	}
}

func (h ListInvestmentsHandler) Handle(ctx context.Context, q ListInvestments) (res *investment.Collection, err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "query", "ListInvestments", start, err)
	}()

	if q.RequesterID.IsNil() {
		return nil, exceptions.NewUnauthorizedException(failedToFindInvestment, nil)
	}

	var proj ProjectDescriptor
	if !q.ProjectID.IsNil() {
		proj, err = verifyProjectAccess(ctx, h.projects, q.ProjectID, q.RequesterID)
		if err != nil {
			return nil, err
		}
	}

	col, err := h.repo.Find(ctx, investment.CollectionFilter{
		Pagination:   q.Pagination,
		Sorting:      q.Sorting,
		IDs:          q.IDs,
		ProjectID:    q.ProjectID,
		AssetID:      q.AssetID,
		OwnerID:      q.OwnerID,
		ResourceType: q.ResourceType,
		Tag:          q.Tag,
		StartDate:    q.StartDate,
		EndDate:      q.EndDate,
	})
	if err != nil {
		return nil, preserveOrWrapError(err, failedToFindInvestment)
	}
	if col != nil {
		for _, inv := range col.Elements() {
			if inv != nil {
				invProjCurrency := proj.MainCurrency
				if invProjCurrency == "" && !inv.ProjectID().IsNil() {
					if p, err := verifyProjectAccess(ctx, h.projects, inv.ProjectID(), q.RequesterID); err == nil {
						invProjCurrency = p.MainCurrency
					}
				}
				inv.EvaluateHomeAmount(h.converter, invProjCurrency)
			}
		}
	}
	return col, nil
}
