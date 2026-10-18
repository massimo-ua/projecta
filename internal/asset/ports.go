package asset

import (
	"context"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type Service interface {
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
	FindOne(ctx context.Context, filter Filter) (*Asset, error)
	Create(ctx context.Context, command CreateAssetCommand) (*Asset, error)
	CreateFromInvestments(ctx context.Context, command CreateAssetFromInvestmentsCommand) (*Asset, error)
	CreateFromPayments(ctx context.Context, command CreateAssetFromPaymentsCommand) (*Asset, error)
	Update(ctx context.Context, command UpdateAssetCommand) error
	Remove(ctx context.Context, command RemoveAssetCommand) error
	LinkChild(ctx context.Context, command LinkChildCommand) error
	UnlinkChild(ctx context.Context, command UnlinkChildCommand) error
	AssignInvestments(ctx context.Context, command AssignInvestmentsCommand) error
	Group(ctx context.Context, command GroupAssetsCommand) (*Asset, error)
	LinkChildren(ctx context.Context, command LinkChildrenCommand) error
}

type Repository interface {
	Save(ctx context.Context, asset *Asset) error
	Remove(ctx context.Context, asset *Asset) error
	FindOne(ctx context.Context, filter Filter) (*Asset, error)
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
	AddChild(ctx context.Context, parentID core.AssetID, childID core.AssetID, sharePercentage float64) error
	RemoveChild(ctx context.Context, parentID core.AssetID, childID core.AssetID) error
	FindChildren(ctx context.Context, parentID core.AssetID) ([]ChildAssetLink, error)
	FindParents(ctx context.Context, childID core.AssetID) ([]ParentAssetLink, error)
	FindAncestors(ctx context.Context, assetID core.AssetID) ([]core.AssetID, error)
}

type InvestmentItem struct {
	ID     core.InvestmentID
	Amount *money.Money
	Date   time.Time
	Tags   []string
}

type InitialInvestment struct {
	ID            core.InvestmentID
	ProjectID     core.ProjectID
	AssetID       core.AssetID
	ContributorID core.PersonID
	Amount        *money.Money
	Date          time.Time
	Description   string
	Tags          []string
}

type InvestmentSource interface {
	FindInvestments(ctx context.Context, projectID core.ProjectID, ids []core.InvestmentID) ([]InvestmentItem, error)
	AssignToAsset(ctx context.Context, assetID core.AssetID, investmentIDs []core.InvestmentID) error
	CreateInitialInvestment(ctx context.Context, item InitialInvestment) error
}

// ProjectDescriptor represents the narrow read view of a project needed by the asset context.
type ProjectDescriptor struct {
	ID           core.ProjectID
	MainCurrency string
}

// ProjectReader is a consumer-owned interface providing read access to projects.
type ProjectReader interface {
	FindProject(ctx context.Context, projectID core.ProjectID) (ProjectDescriptor, error)
}

// OwnerReader is a consumer-owned interface providing read access to asset owners.
type OwnerReader interface {
	FindOwner(ctx context.Context, personID core.PersonID) (OwnerRef, error)
}
