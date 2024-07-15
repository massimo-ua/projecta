package projecta

import (
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type CostType struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Category    *CostCategory
	Name        string
	Description string
}

func NewCostType(projectID uuid.UUID, category *CostCategory, name string, description string) (*CostType, error) {
	t := &CostType{
		ID:          uuid.New(),
		ProjectID:   projectID,
		Category:    category,
		Name:        name,
		Description: description,
	}

	return t, nil
}

type CostTypeCollection = core.PaginatedCollection[*CostType]

func NewCostTypeCollection(total int) *CostTypeCollection {
	return core.NewPaginatedCollection[*CostType](total)
}
