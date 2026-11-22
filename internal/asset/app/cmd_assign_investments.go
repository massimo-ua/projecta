package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type AssignInvestments struct {
	RequesterID   core.PersonID
	AssetID       core.AssetID
	ProjectID     core.ProjectID
	InvestmentIDs []core.InvestmentID
}

type AssignInvestmentsHandler struct {
	repo        asset.Repository
	projects    ProjectReader
	investments InvestmentSource
	logger      core.Logger
}

func NewAssignInvestmentsHandler(repo asset.Repository, projects ProjectReader, investments InvestmentSource, logger core.Logger) AssignInvestmentsHandler {
	if repo == nil {
		panic("missing asset repository")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return AssignInvestmentsHandler{repo: repo, projects: projects, investments: investments, logger: logger}
}

func (h AssignInvestmentsHandler) Handle(ctx context.Context, cmd AssignInvestments) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "AssignInvestments", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to assign investments to asset", nil)
	}
	if cmd.AssetID.IsNil() {
		return exceptions.NewValidationException("asset id is required", nil)
	}
	if cmd.ProjectID.IsNil() {
		return exceptions.NewValidationException("project id is required", nil)
	}
	if len(cmd.InvestmentIDs) == 0 {
		return exceptions.NewValidationException("at least one investment must be selected", nil)
	}

	if _, err := h.projects.FindProject(ctx, cmd.ProjectID, cmd.RequesterID); err != nil {
		return preserveOrWrapError(err, "failed to assign investments to asset")
	}

	if _, err := h.repo.FindOne(ctx, asset.Filter{
		ID:        cmd.AssetID,
		ProjectID: cmd.ProjectID,
		OwnerID:   cmd.RequesterID,
	}); err != nil {
		return exceptions.NewValidationException("asset not found in project", err)
	}

	if h.investments != nil {
		items, err := h.investments.FindInvestments(ctx, cmd.ProjectID, cmd.RequesterID, cmd.InvestmentIDs)
		if err != nil || len(items) != len(cmd.InvestmentIDs) {
			return exceptions.NewValidationException("one or more investments not found in project", err)
		}
		if err = h.investments.AssignToAsset(ctx, cmd.AssetID, cmd.RequesterID, cmd.InvestmentIDs); err != nil {
			return preserveOrWrapError(err, "failed to assign investments to asset")
		}
	}
	return nil
}
