package projecta

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type CategoryServiceImpl struct {
	repository     CategoryRepository
	projectService ProjectService
}

func NewCategoryService(repository CategoryRepository, projectService ProjectService) *CategoryServiceImpl {
	return &CategoryServiceImpl{repository: repository, projectService: projectService}
}

func (s *CategoryServiceImpl) Find(ctx context.Context, filter CategoryCollectionFilter) ([]*CostCategory, error) {
	categories, err := s.repository.Find(ctx, filter)

	if err != nil {
		return nil, exceptions.NewInternalException("failed to find cost categories", err)
	}

	return categories, nil
}

func (s *CategoryServiceImpl) Create(ctx context.Context, command CreateCategoryCommand) (*CostCategory, error) {
	project, err := s.projectService.FindOne(ctx, ProjectFilter{ProjectID: command.ProjectID})

	if err != nil {
		return nil, exceptions.NewInternalException(
			"failed to create category",
			errors.Join(err, errors.New("failed to find project")),
		)
	}

	if project == nil {
		return nil, exceptions.NewValidationException(
			"failed to create category",
			errors.New("project not found"),
		)
	}

	category, err := NewCostCategory(
		uuid.New(),
		command.ProjectID,
		command.Name,
		command.Description,
	)

	err = s.repository.Save(ctx, category)

	if err != nil {
		return nil, exceptions.NewInternalException("failed to find cost category", err)
	}

	return category, nil
}

func (s *CategoryServiceImpl) Remove(ctx context.Context, command RemoveCategoryCommand) error {
	category, err := s.repository.FindOne(ctx, CategoryFilter{CategoryID: command.ID, ProjectID: command.ProjectID})

	if err != nil {
		return exceptions.NewInternalException("failed to find cost category", err)
	}

	err = s.repository.Remove(ctx, category)

	if err != nil {
		return exceptions.NewInternalException("failed to remove cost category", err)
	}

	return nil
}

func (s *CategoryServiceImpl) Update(ctx context.Context, command UpdateCategoryCommand) error {
	category, err := s.repository.FindOne(ctx, CategoryFilter{CategoryID: command.ID, ProjectID: command.ProjectID})

	if err != nil {
		return exceptions.NewInternalException("failed to find cost category", err)
	}

	category.Name = command.Name
	category.Description = command.Description

	err = s.repository.Save(ctx, category)

	if err != nil {
		return exceptions.NewInternalException("failed to save cost category", err)
	}

	return nil
}
