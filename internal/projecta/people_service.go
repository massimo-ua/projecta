package projecta

import (
    "context"
    "github.com/google/uuid"
    "gitlab.com/massimo-ua/projecta/internal/people"
)

type PeopleServiceImpl struct {
    repository people.Repository
}

func NewPeopleService(repository people.Repository) *PeopleServiceImpl {
    return &PeopleServiceImpl{repository: repository}
}

func (s *PeopleServiceImpl) FindOwner(ctx context.Context, personID uuid.UUID) (*Owner, error) {
    person, err := s.repository.FindByID(ctx, personID)

    if err != nil {
        return nil, err
    }

    return &Owner{PersonID: person.ID, DisplayName: person.DisplayName()}, nil
}
