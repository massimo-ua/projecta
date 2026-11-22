package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
)

type GetInvestment struct {
	ID          core.InvestmentID
	ProjectID   core.ProjectID
	AssetID     core.AssetID
	RequesterID core.PersonID
}

type GetInvestmentHandler struct {
	repo      investment.Repository
	projects  ProjectReader
	converter investment.CurrencyConverter
	logger    core.Logger
}

func NewGetInvestmentHandler(
	repo investment.Repository,
	projects ProjectReader,
	converter investment.CurrencyConverter,
	logger core.Logger,
) GetInvestmentHandler {
	if repo == nil {
		panic("missing investment repository")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return GetInvestmentHandler{
		repo:      repo,
		projects:  projects,
		converter: converter,
		logger:    logger,
	}
}

func (h GetInvestmentHandler) Handle(ctx context.Context, q GetInvestment) (res *investment.Investment, err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "query", "GetInvestment", start, err)
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

	inv, err := h.repo.FindOne(ctx, investment.Filter{
		ID:        q.ID,
		ProjectID: q.ProjectID,
		AssetID:   q.AssetID,
	})
	if err != nil {
		return nil, preserveOrWrapError(err, failedToFindInvestment)
	}
	if inv != nil {
		if proj.MainCurrency == "" && !inv.ProjectID().IsNil() {
			p, err := verifyProjectAccess(ctx, h.projects, inv.ProjectID(), q.RequesterID)
			if err != nil {
				return nil, err
			}
			proj = p
		}
		inv.EvaluateHomeAmount(h.converter, proj.MainCurrency)
	}
	return inv, nil
}
