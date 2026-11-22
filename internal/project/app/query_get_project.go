package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

type GetProject struct {
	ProjectID   core.ProjectID
	Name        string
	ShareToken  uuid.UUID
	RequesterID core.PersonID
}

type GetProjectHandler struct {
	repo      project.ProjectRepository
	converter project.CurrencyConverter
	logger    core.Logger
}

func NewGetProjectHandler(repo project.ProjectRepository, converter project.CurrencyConverter, logger core.Logger) GetProjectHandler {
	if repo == nil {
		panic("missing project repository")
	}
	return GetProjectHandler{
		repo:      repo,
		converter: converter,
		logger:    logger,
	}
}

func (h GetProjectHandler) Handle(ctx context.Context, q GetProject) (res *project.Project, err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "query", "GetProject", start, err)
	}()

	if q.RequesterID.IsNil() {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}

	var p *project.Project
	switch {
	case q.ShareToken != uuid.Nil:
		p, err = h.repo.FindByShareToken(ctx, q.ShareToken)
		if err != nil {
			return nil, preserveOrWrapError(err, "project not found")
		}
		if err := project.CanPersonViewProject(q.RequesterID, *p); err != nil {
			return nil, err
		}
		if !p.IsOwnedByPerson(q.RequesterID) {
			p.MarkShared()
		}
	case !q.ProjectID.IsNil() && q.Name == "":
		p, err = h.repo.GetProject(ctx, q.ProjectID, q.RequesterID)
	default:
		p, err = h.repo.FindOne(ctx, project.ProjectFilter{
			ProjectID:   q.ProjectID,
			Name:        q.Name,
			RequesterID: q.RequesterID,
		})
	}
	if err != nil {
		return nil, preserveOrWrapError(err, "project not found")
	}

	p.EvaluateTagBudgets(h.converter)
	return p, nil
}
