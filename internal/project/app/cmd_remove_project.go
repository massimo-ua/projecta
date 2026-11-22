package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

type RemoveProject struct {
	ProjectID   core.ProjectID
	RequesterID core.PersonID
}

type RemoveProjectHandler struct {
	repo   project.ProjectRepository
	logger core.Logger
}

func NewRemoveProjectHandler(repo project.ProjectRepository, logger core.Logger) RemoveProjectHandler {
	if repo == nil {
		panic("missing project repository")
	}
	return RemoveProjectHandler{repo: repo, logger: logger}
}

func (h RemoveProjectHandler) Handle(ctx context.Context, cmd RemoveProject) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "RemoveProject", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}

	p, err := h.repo.GetProject(ctx, cmd.ProjectID, cmd.RequesterID)
	if err != nil {
		return preserveOrWrapError(err, "project not found")
	}

	if err := project.CanPersonManageProject(cmd.RequesterID, *p); err != nil {
		return err
	}

	if err := h.repo.Remove(ctx, p); err != nil {
		return preserveOrWrapError(err, "failed to remove project")
	}
	return nil
}
