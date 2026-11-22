package asset

import (
	"context"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

// Repository defines persistence and transactional read-modify-write access for Asset aggregates.
type Repository interface {
	Save(ctx context.Context, asset *Asset) error
	Remove(ctx context.Context, asset *Asset) error
	FindOne(ctx context.Context, filter Filter) (*Asset, error)
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
	UpdateAsset(
		ctx context.Context,
		assetID core.AssetID,
		projectID core.ProjectID,
		ownerID core.PersonID,
		updateFn func(ctx context.Context, a *Asset) (*Asset, error),
	) error
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

type InvestmentAllocation struct {
	AssetID         core.AssetID
	SharePercentage float64
}

type InvestmentRecordSummary struct {
	ID               core.InvestmentID
	ResourceType     string
	Amount           *money.Money
	TimeHours        float64
	Date             time.Time
	AssetAllocations []InvestmentAllocation
	Tags             []string
}

// ProjectDescriptor represents the narrow read view of a project needed by the asset domain.
type ProjectDescriptor struct {
	ID           core.ProjectID
	MainCurrency string
	StartDate    time.Time
	EndDate      time.Time
}

type GetCashFlowForecastQuery struct {
	ProjectID core.ProjectID
	FromMonth string
	ToMonth   string
	AsOfDate  time.Time
	Tag       string
}
