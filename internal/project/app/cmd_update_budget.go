package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

type UpdateProjectBudget struct {
	ProjectID   core.ProjectID
	RequesterID core.PersonID
	TagBudgets  []project.TagBudgetInput
}

type UpdateProjectBudgetHandler struct {
	repo      project.ProjectRepository
	converter project.CurrencyConverter
	logger    core.Logger
}

func NewUpdateProjectBudgetHandler(repo project.ProjectRepository, converter project.CurrencyConverter, logger core.Logger) UpdateProjectBudgetHandler {
	if repo == nil {
		panic("missing project repository")
	}
	return UpdateProjectBudgetHandler{repo: repo, converter: converter, logger: logger}
}

func (h UpdateProjectBudgetHandler) Handle(ctx context.Context, cmd UpdateProjectBudget) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "UpdateProjectBudget", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}

	err = h.repo.UpdateProject(
		ctx,
		cmd.ProjectID,
		cmd.RequesterID,
		func(_ context.Context, p *project.Project) (*project.Project, error) {
			if err := project.CanPersonManageProject(cmd.RequesterID, *p); err != nil {
				return nil, err
			}
			if err := p.SetTagBudgets(cmd.TagBudgets); err != nil {
				return nil, err
			}
			p.EvaluateTagBudgets(h.converter)
			return p, nil
		},
	)
	return preserveOrWrapError(err, "failed to update project budget")
}
