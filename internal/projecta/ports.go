package projecta

import (
	"context"

	"github.com/google/uuid"
)

type PeopleService interface {
	FindOwner(ctx context.Context, personID uuid.UUID) (*Owner, error)
}

type ProjectService interface {
	Find(ctx context.Context, filter ProjectCollectionFilter) ([]*Project, error)
	FindOne(ctx context.Context, filter ProjectFilter) (*Project, error)
	Create(ctx context.Context, command CreateProjectCommand) (*Project, error)
	Remove(ctx context.Context, command RemoveProjectCommand) error
	Update(ctx context.Context, command UpdateProjectCommand) (*Project, error)
	AcceptShare(ctx context.Context, token uuid.UUID, personID uuid.UUID) (*Project, error)
}

type ProjectRepository interface {
	Find(ctx context.Context, filter ProjectCollectionFilter) ([]*Project, error)
	FindOne(ctx context.Context, filter ProjectFilter) (*Project, error)
	FindByShareToken(ctx context.Context, token uuid.UUID) (*Project, error)
	Create(ctx context.Context, project *Project) error
	Update(ctx context.Context, project *Project) error
	Remove(ctx context.Context, project *Project) error
	CreateShareRecord(ctx context.Context, projectID uuid.UUID, personID uuid.UUID) error
}
