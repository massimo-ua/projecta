package project

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/people"
)

type PersonReader interface {
	FindByID(ctx context.Context, personID uuid.UUID) (*people.Person, error)
}

type PeopleServiceImpl struct {
	repository PersonReader
}

func NewPeopleService(repository PersonReader) *PeopleServiceImpl {
	return &PeopleServiceImpl{repository: repository}
}

func (s *PeopleServiceImpl) FindOwner(ctx context.Context, personID uuid.UUID) (*Owner, error) {
	person, err := s.repository.FindByID(ctx, personID)

	if err != nil {
		return nil, err
	}

	return &Owner{
		PersonID:        person.ID(),
		DisplayName:     person.DisplayName(),
		FirstName:       person.FirstName(),
		CanHaveProjects: person.CanHaveProjects(),
	}, nil
}
