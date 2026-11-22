package app

import (
	"context"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
)

type CreateInvestment struct {
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

type CreateInvestmentHandler struct {
	db           core.DbConnection
	repo         investment.Repository
	assets       AssetReader
	projects     ProjectReader
	contributors ContributorReader
	converter    investment.CurrencyConverter
	logger       core.Logger
}

func NewCreateInvestmentHandler(
	db core.DbConnection,
	repo investment.Repository,
	assets AssetReader,
	projects ProjectReader,
	contributors ContributorReader,
	converter investment.CurrencyConverter,
	logger core.Logger,
) CreateInvestmentHandler {
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
	return CreateInvestmentHandler{
		db:           db,
		repo:         repo,
		assets:       assets,
		projects:     projects,
		contributors: contributors,
		converter:    converter,
		logger:       logger,
	}
}

func (h CreateInvestmentHandler) Handle(ctx context.Context, cmd CreateInvestment) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "CreateInvestment", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException(failedToCreateInvestment, nil)
	}
	if cmd.ProjectID.IsNil() {
		return exceptions.NewValidationException("project id is required", nil)
	}
	if cmd.AssetID.IsNil() && len(cmd.AssetAllocations) == 0 {
		return exceptions.NewValidationException("asset id is required", nil)
	}
	if strings.TrimSpace(cmd.Description) == "" {
		return exceptions.NewValidationException("description is required", nil)
	}

	contributor, err := h.contributors.FindContributor(ctx, cmd.RequesterID)
	if err != nil {
		return exceptions.NewInternalException(failedToCreateInvestment, err)
	}

	proj, err := verifyProjectAccess(ctx, h.projects, cmd.ProjectID, cmd.RequesterID)
	if err != nil {
		return err
	}

	allocations, primaryID, primaryName, err := resolveAllocations(ctx, h.assets, cmd.ProjectID, cmd.AssetID, cmd.AssetAllocations)
	if err != nil {
		return err
	}

	resType := cmd.ResourceType
	if resType == "" {
		resType = investment.ResourceTypeMoney
	}

	invID := cmd.ID
	if invID.IsNil() {
		invID = core.NewInvestmentID()
	}

	inv, err := investment.NewInvestment(
		invID,
		proj.ID,
		proj.MainCurrency,
		primaryID,
		primaryName,
		contributor,
		resType,
		cmd.Amount,
		cmd.TimeHours,
		cmd.TimeHourlyRate,
		cmd.GoodsQuantity,
		cmd.GoodsUnit,
		cmd.GoodsItemName,
		cmd.Description,
		cmd.Date,
		cmd.Tags,
		allocations...,
	)
	if err != nil {
		return err
	}

	inv.EvaluateHomeAmount(h.converter)

	_, err = h.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		if err := h.repo.Save(txCtx, inv); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return preserveOrWrapError(err, failedToCreateInvestment)
}
