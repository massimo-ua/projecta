package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
)

type AssignInvestmentsToAsset struct {
	RequesterID   core.PersonID
	AssetID       core.AssetID
	InvestmentIDs []core.InvestmentID
}

type AssignInvestmentsToAssetHandler struct {
	db       core.DbConnection
	repo     investment.Repository
	assets   AssetReader
	projects ProjectReader
	logger   core.Logger
}

func NewAssignInvestmentsToAssetHandler(
	db core.DbConnection,
	repo investment.Repository,
	assets AssetReader,
	projects ProjectReader,
	logger core.Logger,
) AssignInvestmentsToAssetHandler {
	if db == nil {
		panic("missing db connection")
	}
	if repo == nil {
		panic("missing investment repository")
	}
	if assets == nil {
		panic("missing asset reader")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return AssignInvestmentsToAssetHandler{
		db:       db,
		repo:     repo,
		assets:   assets,
		projects: projects,
		logger:   logger,
	}
}

func (h AssignInvestmentsToAssetHandler) Handle(ctx context.Context, cmd AssignInvestmentsToAsset) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "AssignInvestmentsToAsset", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException(failedToUpdateInvestment, nil)
	}
	if len(cmd.InvestmentIDs) == 0 {
		return nil
	}

	ast, err := h.assets.FindAsset(ctx, core.NilProjectID, cmd.AssetID)
	if err != nil {
		return preserveOrWrapError(err, failedToUpdateInvestment)
	}
	if !ast.ProjectID.IsNil() {
		if _, err := verifyProjectAccess(ctx, h.projects, ast.ProjectID, cmd.RequesterID); err != nil {
			return err
		}
	}

	_, err = h.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		for _, invID := range cmd.InvestmentIDs {
			if err := h.repo.UpdateInvestment(
				txCtx,
				invID,
				ast.ProjectID,
				func(_ context.Context, inv *investment.Investment) (*investment.Investment, error) {
					if err := inv.Allocate([]investment.InvestmentAssetLink{
						{
							AssetID:         ast.ID,
							AssetName:       ast.Name,
							SharePercentage: 100.0,
						},
					}); err != nil {
						return nil, err
					}
					return inv, nil
				},
			); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return preserveOrWrapError(err, failedToUpdateInvestment)
}
