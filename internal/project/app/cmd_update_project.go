package app

import (
	"context"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

type UpdateProject struct {
	ProjectID    core.ProjectID
	RequesterID  core.PersonID
	Name         string
	Description  string
	MainCurrency string
	StartDate    *time.Time
	EndDate      *time.Time
	TagBudgets   *[]project.TagBudgetInput
}

type UpdateProjectHandler struct {
	repo      project.ProjectRepository
	converter project.CurrencyConverter
	logger    core.Logger
}

func NewUpdateProjectHandler(repo project.ProjectRepository, converter project.CurrencyConverter, logger core.Logger) UpdateProjectHandler {
	if repo == nil {
		panic("missing project repository")
	}
	return UpdateProjectHandler{repo: repo, converter: converter, logger: logger}
}

func (h UpdateProjectHandler) Handle(ctx context.Context, cmd UpdateProject) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "UpdateProject", start, err)
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
			if err := p.UpdateDetails(cmd.Name, cmd.Description, cmd.MainCurrency); err != nil {
				return nil, err
			}
			if cmd.StartDate != nil || cmd.EndDate != nil {
				st := p.StartDate()
				if cmd.StartDate != nil {
					st = *cmd.StartDate
				}
				et := p.EndDate()
				if cmd.EndDate != nil {
					et = *cmd.EndDate
				}
				if err := p.SetDates(st, et); err != nil {
					return nil, err
				}
			}
			if cmd.TagBudgets != nil {
				if err := p.SetTagBudgets(*cmd.TagBudgets); err != nil {
					return nil, err
				}
			}
			p.EvaluateTagBudgets(h.converter)
			return p, nil
		},
	)
	return preserveOrWrapError(err, "failed to update project")
}
