package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type GetCashFlowForecast struct {
	RequesterID core.PersonID
	ProjectID   core.ProjectID
	FromMonth   string
	ToMonth     string
	AsOfDate    time.Time
	Tag         string
}

type GetCashFlowForecastHandler struct {
	repo        asset.Repository
	projects    ProjectReader
	investments InvestmentSource
	converter   asset.CurrencyConverter
	logger      core.Logger
}

func NewGetCashFlowForecastHandler(
	repo asset.Repository,
	projects ProjectReader,
	investments InvestmentSource,
	converter asset.CurrencyConverter,
	logger core.Logger,
) GetCashFlowForecastHandler {
	if repo == nil {
		panic("missing asset repository")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return GetCashFlowForecastHandler{
		repo:        repo,
		projects:    projects,
		investments: investments,
		converter:   converter,
		logger:      logger,
	}
}

func (h GetCashFlowForecastHandler) Handle(ctx context.Context, q GetCashFlowForecast) (res *asset.CashFlowForecast, err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "query", "GetCashFlowForecast", start, err)
	}()

	if q.RequesterID.IsNil() {
		return nil, exceptions.NewUnauthorizedException("failed to calculate cash flow forecast", nil)
	}
	if q.ProjectID.IsNil() {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}
	if err := asset.ValidateForecastTag(q.Tag); err != nil {
		return nil, err
	}

	proj, err := h.projects.FindProject(ctx, q.ProjectID, q.RequesterID)
	if err != nil {
		return nil, preserveOrWrapError(err, "failed to calculate cash flow forecast")
	}

	col, err := h.repo.Find(ctx, asset.CollectionFilter{
		OwnerID:   q.RequesterID,
		ProjectID: q.ProjectID,
	})
	if err != nil {
		return nil, preserveOrWrapError(err, "failed to calculate cash flow forecast")
	}
	var assets []*asset.Asset
	if col != nil {
		assets = col.Elements()
	}

	var investments []asset.InvestmentRecordSummary
	if h.investments != nil {
		investments, err = h.investments.ListProjectInvestments(ctx, q.ProjectID, q.RequesterID)
		if err != nil {
			return nil, preserveOrWrapError(err, "failed to calculate cash flow forecast")
		}
	}

	domainQuery := asset.GetCashFlowForecastQuery{
		ProjectID: q.ProjectID,
		FromMonth: q.FromMonth,
		ToMonth:   q.ToMonth,
		AsOfDate:  q.AsOfDate,
		Tag:       q.Tag,
	}

	months, _, _, err := asset.ResolveForecastMonths(proj, assets, investments, domainQuery)
	if err != nil {
		return nil, err
	}

	ceilings, err := h.projects.FindMonthlyCashCeilings(ctx, q.ProjectID, q.RequesterID, months, q.Tag)
	if err != nil {
		return nil, preserveOrWrapError(err, "failed to calculate cash flow forecast")
	}

	return asset.CalculateProjectCashFlowForecast(proj, assets, investments, ceilings, months, domainQuery, h.converter), nil
}
