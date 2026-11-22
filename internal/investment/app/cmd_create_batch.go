package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
)

type BatchCreateInvestmentItem struct {
	ID             core.InvestmentID
	AssetID        core.AssetID
	ResourceType   investment.ResourceType
	Amount         *money.Money
	TimeHours      *float64
	TimeHourlyRate *money.Money
	GoodsQuantity  *float64
	GoodsUnit      string
	GoodsItemName  string
	Description    string
	Date           time.Time
	Tags           []string
}

type CreateBatchInvestments struct {
	RequesterID  core.PersonID
	ProjectID    core.ProjectID
	DefaultAsset core.AssetID
	Items        []BatchCreateInvestmentItem
}

type CreateBatchInvestmentsHandler struct {
	db           core.DbConnection
	repo         investment.Repository
	assets       AssetReader
	projects     ProjectReader
	contributors ContributorReader
	converter    investment.CurrencyConverter
	logger       core.Logger
}

func NewCreateBatchInvestmentsHandler(
	db core.DbConnection,
	repo investment.Repository,
	assets AssetReader,
	projects ProjectReader,
	contributors ContributorReader,
	converter investment.CurrencyConverter,
	logger core.Logger,
) CreateBatchInvestmentsHandler {
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
	if contributors == nil {
		panic("missing contributor reader")
	}
	return CreateBatchInvestmentsHandler{
		db:           db,
		repo:         repo,
		assets:       assets,
		projects:     projects,
		contributors: contributors,
		converter:    converter,
		logger:       logger,
	}
}

func (h CreateBatchInvestmentsHandler) Handle(ctx context.Context, cmd CreateBatchInvestments) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "CreateBatchInvestments", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException(failedToCreateInvestment, nil)
	}
	if cmd.ProjectID.IsNil() {
		return exceptions.NewValidationException("project id is required", nil)
	}
	if len(cmd.Items) == 0 {
		return exceptions.NewValidationException("at least one investment item is required", nil)
	}

	contributor, err := h.contributors.FindContributor(ctx, cmd.RequesterID)
	if err != nil {
		return exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	proj, err := verifyProjectAccess(ctx, h.projects, cmd.ProjectID, cmd.RequesterID)
	if err != nil {
		return err
	}

	_, err = h.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		for i, item := range cmd.Items {
			targetAssetID := item.AssetID
			if targetAssetID.IsNil() {
				targetAssetID = cmd.DefaultAsset
			}
			if targetAssetID.IsNil() {
				return nil, exceptions.NewValidationException(fmt.Sprintf("item %d requires an asset", i), nil)
			}

			ast, err := h.assets.FindAsset(txCtx, cmd.ProjectID, targetAssetID)
			if err != nil {
				return nil, exceptions.NewValidationException(fmt.Sprintf("asset %s not found in project", targetAssetID), err)
			}

			resType := item.ResourceType
			if resType == "" {
				resType = investment.ResourceTypeMoney
			}

			invID := item.ID
			if invID.IsNil() {
				invID = core.NewInvestmentID()
			}

			inv, err := investment.NewInvestment(
				invID,
				proj.ID,
				proj.MainCurrency,
				ast.ID,
				ast.Name,
				contributor,
				resType,
				item.Amount,
				item.TimeHours,
				item.TimeHourlyRate,
				item.GoodsQuantity,
				item.GoodsUnit,
				item.GoodsItemName,
				item.Description,
				core.DateOrNow(item.Date),
				item.Tags,
			)
			if err != nil {
				return nil, err
			}
			inv.EvaluateHomeAmount(h.converter)

			if err := h.repo.Save(txCtx, inv); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return preserveOrWrapError(err, failedToCreateInvestment)
}
