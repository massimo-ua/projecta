package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

type GetProjectTotals struct {
	ProjectID   core.ProjectID
	RequesterID core.PersonID
}

type GetProjectTotalsHandler struct {
	repo            project.ProjectRepository
	totalsReadModel TotalsReadModel
	logger          core.Logger
}

func NewGetProjectTotalsHandler(repo project.ProjectRepository, totalsReadModel TotalsReadModel, logger core.Logger) GetProjectTotalsHandler {
	if repo == nil {
		panic("missing project repository")
	}
	return GetProjectTotalsHandler{
		repo:            repo,
		totalsReadModel: totalsReadModel,
		logger:          logger,
	}
}

func (h GetProjectTotalsHandler) Handle(ctx context.Context, q GetProjectTotals) (res project.ProjectTotals, err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "query", "GetProjectTotals", start, err)
	}()

	if q.RequesterID.IsNil() {
		return project.ProjectTotals{}, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}

	proj, err := h.repo.GetProject(ctx, q.ProjectID, q.RequesterID)
	if err != nil {
		return project.ProjectTotals{}, preserveOrWrapError(err, "project not found")
	}

	homeCurrency := proj.MainCurrency()
	if homeCurrency == "" {
		homeCurrency = "UAH"
	}

	if h.totalsReadModel == nil {
		return project.ProjectTotals{Currency: homeCurrency}, nil
	}

	return h.totalsReadModel.CalculateTotals(ctx, q.ProjectID, q.RequesterID, homeCurrency)
}
