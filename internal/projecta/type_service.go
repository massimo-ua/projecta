package projecta

import (
	"context"
	"errors"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

const (
	failedToCreateCostType = "failed to create cost type"
)

type TypeServiceImpl struct {
	types      TypeRepository
	categories CategoryRepository
	projects   ProjectRepository
}

func (s *TypeServiceImpl) FindOne(ctx context.Context, filter TypeFilter) (*CostType, error) {
	return s.types.FindOne(ctx, filter)
}

func (s *TypeServiceImpl) Find(ctx context.Context, filter TypeCollectionFilter) ([]*CostType, error) {
	return s.types.Find(ctx, filter)
}

func NewTypeService(types TypeRepository, categories CategoryRepository, projects ProjectRepository) *TypeServiceImpl {
	return &TypeServiceImpl{
		types:      types,
		categories: categories,
		projects:   projects,
	}
}

func (s *TypeServiceImpl) Create(ctx context.Context, command CreateTypeCommand) (*CostType, error) {
	project, err := s.projects.FindOne(ctx, ProjectFilter{ProjectID: command.ProjectID})

	if err != nil {
		return nil, exceptions.NewValidationException(failedToCreateCostType, err)
	}

	category, err := s.categories.FindOne(ctx, CategoryFilter{CategoryID: command.CategoryID})

	if err != nil {
		return nil, exceptions.NewValidationException(failedToCreateCostType, err)
	}

	t, err := NewCostType(
		project.ProjectID,
		category,
		command.Name,
		command.Description,
	)

	if err != nil {
		return nil, exceptions.NewValidationException(failedToCreateCostType, err)
	}

	err = s.types.Save(ctx, t)

	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateCostType, err)
	}

	return t, nil
}

func (s *TypeServiceImpl) Remove(ctx context.Context, command RemoveProjectResourceCommand) error {
	t, err := s.types.FindOne(ctx, TypeFilter{
		TypeID:    command.ResourceID,
		ProjectID: command.ProjectID,
	})

	if errors.Is(err, exceptions.NotFoundError) {
		return exceptions.NewNotFoundException("failed to remove cost type", err)
	}

	if err != nil {
		return exceptions.NewInternalException("failed to remove cost type", err)
	}

	err = s.types.Remove(ctx, t)

	if err != nil {
		return exceptions.NewInternalException("failed to remove cost type", err)
	}

	return nil
}

func (s *TypeServiceImpl) Update(ctx context.Context, command UpdateTypeCommand) error {
	//TODO implement me
	panic("implement me")
}
