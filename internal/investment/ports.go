package investment

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, command CreateInvestmentCommand) (*Investment, error)
	CreateBatch(ctx context.Context, command CreateBatchInvestmentsCommand) ([]*Investment, error)
	Update(ctx context.Context, command UpdateInvestmentCommand) error
	AssignToAsset(ctx context.Context, assetID uuid.UUID, investmentIDs []uuid.UUID) error
	Remove(ctx context.Context, command RemoveInvestmentCommand) error
	FindOne(ctx context.Context, filter Filter) (*Investment, error)
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
	FindTags(ctx context.Context, projectID uuid.UUID) ([]string, error)
}

type Repository interface {
	Save(ctx context.Context, inv *Investment) error
	Remove(ctx context.Context, inv *Investment) error
	FindOne(ctx context.Context, filter Filter) (*Investment, error)
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
	FindTags(ctx context.Context, projectID uuid.UUID) ([]string, error)
}

// AssetDescriptor represents the narrow read view of an asset needed by the investment context.
type AssetDescriptor struct {
	ID   uuid.UUID
	Name string
}

// AssetReader is a consumer-owned interface providing read access to assets.
type AssetReader interface {
	FindAsset(ctx context.Context, projectID uuid.UUID, assetID uuid.UUID) (AssetDescriptor, error)
}

// ProjectDescriptor represents the narrow read view of a project needed by the investment context.
type ProjectDescriptor struct {
	ID           uuid.UUID
	MainCurrency string
}

// ProjectReader is a consumer-owned interface providing read access to projects.
type ProjectReader interface {
	FindProject(ctx context.Context, projectID uuid.UUID) (ProjectDescriptor, error)
}

// ContributorReader is a consumer-owned interface providing read access to contributors.
type ContributorReader interface {
	FindContributor(ctx context.Context, personID uuid.UUID) (ContributorRef, error)
}
