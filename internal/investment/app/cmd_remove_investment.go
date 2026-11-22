package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
)

type RemoveInvestment struct {
	ID          core.InvestmentID
	ProjectID   core.ProjectID
	RequesterID core.PersonID
}

type RemoveInvestmentHandler struct {
	db       core.DbConnection
	repo     investment.Repository
	projects ProjectReader
	logger   core.Logger
}

func NewRemoveInvestmentHandler(
	db core.DbConnection,
	repo investment.Repository,
	projects ProjectReader,
	logger core.Logger,
) RemoveInvestmentHandler {
	if db == nil {
		panic("missing db connection")
	}
	if repo == nil {
		panic("missing investment repository")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return RemoveInvestmentHandler{
		db:       db,
		repo:     repo,
		projects: projects,
		logger:   logger,
	}
}

func (h RemoveInvestmentHandler) Handle(ctx context.Context, cmd RemoveInvestment) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "RemoveInvestment", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException(failedToFindInvestment, nil)
	}
	if !cmd.ProjectID.IsNil() {
		if _, err := verifyProjectAccess(ctx, h.projects, cmd.ProjectID, cmd.RequesterID); err != nil {
			return err
		}
	}

	inv, err := h.repo.FindOne(ctx, investment.Filter{
		ID:        cmd.ID,
		ProjectID: cmd.ProjectID,
	})
	if err != nil {
		return preserveOrWrapError(err, failedToFindInvestment)
	}
	if cmd.ProjectID.IsNil() && !inv.ProjectID().IsNil() {
		if _, err := verifyProjectAccess(ctx, h.projects, inv.ProjectID(), cmd.RequesterID); err != nil {
			return err
		}
	}

	_, err = h.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		if err := h.repo.Remove(txCtx, inv); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return preserveOrWrapError(err, "failed to remove investment")
}
