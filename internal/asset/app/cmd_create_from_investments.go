package app

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type CreateAssetFromInvestments struct {
	AssetID        core.AssetID
	RequesterID    core.PersonID
	ProjectID      core.ProjectID
	InvestmentIDs  []core.InvestmentID
	Name           string
	Description    string
	TypeID         uuid.UUID
	AcquiredAt     time.Time
	TargetCurrency string
}

type CreateAssetFromInvestmentsHandler struct {
	db          core.DbConnection
	repo        asset.Repository
	owners      OwnerReader
	projects    ProjectReader
	investments InvestmentSource
	converter   asset.CurrencyConverter
	logger      core.Logger
}

func NewCreateAssetFromInvestmentsHandler(
	db core.DbConnection,
	repo asset.Repository,
	owners OwnerReader,
	projects ProjectReader,
	investments InvestmentSource,
	converter asset.CurrencyConverter,
	logger core.Logger,
) CreateAssetFromInvestmentsHandler {
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
	return CreateAssetFromInvestmentsHandler{
		db:          db,
		repo:        repo,
		owners:      owners,
		projects:    projects,
		investments: investments,
		converter:   converter,
		logger:      logger,
	}
}

func (h CreateAssetFromInvestmentsHandler) Handle(ctx context.Context, cmd CreateAssetFromInvestments) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "CreateAssetFromInvestments", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException(failedToCreateAsset, nil)
	}
	if cmd.ProjectID.IsNil() {
		return exceptions.NewValidationException("project id is required", nil)
	}
	if len(cmd.InvestmentIDs) == 0 {
		return exceptions.NewValidationException("at least one investment must be selected", nil)
	}
	if strings.TrimSpace(cmd.Name) == "" {
		return exceptions.NewValidationException("name is required", nil)
	}

	owner, err := h.owners.FindOwner(ctx, cmd.RequesterID)
	if err != nil {
		return exceptions.NewInternalException(failedToCreateAsset, err)
	}

	proj, err := h.projects.FindProject(ctx, cmd.ProjectID, cmd.RequesterID)
	if err != nil {
		return preserveOrWrapError(err, failedToCreateAsset)
	}

	if h.investments == nil {
		return exceptions.NewInternalException("investment source not configured", nil)
	}

	invItems, err := h.investments.FindInvestments(ctx, cmd.ProjectID, cmd.RequesterID, cmd.InvestmentIDs)
	if err != nil {
		return preserveOrWrapError(err, failedToCreateAsset)
	}
	if len(invItems) != len(cmd.InvestmentIDs) {
		return exceptions.NewValidationException("one or more investments not found in project", nil)
	}

	mainCurrency := "UAH"
	if proj.MainCurrency != "" {
		mainCurrency = proj.MainCurrency
	}

	assetID := cmd.AssetID
	if assetID.IsNil() {
		assetID = core.NewAssetID()
	}

	anAsset, err := asset.CreateAssetFromInvestments(
		assetID,
		cmd.Name,
		cmd.Description,
		cmd.ProjectID,
		mainCurrency,
		cmd.TargetCurrency,
		cmd.AcquiredAt,
		owner,
		invItems,
		h.converter,
	)
	if err != nil {
		return err
	}

	_, err = h.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		if err = h.repo.Save(txCtx, anAsset); err != nil {
			return nil, err
		}
		if err = h.investments.AssignToAsset(txCtx, anAsset.ID(), cmd.RequesterID, cmd.InvestmentIDs); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return preserveOrWrapError(err, failedToCreateAsset)
}
