package projecta

import (
    "github.com/google/uuid"
    "gitlab.com/massimo-ua/projecta/internal/exceptions"
    "time"
)

const (
    MinProjectNameLength = 3
    MaxProjectNameLength = 100
)

type Project struct {
    ProjectID   uuid.UUID
    Name        string
    Description string
    Owner       *Owner
    StartDate   time.Time
    EndDate     time.Time
}

func (p *Project) IsOwnedBy(owner *Owner) bool {
    return p.Owner.PersonID == owner.PersonID
}

func NewProject(id uuid.UUID, name string, description string, owner *Owner, startDate time.Time, endDate time.Time) (*Project, error) {
    if name == "" || len(name) < MinProjectNameLength || len(name) > MaxProjectNameLength {
        return nil, exceptions.NewValidationException("project name must be between 3 and 100 characters", nil)
    }

    return &Project{
        ProjectID:   id,
        Name:        name,
        Description: description,
        Owner:       owner,
        StartDate:   startDate,
        EndDate:     endDate,
    }, nil
}
