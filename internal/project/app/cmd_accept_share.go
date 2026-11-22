package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

type AcceptProjectShare struct {
	ShareToken  uuid.UUID
	RequesterID core.PersonID
}

type AcceptProjectShareHandler struct {
	repo   project.ProjectRepository
	logger core.Logger
}

func NewAcceptProjectShareHandler(repo project.ProjectRepository, logger core.Logger) AcceptProjectShareHandler {
	if repo == nil {
		panic("missing project repository")
	}
	return AcceptProjectShareHandler{repo: repo, logger: logger}
}

func (h AcceptProjectShareHandler) Handle(ctx context.Context, cmd AcceptProjectShare) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "AcceptProjectShare", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}
	if cmd.ShareToken == uuid.Nil {
		return exceptions.NewValidationException("invalid share token", nil)
	}

	proj, err := h.repo.FindByShareToken(ctx, cmd.ShareToken)
	if err != nil {
		return preserveOrWrapError(err, "project not found")
	}
	if proj.IsOwnedByPerson(cmd.RequesterID) {
		return nil
	}

	if err := h.repo.CreateShareRecord(ctx, proj.ID(), cmd.RequesterID); err != nil {
		return preserveOrWrapError(err, "failed to record project share")
	}
	return nil
}
