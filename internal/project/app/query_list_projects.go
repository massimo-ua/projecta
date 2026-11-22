package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

type ListProjects struct {
	Pagination  core.Pagination
	Name        string
	RequesterID core.PersonID
}

type ListProjectsHandler struct {
	repo      project.ProjectRepository
	converter project.CurrencyConverter
	logger    core.Logger
}

func NewListProjectsHandler(repo project.ProjectRepository, converter project.CurrencyConverter, logger core.Logger) ListProjectsHandler {
	if repo == nil {
		panic("missing project repository")
	}
	return ListProjectsHandler{
		repo:      repo,
		converter: converter,
		logger:    logger,
	}
}

func (h ListProjectsHandler) Handle(ctx context.Context, q ListProjects) (res []*project.Project, err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "query", "ListProjects", start, err)
	}()

	if q.RequesterID.IsNil() {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}

	projects, err := h.repo.Find(ctx, project.ProjectCollectionFilter{
		Pagination:  q.Pagination,
		Name:        q.Name,
		RequesterID: q.RequesterID,
	})
	if err != nil {
		return nil, preserveOrWrapError(err, "failed to list projects")
	}
	for _, p := range projects {
		p.EvaluateTagBudgets(h.converter)
	}
	return projects, nil
}
