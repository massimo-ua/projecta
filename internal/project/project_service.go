package project

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type ProjectServiceImpl struct {
	repository       ProjectRepository
	peopleService    PeopleService
	totalsCalculator TotalsCalculator
}

func (s *ProjectServiceImpl) FindOne(ctx context.Context, filter ProjectFilter) (*Project, error) {
	if filter.RequesterID.IsNil() {
		personID, err := core.AuthGuard(ctx)
		if err != nil {
			return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
		}
		filter.RequesterID = personID
	}
	return s.repository.FindOne(ctx, filter)
}

func (s *ProjectServiceImpl) Find(ctx context.Context, filter ProjectCollectionFilter) ([]*Project, error) {
	if filter.RequesterID.IsNil() {
		personID, err := core.AuthGuard(ctx)
		if err != nil {
			return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
		}
		filter.RequesterID = personID
	}
	return s.repository.Find(ctx, filter)
}

func resolveRequesterID(ctx context.Context, explicitID core.PersonID) (core.PersonID, error) {
	if !explicitID.IsNil() {
		return explicitID, nil
	}
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return core.NilPersonID, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}
	return personID, nil
}

func (s *ProjectServiceImpl) Remove(ctx context.Context, command RemoveProjectCommand) error {
	personID, err := resolveRequesterID(ctx, command.PersonID)
	if err != nil {
		return err
	}
	p, err := s.FindOne(ctx, ProjectFilter{ProjectID: command.ProjectID, RequesterID: personID})
	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return exceptions.NewNotFoundException("project not found", err)
		}
		return exceptions.NewInternalException("failed to find project", err)
	}

	if !p.IsOwnedByPerson(personID) {
		return exceptions.NewForbiddenException("only project owner can remove project", nil)
	}

	if err := s.repository.Remove(ctx, p); err != nil {
		return exceptions.NewInternalException("failed to remove project", err)
	}

	return nil
}

func (s *ProjectServiceImpl) Update(ctx context.Context, command UpdateProjectCommand) (*Project, error) {
	personID, err := resolveRequesterID(ctx, command.PersonID)
	if err != nil {
		return nil, err
	}
	p, err := s.repository.FindOne(ctx, ProjectFilter{ProjectID: command.ProjectID, RequesterID: personID})
	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return nil, exceptions.NewNotFoundException("project not found", err)
		}
		return nil, exceptions.NewInternalException("failed to find project", err)
	}

	if !p.IsOwnedByPerson(personID) {
		return nil, exceptions.NewForbiddenException("only project owner can update project", nil)
	}

	if err := p.UpdateDetails(command.Name, command.Description, command.MainCurrency); err != nil {
		return nil, err
	}

	err = s.repository.Update(ctx, p)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to update project", err)
	}

	return p, nil
}

func (s *ProjectServiceImpl) GetTotals(ctx context.Context, projectID core.ProjectID) (ProjectTotals, error) {
	proj, err := s.FindOne(ctx, ProjectFilter{ProjectID: projectID})
	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return ProjectTotals{}, exceptions.NewNotFoundException("project not found", err)
		}
		return ProjectTotals{}, err
	}

	homeCurrency := proj.MainCurrency()
	if homeCurrency == "" {
		homeCurrency = "UAH"
	}

	if s.totalsCalculator == nil {
		return ProjectTotals{Currency: homeCurrency}, nil
	}

	return s.totalsCalculator.CalculateTotals(ctx, projectID, homeCurrency)
}

func NewProjectService(repository ProjectRepository, peopleService PeopleService, totalsCalculator ...TotalsCalculator) *ProjectServiceImpl {
	var tc TotalsCalculator
	if len(totalsCalculator) > 0 {
		tc = totalsCalculator[0]
	}
	return &ProjectServiceImpl{repository: repository, peopleService: peopleService, totalsCalculator: tc}
}

func (s *ProjectServiceImpl) Create(ctx context.Context, command CreateProjectCommand) (*Project, error) {
	if strings.TrimSpace(command.Name) == "" {
		return nil, exceptions.NewValidationException("project name is required", nil)
	}

	owner, err := s.peopleService.FindOwner(ctx, command.PersonID)

	if err != nil {
		return nil, exceptions.NewInternalException("failed to find owner", err)
	}

	if !owner.CanHaveProjects {
		return nil, exceptions.NewForbiddenException("owner cannot have projects", nil)
	}

	p, err := s.repository.FindOne(ctx, ProjectFilter{
		Name:        command.Name,
		RequesterID: command.PersonID,
	})

	if p != nil {
		return p, nil
	}

	if errors.Is(err, exceptions.NotFoundError) {
		p, err = NewProject(
			core.NewProjectID(),
			command.Name,
			command.Description,
			owner,
			time.Time{},
			time.Time{},
			command.MainCurrency,
		)
		if err != nil {
			return nil, err
		}

		err = s.repository.Create(ctx, p)

		if err != nil {
			return nil, exceptions.NewInternalException("failed to save project", err)
		}

		return p, nil
	}

	return nil, exceptions.NewInternalException("failed to create project", err)
}

func (s *ProjectServiceImpl) AcceptShare(ctx context.Context, token uuid.UUID, personID core.PersonID) (*Project, error) {
	if token == uuid.Nil {
		return nil, exceptions.NewValidationException("invalid share token", nil)
	}

	project, err := s.repository.FindByShareToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if project.IsOwnedByPerson(personID) {
		return project, nil
	}

	err = s.repository.CreateShareRecord(ctx, project.ID(), personID)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to record project share", err)
	}

	project.MarkShared()
	return project, nil
}
