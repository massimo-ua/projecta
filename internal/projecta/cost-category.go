package projecta

import (
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type CostCategory struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Name        string
	Description string
}

const (
	MinCostCategoryNameLength = 3
	MaxCostCategoryNameLength = 100
)

func NewCostCategory(id uuid.UUID, projectID uuid.UUID, name string, description string) (*CostCategory, error) {
	if name == "" || len(name) < MinCostCategoryNameLength || len(name) > MaxCostCategoryNameLength {
		return nil, exceptions.NewValidationException("cost category name must be between 3 and 100 characters", nil)
	}

	return &CostCategory{
		ID:          id,
		ProjectID:   projectID,
		Name:        name,
		Description: description,
	}, nil
}

type CostCategoryCollection = core.PaginatedCollection[*CostCategory]

func NewCategoryCollection(total int) *CostCategoryCollection {
	return core.NewPaginatedCollection[*CostCategory](total)
}
