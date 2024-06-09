package projecta

import "github.com/google/uuid"

type CostType struct {
    ID          uuid.UUID
    ProjectID   uuid.UUID
    Name        string
    Description string
}

func NewCostType(projectID uuid.UUID, name string, description string) (*CostType, error) {
    t := &CostType{
        ID:          uuid.New(),
        ProjectID:   projectID,
        Name:        name,
        Description: description,
    }

    return t, nil
}
