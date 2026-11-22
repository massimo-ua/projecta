package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type ListProjectTags struct {
	ProjectID   core.ProjectID
	RequesterID core.PersonID
}

type ListProjectTagsHandler struct {
	tagsReadModel TagsReadModel
	projects      ProjectReader
	logger        core.Logger
}

func NewListProjectTagsHandler(
	tagsReadModel TagsReadModel,
	projects ProjectReader,
	logger core.Logger,
) ListProjectTagsHandler {
	if tagsReadModel == nil {
		panic("missing tags read model")
	}
	if projects == nil {
		panic("missing project reader")
	}
	return ListProjectTagsHandler{
		tagsReadModel: tagsReadModel,
		projects:      projects,
		logger:        logger,
	}
}

func (h ListProjectTagsHandler) Handle(ctx context.Context, q ListProjectTags) (res []string, err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "query", "ListProjectTags", start, err)
	}()

	if q.RequesterID.IsNil() {
		return nil, exceptions.NewUnauthorizedException("failed to find tags", nil)
	}
	if _, err := verifyProjectAccess(ctx, h.projects, q.ProjectID, q.RequesterID); err != nil {
		return nil, err
	}

	tags, err := h.tagsReadModel.FindTags(ctx, q.ProjectID)
	if err != nil {
		return nil, preserveOrWrapError(err, "failed to find tags")
	}
	return tags, nil
}
