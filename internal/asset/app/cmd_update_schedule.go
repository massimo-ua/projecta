package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type RescheduleAsset struct {
	RequesterID          core.PersonID
	AssetID              core.AssetID
	ProjectID            core.ProjectID
	StartDate            *time.Time
	TargetCompletionDate *time.Time
	ClearTargetDate      bool
}

type RescheduleAssetHandler struct {
	repo     asset.Repository
	projects ProjectReader
	logger   core.Logger
}

func NewRescheduleAssetHandler(
	repo asset.Repository,
	projects ProjectReader,
	logger core.Logger,
) RescheduleAssetHandler {
	if repo == nil {
		panic("missing asset repository")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return RescheduleAssetHandler{
		repo:     repo,
		projects: projects,
		logger:   logger,
	}
}

func (h RescheduleAssetHandler) Handle(ctx context.Context, cmd RescheduleAsset) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "RescheduleAsset", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException(failedToUpdateAsset, nil)
	}

	if _, err := h.projects.FindProject(ctx, cmd.ProjectID, cmd.RequesterID); err != nil {
		return preserveOrWrapError(err, failedToUpdateAsset)
	}

	err = h.repo.UpdateAsset(
		ctx,
		cmd.AssetID,
		cmd.ProjectID,
		cmd.RequesterID,
		func(_ context.Context, anAsset *asset.Asset) (*asset.Asset, error) {
			if err := anAsset.UpdateSchedule(cmd.StartDate, cmd.TargetCompletionDate, cmd.ClearTargetDate); err != nil {
				return nil, err
			}
			return anAsset, nil
		},
	)
	return preserveOrWrapError(err, failedToUpdateAsset)
}
