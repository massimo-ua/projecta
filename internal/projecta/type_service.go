package projecta

import (
    "context"
    "gitlab.com/massimo-ua/projecta/internal/exceptions"
)

const (
    failedToCreateCostType = "failed to create cost type"
)

type TypeServiceImpl struct {
    repository TypeRepository
}

func (s *TypeServiceImpl) FindOne(ctx context.Context, filter TypeFilter) (*CostType, error) {
    return s.repository.FindOne(ctx, filter)
}

func NewTypeService(repository TypeRepository) *TypeServiceImpl {
    return &TypeServiceImpl{repository: repository}
}

func (s *TypeServiceImpl) Create(ctx context.Context, command CreateTypeCommand) (*CostType, error) {
    t, err := NewCostType(
        command.ProjectID,
        command.Name,
        command.Description,
    )

    if err != nil {
        return nil, exceptions.NewValidationException(failedToCreateCostType, err)
    }

    err = s.repository.Save(ctx, t)

    if err != nil {
        return nil, exceptions.NewInternalException(failedToCreateCostType, err)
    }

    return t, nil
}

func (s *TypeServiceImpl) Remove(ctx context.Context, command RemoveTypeCommand) error {
    //TODO implement me
    panic("implement me")
}

func (s *TypeServiceImpl) Update(ctx context.Context, command UpdateTypeCommand) error {
    //TODO implement me
    panic("implement me")
}
