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

type UpdateAsset struct {
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
	ClearTargetDate      bool
	TargetPrice          *money.Money
	Price                *money.Money
	AcquiredAt           time.Time
	Tags                 []string
}

type UpdateAssetHandler struct {
	repo      asset.Repository
	projects  ProjectReader
	converter asset.CurrencyConverter
	logger    core.Logger
}

func NewUpdateAssetHandler(
	repo asset.Repository,
	projects ProjectReader,
	converter asset.CurrencyConverter,
	logger core.Logger,
) UpdateAssetHandler {
	if repo == nil {
		panic("missing asset repository")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return UpdateAssetHandler{
		repo:      repo,
		projects:  projects,
		converter: converter,
		logger:    logger,
	}
}

func (h UpdateAssetHandler) Handle(ctx context.Context, cmd UpdateAsset) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "UpdateAsset", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException(failedToUpdateAsset, nil)
	}

	proj, err := h.projects.FindProject(ctx, cmd.ProjectID, cmd.RequesterID)
	if err != nil {
		return preserveOrWrapError(err, failedToUpdateAsset)
	}

	homeCurrency := "UAH"
	if proj.MainCurrency != "" {
		homeCurrency = proj.MainCurrency
	}

	err = h.repo.UpdateAsset(
		ctx,
		cmd.AssetID,
		cmd.ProjectID,
		cmd.RequesterID,
		func(_ context.Context, anAsset *asset.Asset) (*asset.Asset, error) {
			price := cmd.TargetPrice
			if price == nil {
				price = cmd.Price
			}
			if err := anAsset.UpdateDetails(cmd.Name, cmd.Description, price, cmd.Tags); err != nil {
				return nil, err
			}
			anAsset.TransitionStatus(cmd.Status, cmd.CompletedDate)

			var schedStart *time.Time
			if !cmd.StartDate.IsZero() {
				sDate := cmd.StartDate
				schedStart = &sDate
			} else if !cmd.AcquiredAt.IsZero() {
				aDate := cmd.AcquiredAt
				schedStart = &aDate
			}
			if err := anAsset.UpdateSchedule(schedStart, cmd.TargetCompletionDate, cmd.ClearTargetDate); err != nil {
				return nil, err
			}

			anAsset.EvaluateValuation(homeCurrency, h.converter)
			return anAsset, nil
		},
	)
	return preserveOrWrapError(err, failedToUpdateAsset)
}
