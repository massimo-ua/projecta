package app

import (
	"context"
	"errors"
	"strings"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

type CreateProject struct {
	ProjectID    core.ProjectID
	RequesterID  core.PersonID
	Name         string
	Description  string
	MainCurrency string
	StartDate    time.Time
	EndDate      *time.Time
	TagBudgets   []project.TagBudgetInput
}

type CreateProjectHandler struct {
	repo        project.ProjectRepository
	ownerReader OwnerReader
	converter   project.CurrencyConverter
	logger      core.Logger
}

func NewCreateProjectHandler(repo project.ProjectRepository, ownerReader OwnerReader, converter project.CurrencyConverter, logger core.Logger) CreateProjectHandler {
	if repo == nil {
		panic("missing project repository")
	}
	if ownerReader == nil {
		panic("missing owner reader")
	}
	return CreateProjectHandler{
		repo:        repo,
		ownerReader: ownerReader,
		converter:   converter,
		logger:      logger,
	}
}

func (h CreateProjectHandler) Handle(ctx context.Context, cmd CreateProject) (err error) {
	start := time.Now()
	defer func() {
		logExecution(h.logger, "command", "CreateProject", start, err)
	}()

	if cmd.RequesterID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}

	trimmedName := strings.TrimSpace(cmd.Name)
	if trimmedName == "" {
		return exceptions.NewValidationException("project name is required", nil)
	}

	owner, err := h.ownerReader.FindOwner(ctx, cmd.RequesterID)
	if err != nil {
		return exceptions.NewInternalException("failed to find owner", err)
	}
	if !owner.CanHaveProjects {
		return exceptions.NewForbiddenException("owner cannot have projects", nil)
	}

	existing, err := h.repo.FindOne(ctx, project.ProjectFilter{
		Name:        trimmedName,
		RequesterID: cmd.RequesterID,
	})
	if existing != nil && existing.IsOwnedByPerson(cmd.RequesterID) {
		// Preserve idempotent creation by name for the same owner.
		return nil
	}
	if err != nil && !errors.Is(err, exceptions.NotFoundError) && existing == nil {
		return exceptions.NewInternalException("failed to create project", err)
	}

	startDate := cmd.StartDate
	if startDate.IsZero() {
		startDate = time.Now().UTC()
	}
	var endDate time.Time
	if cmd.EndDate != nil {
		endDate = *cmd.EndDate
	}

	projectID := cmd.ProjectID
	if projectID.IsNil() {
		projectID = core.NewProjectID()
	}

	p, err := project.NewProject(
		projectID,
		cmd.Name,
		cmd.Description,
		owner,
		startDate,
		endDate,
		cmd.MainCurrency,
	)
	if err != nil {
		return err
	}

	if len(cmd.TagBudgets) > 0 {
		if err := p.SetTagBudgets(cmd.TagBudgets); err != nil {
			return err
		}
		p.EvaluateTagBudgets(h.converter)
	}

	if err := h.repo.Create(ctx, p); err != nil {
		return preserveOrWrapError(err, "failed to save project")
	}
	return nil
}
