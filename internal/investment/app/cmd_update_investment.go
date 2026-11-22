package app

import (
	"context"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
)

type UpdateInvestment struct {
	ID               core.InvestmentID
	RequesterID      core.PersonID
	ProjectID        core.ProjectID
	AssetID          core.AssetID
	AssetAllocations []AssetAllocationInput
	ResourceType     investment.ResourceType
	Amount           *money.Money
	TimeHours        *float64
	TimeHourlyRate   *money.Money
	GoodsQuantity    *float64
	GoodsUnit        string
	GoodsItemName    string
	Description      string
	Date             time.Time
	Tags             []string
}

type UpdateInvestmentHandler struct {
	repo      investment.Repository
	assets    AssetReader
	projects  ProjectReader
	converter investment.CurrencyConverter
	logger    core.Logger
}

func NewUpdateInvestmentHandler(
	repo investment.Repository,
	assets AssetReader,
	projects ProjectReader,
	converter investment.CurrencyConverter,
	logger core.Logger,
) UpdateInvestmentHandler {
	if repo == nil {
		panic("missing investment repository")
	}
	if assets == nil {
		panic("missing asset reader")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return UpdateInvestmentHandler{
		repo:      repo,
		assets:    assets,
		projects:  projects,
		converter: converter,
		logger:    logger,
	}
}

func (h UpdateInvestmentHandler) Handle(ctx context.Context, cmd UpdateInvestment) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "UpdateInvestment", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException(failedToUpdateInvestment, nil)
	}

	var proj ProjectDescriptor
	if !cmd.ProjectID.IsNil() {
		proj, err = verifyProjectAccess(ctx, h.projects, cmd.ProjectID, cmd.RequesterID)
		if err != nil {
			return err
		}
	}

	err = h.repo.UpdateInvestment(
		ctx,
		cmd.ID,
		cmd.ProjectID,
		func(txCtx context.Context, inv *investment.Investment) (*investment.Investment, error) {
			if proj.MainCurrency == "" && !inv.ProjectID().IsNil() {
				p, err := verifyProjectAccess(txCtx, h.projects, inv.ProjectID(), cmd.RequesterID)
				if err != nil {
					return nil, err
				}
				proj = p
			}

			if len(cmd.AssetAllocations) > 0 || (!cmd.AssetID.IsNil() && cmd.AssetID != inv.PrimaryAssetID()) {
				targetProjID := cmd.ProjectID
				if targetProjID.IsNil() {
					targetProjID = inv.ProjectID()
				}
				allocations, _, _, err := resolveAllocations(txCtx, h.assets, targetProjID, cmd.AssetID, cmd.AssetAllocations)
				if err != nil {
					return nil, err
				}
				if err := inv.Allocate(allocations); err != nil {
					return nil, err
				}
			}

			inv.UpdateDetails(cmd.Description, cmd.Date, cmd.Tags)
			if err := inv.UpdateValuation(cmd.ResourceType, cmd.Amount, cmd.TimeHours, cmd.TimeHourlyRate, cmd.GoodsQuantity, cmd.GoodsUnit, cmd.GoodsItemName); err != nil {
				return nil, err
			}
			inv.EvaluateHomeAmount(h.converter, proj.MainCurrency)
			return inv, nil
		},
	)
	return preserveOrWrapError(err, failedToUpdateInvestment)
}
