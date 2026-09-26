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
    ProjectID    uuid.UUID
    Name         string
    Description  string
    Owner        *Owner
    StartDate    time.Time
    EndDate      time.Time
    ShareToken   uuid.UUID
    IsShared     bool
    MainCurrency string
    Participants []Participant
}

func (p *Project) IsOwnedBy(owner *Owner) bool {
    return p.Owner.PersonID == owner.PersonID
}

func (p *Project) AddParticipant(participant Participant) {
    for _, existing := range p.Participants {
        if existing.Equals(participant) {
            return
        }
    }
    p.Participants = append(p.Participants, participant)
}

func NewProject(id uuid.UUID, name string, description string, owner *Owner, startDate time.Time, endDate time.Time, mainCurrency ...string) (*Project, error) {
    if name == "" || len(name) < MinProjectNameLength || len(name) > MaxProjectNameLength {
        return nil, exceptions.NewValidationException("project name must be between 3 and 100 characters", nil)
    }

    curr := "UAH"
    if len(mainCurrency) > 0 && mainCurrency[0] != "" {
        curr = mainCurrency[0]
    }

    proj := &Project{
        ProjectID:    id,
        Name:         name,
        Description:  description,
        Owner:        owner,
        StartDate:    startDate,
        EndDate:      endDate,
        ShareToken:   uuid.New(),
        MainCurrency: curr,
        Participants: make([]Participant, 0),
    }

    if owner != nil && owner.DisplayName != "" {
        if participant, err := NewParticipant(owner.DisplayName); err == nil {
            proj.AddParticipant(participant)
        }
    }

    return proj, nil
}
