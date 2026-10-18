package investment

import (
	"context"

	"gitlab.com/massimo-ua/projecta/internal/core"
)

type Service interface {
	Create(ctx context.Context, command CreateInvestmentCommand) (*Investment, error)
	CreateBatch(ctx context.Context, command CreateBatchInvestmentsCommand) ([]*Investment, error)
	Update(ctx context.Context, command UpdateInvestmentCommand) error
	AssignToAsset(ctx context.Context, assetID core.AssetID, investmentIDs []core.InvestmentID) error
	Remove(ctx context.Context, command RemoveInvestmentCommand) error
	FindOne(ctx context.Context, filter Filter) (*Investment, error)
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
	FindTags(ctx context.Context, projectID core.ProjectID) ([]string, error)
}

type Repository interface {
	Save(ctx context.Context, inv *Investment) error
	Remove(ctx context.Context, inv *Investment) error
	FindOne(ctx context.Context, filter Filter) (*Investment, error)
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
	FindTags(ctx context.Context, projectID core.ProjectID) ([]string, error)
}

// AssetDescriptor represents the narrow read view of an asset needed by the investment context.
type AssetDescriptor struct {
	ID   core.AssetID
	Name string
}

// AssetReader is a consumer-owned interface providing read access to assets.
type AssetReader interface {
	FindAsset(ctx context.Context, projectID core.ProjectID, assetID core.AssetID) (AssetDescriptor, error)
}

// ProjectDescriptor represents the narrow read view of a project needed by the investment context.
type ProjectDescriptor struct {
	ID           core.ProjectID
	MainCurrency string
}

// ProjectReader is a consumer-owned interface providing read access to projects.
type ProjectReader interface {
	FindProject(ctx context.Context, projectID core.ProjectID) (ProjectDescriptor, error)
}

// ContributorReader is a consumer-owned interface providing read access to contributors.
type ContributorReader interface {
	FindContributor(ctx context.Context, personID core.PersonID) (ContributorRef, error)
}
