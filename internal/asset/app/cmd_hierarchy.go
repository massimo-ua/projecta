package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type LinkChildAsset struct {
	RequesterID     core.PersonID
	ParentID        core.AssetID
	ChildID         core.AssetID
	ProjectID       core.ProjectID
	SharePercentage float64
}

type LinkChildAssetHandler struct {
	repo     asset.Repository
	projects ProjectReader
	logger   core.Logger
}

func NewLinkChildAssetHandler(repo asset.Repository, projects ProjectReader, logger core.Logger) LinkChildAssetHandler {
	if repo == nil {
		panic("missing asset repository")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return LinkChildAssetHandler{repo: repo, projects: projects, logger: logger}
}

func (h LinkChildAssetHandler) Handle(ctx context.Context, cmd LinkChildAsset) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "LinkChildAsset", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to link child asset", nil)
	}
	if !cmd.ProjectID.IsNil() {
		if _, err := h.projects.FindProject(ctx, cmd.ProjectID, cmd.RequesterID); err != nil {
			return preserveOrWrapError(err, "failed to link child asset")
		}
	}
	if err := asset.ValidateEdge(cmd.ParentID, cmd.ChildID, cmd.SharePercentage, nil); err != nil {
		return err
	}
	if _, err := h.repo.FindOne(ctx, asset.Filter{ID: cmd.ParentID, ProjectID: cmd.ProjectID, OwnerID: cmd.RequesterID}); err != nil {
		return exceptions.NewValidationException("parent asset not found", err)
	}
	if _, err := h.repo.FindOne(ctx, asset.Filter{ID: cmd.ChildID, ProjectID: cmd.ProjectID, OwnerID: cmd.RequesterID}); err != nil {
		return exceptions.NewValidationException("child asset not found", err)
	}

	ancestors, err := h.repo.FindAncestors(ctx, cmd.ParentID)
	if err != nil {
		return exceptions.NewInternalException("failed to check hierarchy cycles", err)
	}
	if err := asset.ValidateEdge(cmd.ParentID, cmd.ChildID, cmd.SharePercentage, ancestors); err != nil {
		return err
	}

	return preserveOrWrapError(h.repo.AddChild(ctx, cmd.ParentID, cmd.ChildID, cmd.SharePercentage), "failed to link child asset")
}

type UnlinkChildAsset struct {
	RequesterID core.PersonID
	ParentID    core.AssetID
	ChildID     core.AssetID
	ProjectID   core.ProjectID
}

type UnlinkChildAssetHandler struct {
	repo     asset.Repository
	projects ProjectReader
	logger   core.Logger
}

func NewUnlinkChildAssetHandler(repo asset.Repository, projects ProjectReader, logger core.Logger) UnlinkChildAssetHandler {
	if repo == nil {
		panic("missing asset repository")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return UnlinkChildAssetHandler{repo: repo, projects: projects, logger: logger}
}

func (h UnlinkChildAssetHandler) Handle(ctx context.Context, cmd UnlinkChildAsset) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "UnlinkChildAsset", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to unlink child asset", nil)
	}
	if !cmd.ProjectID.IsNil() {
		if _, err := h.projects.FindProject(ctx, cmd.ProjectID, cmd.RequesterID); err != nil {
			return preserveOrWrapError(err, "failed to unlink child asset")
		}
	}
	if _, err := h.repo.FindOne(ctx, asset.Filter{
		ID:        cmd.ParentID,
		ProjectID: cmd.ProjectID,
		OwnerID:   cmd.RequesterID,
	}); err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return exceptions.NewNotFoundException("parent asset not found", err)
		}
		return preserveOrWrapError(err, "failed to unlink child asset")
	}

	return preserveOrWrapError(h.repo.RemoveChild(ctx, cmd.ParentID, cmd.ChildID), "failed to unlink child asset")
}

type LinkChildrenAssets struct {
	RequesterID   core.PersonID
	ProjectID     core.ProjectID
	ParentID      core.AssetID
	ChildAssetIDs []core.AssetID
}

type LinkChildrenAssetsHandler struct {
	db        core.DbConnection
	linkChild LinkChildAssetHandler
	logger    core.Logger
}

func NewLinkChildrenAssetsHandler(db core.DbConnection, linkChild LinkChildAssetHandler, logger core.Logger) LinkChildrenAssetsHandler {
	if db == nil {
		panic("missing db connection")
	}
	return LinkChildrenAssetsHandler{db: db, linkChild: linkChild, logger: logger}
}

func (h LinkChildrenAssetsHandler) Handle(ctx context.Context, cmd LinkChildrenAssets) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "LinkChildrenAssets", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to link child assets", nil)
	}

	_, err = h.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		for _, childID := range cmd.ChildAssetIDs {
			if err := h.linkChild.Handle(txCtx, LinkChildAsset{
				RequesterID:     cmd.RequesterID,
				ParentID:        cmd.ParentID,
				ChildID:         childID,
				ProjectID:       cmd.ProjectID,
				SharePercentage: 100.0,
			}); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return preserveOrWrapError(err, "failed to link child assets")
}

type GroupAssets struct {
	GroupAssetID  core.AssetID
	RequesterID   core.PersonID
	ProjectID     core.ProjectID
	Name          string
	Description   string
	ChildAssetIDs []core.AssetID
	Tags          []string
}

type GroupAssetsHandler struct {
	db       core.DbConnection
	repo     asset.Repository
	owners   OwnerReader
	projects ProjectReader
	logger   core.Logger
}

func NewGroupAssetsHandler(
	db core.DbConnection,
	repo asset.Repository,
	owners OwnerReader,
	projects ProjectReader,
	logger core.Logger,
) GroupAssetsHandler {
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
	return GroupAssetsHandler{
		db:       db,
		repo:     repo,
		owners:   owners,
		projects: projects,
		logger:   logger,
	}
}

func (h GroupAssetsHandler) Handle(ctx context.Context, cmd GroupAssets) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "GroupAssets", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to group assets", nil)
	}
	if cmd.ProjectID.IsNil() {
		return exceptions.NewValidationException("project id is required", nil)
	}
	if strings.TrimSpace(cmd.Name) == "" {
		return exceptions.NewValidationException("name is required", nil)
	}
	if len(cmd.ChildAssetIDs) == 0 {
		return exceptions.NewValidationException("at least one asset must be selected to group", nil)
	}

	owner, err := h.owners.FindOwner(ctx, cmd.RequesterID)
	if err != nil {
		return exceptions.NewInternalException("failed to group assets", err)
	}

	if _, err := h.projects.FindProject(ctx, cmd.ProjectID, cmd.RequesterID); err != nil {
		return preserveOrWrapError(err, "failed to group assets")
	}

	for _, childID := range cmd.ChildAssetIDs {
		child, err := h.repo.FindOne(ctx, asset.Filter{ID: childID, ProjectID: cmd.ProjectID, OwnerID: cmd.RequesterID})
		if err != nil || child == nil {
			return exceptions.NewValidationException(fmt.Sprintf("child asset %s not found in project", childID), err)
		}
	}

	parentID := cmd.GroupAssetID
	if parentID.IsNil() {
		parentID = core.NewAssetID()
	}

	parentAsset, err := asset.CreateAssetGroup(
		parentID,
		cmd.Name,
		cmd.Description,
		cmd.ProjectID,
		owner,
		cmd.Tags,
	)
	if err != nil {
		return err
	}

	_, err = h.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		if err := h.repo.Save(txCtx, parentAsset); err != nil {
			return nil, err
		}
		for _, childID := range cmd.ChildAssetIDs {
			if err := h.repo.AddChild(txCtx, parentID, childID, 100.0); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return preserveOrWrapError(err, "failed to group assets")
}
