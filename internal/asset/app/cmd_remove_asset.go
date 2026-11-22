package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type RemoveAsset struct {
	AssetID     core.AssetID
	ProjectID   core.ProjectID
	RequesterID core.PersonID
}

type RemoveAssetHandler struct {
	db       core.DbConnection
	repo     asset.Repository
	projects ProjectReader
	logger   core.Logger
}

func NewRemoveAssetHandler(db core.DbConnection, repo asset.Repository, projects ProjectReader, logger core.Logger) RemoveAssetHandler {
	if db == nil {
		panic("missing db connection")
	}
	if repo == nil {
		panic("missing asset repository")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return RemoveAssetHandler{db: db, repo: repo, projects: projects, logger: logger}
}

func (h RemoveAssetHandler) Handle(ctx context.Context, cmd RemoveAsset) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "RemoveAsset", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException(failedToFindAsset, nil)
	}
	if !cmd.ProjectID.IsNil() {
		if _, err := h.projects.FindProject(ctx, cmd.ProjectID, cmd.RequesterID); err != nil {
			return preserveOrWrapError(err, failedToFindAsset)
		}
	}

	anAsset, err := h.repo.FindOne(ctx, asset.Filter{
		ID:        cmd.AssetID,
		ProjectID: cmd.ProjectID,
		OwnerID:   cmd.RequesterID,
	})
	if err != nil {
		return preserveOrWrapError(err, failedToFindAsset)
	}
	if cmd.ProjectID.IsNil() && !anAsset.ProjectID().IsNil() {
		if _, err := h.projects.FindProject(ctx, anAsset.ProjectID(), cmd.RequesterID); err != nil {
			return preserveOrWrapError(err, failedToFindAsset)
		}
	}

	_, err = h.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		if err := h.repo.Remove(txCtx, anAsset); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return preserveOrWrapError(err, "failed to remove asset")
}
