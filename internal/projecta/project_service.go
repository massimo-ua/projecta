package projecta

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type ProjectServiceImpl struct {
	repository    ProjectRepository
	peopleService PeopleService
}

func (s *ProjectServiceImpl) Save(ctx context.Context, expense *Payment) error {
	panic("implement me")
}

func (s *ProjectServiceImpl) FindOne(ctx context.Context, filter ProjectFilter) (*Project, error) {
	return s.repository.FindOne(ctx, filter)
}

func (s *ProjectServiceImpl) Find(ctx context.Context, filter ProjectCollectionFilter) ([]*Project, error) {
	return s.repository.Find(ctx, filter)
}

func (s *ProjectServiceImpl) Remove(ctx context.Context, command RemoveProjectCommand) error {
	p, err := s.repository.FindOne(ctx, ProjectFilter{ProjectID: command.ProjectID})
	if err != nil {
		return err
	}
	return s.repository.Remove(ctx, p)
}

func (s *ProjectServiceImpl) Update(ctx context.Context, command UpdateProjectCommand) error {
	p, err := s.repository.FindOne(ctx, ProjectFilter{ProjectID: command.ProjectID})
	if err != nil {
		return err
	}

	if command.Name != "" {
		p.Name = command.Name
	}
	if command.Description != "" {
		p.Description = command.Description
	}
	if command.MainCurrency != "" {
		p.MainCurrency = command.MainCurrency
	}

	return s.repository.Update(ctx, p)
}

func NewProjectService(repository ProjectRepository, peopleService PeopleService) *ProjectServiceImpl {
	return &ProjectServiceImpl{repository: repository, peopleService: peopleService}
}

func (s *ProjectServiceImpl) Create(ctx context.Context, command CreateProjectCommand) (*Project, error) {
	owner, err := s.peopleService.FindOwner(ctx, command.PersonID)

	if err != nil {
		return nil, exceptions.NewInternalException("failed to find owner", err)
	}

	p, err := s.repository.FindOne(ctx, ProjectFilter{Name: command.Name})

	if p != nil {
		return p, nil
	}

	if errors.Is(err, exceptions.NotFoundError) {
		p, err = NewProject(
			uuid.New(),
			command.Name,
			command.Description,
			owner,
			time.Time{},
			time.Time{},
		)
		if err != nil {
			return nil, exceptions.NewInternalException("failed to create project", err)
		}

		if command.MainCurrency != "" {
			p.MainCurrency = command.MainCurrency
		}

		err = s.repository.Create(ctx, p)

		if err != nil {
			return nil, exceptions.NewInternalException("failed to save project", err)
		}

		return p, nil
	}

	return nil, exceptions.NewInternalException("failed to create project", err)
}
