package app

import (
	"context"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type CreateAsset struct {
	AssetID              core.AssetID
	RequesterID          core.PersonID
	Name                 string
	Description          string
	ProjectID            core.ProjectID
	TypeID               uuid.UUID
	Status               asset.AssetStatus
	StartDate            time.Time
	CompletedDate        *time.Time
	TargetCompletionDate *time.Time
	TargetPrice          *money.Money
	Price                *money.Money
	AcquiredAt           time.Time
	WithPayment          bool
	Tags                 []string
}

type CreateAssetHandler struct {
	db          core.DbConnection
	repo        asset.Repository
	owners      OwnerReader
	projects    ProjectReader
	investments InvestmentSource
	converter   asset.CurrencyConverter
	logger      core.Logger
}

func NewCreateAssetHandler(
	db core.DbConnection,
	repo asset.Repository,
	owners OwnerReader,
	projects ProjectReader,
	investments InvestmentSource,
	converter asset.CurrencyConverter,
	logger core.Logger,
) CreateAssetHandler {
	if db == nil {
		panic("missing db connection")
	}
	if repo == nil {
		panic("missing asset repository")
	}
	if owners == nil {
		panic("missing owner reader")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return CreateAssetHandler{
		db:          db,
		repo:        repo,
		owners:      owners,
		projects:    projects,
		investments: investments,
		converter:   converter,
		logger:      logger,
	}
}

func (h CreateAssetHandler) Handle(ctx context.Context, cmd CreateAsset) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "CreateAsset", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException(failedToCreateAsset, nil)
	}

	owner, err := h.owners.FindOwner(ctx, cmd.RequesterID)
	if err != nil {
		return exceptions.NewInternalException(failedToCreateAsset, err)
	}

	proj, err := h.projects.FindProject(ctx, cmd.ProjectID, cmd.RequesterID)
	if err != nil {
		return preserveOrWrapError(err, failedToCreateAsset)
	}

	price := cmd.TargetPrice
	if price == nil {
		price = cmd.Price
	}

	startDate := cmd.StartDate
	if startDate.IsZero() {
		startDate = core.DateOrNow(cmd.AcquiredAt)
	}

	assetID := cmd.AssetID
	if assetID.IsNil() {
		assetID = core.NewAssetID()
	}

	anAsset, err := asset.CreateAsset(
		assetID,
		cmd.Name,
		cmd.Description,
		cmd.ProjectID,
		price,
		startDate,
		owner,
		cmd.Status,
		cmd.CompletedDate,
		cmd.Tags,
	)
	if err != nil {
		return err
	}

	if cmd.TargetCompletionDate != nil {
		if err := anAsset.UpdateSchedule(nil, cmd.TargetCompletionDate, false); err != nil {
			return err
		}
	}

	if cmd.WithPayment && (price == nil || price.Amount() <= 0) {
		return exceptions.NewValidationException("price must be greater than 0 when creating an asset with payment", nil)
	}

	homeCurrency := "UAH"
	if proj.MainCurrency != "" {
		homeCurrency = proj.MainCurrency
	}
	anAsset.EvaluateValuation(homeCurrency, h.converter)

	paymentDescription := cmd.Description
	if paymentDescription == "" {
		paymentDescription = cmd.Name
	}

	if cmd.WithPayment {
		_, err = h.db.Tx(ctx, func(txCtx context.Context) (any, error) {
			if err = h.repo.Save(txCtx, anAsset); err != nil {
				return nil, err
			}
			if h.investments != nil {
				if err = h.investments.CreateInitialInvestment(txCtx, asset.InitialInvestment{
					ID:            core.NewInvestmentID(),
					ProjectID:     cmd.ProjectID,
					AssetID:       anAsset.ID(),
					ContributorID: owner.PersonID,
					Amount:        price,
					Date:          startDate,
					Description:   paymentDescription,
					Tags:          cmd.Tags,
				}); err != nil {
					return nil, err
				}
			}
			return nil, nil
		})
		return preserveOrWrapError(err, failedToCreateAsset)
	}

	return preserveOrWrapError(h.repo.Save(ctx, anAsset), failedToCreateAsset)
}
