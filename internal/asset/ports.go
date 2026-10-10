package asset

import (
	"context"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
)

type Service interface {
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
	FindOne(ctx context.Context, filter Filter) (*Asset, error)
	Create(ctx context.Context, command CreateAssetCommand) (*Asset, error)
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
	AddChild(ctx context.Context, parentID uuid.UUID, childID uuid.UUID, sharePercentage float64) error
	RemoveChild(ctx context.Context, parentID uuid.UUID, childID uuid.UUID) error
	FindChildren(ctx context.Context, parentID uuid.UUID) ([]ChildAssetLink, error)
	FindParents(ctx context.Context, childID uuid.UUID) ([]ParentAssetLink, error)
	FindAncestors(ctx context.Context, assetID uuid.UUID) ([]uuid.UUID, error)
}

type InvestmentItem struct {
	ID     uuid.UUID
	Amount *money.Money
	Date   time.Time
	Tags   []string
}

type InitialInvestment struct {
	ID            uuid.UUID
	ProjectID     uuid.UUID
	AssetID       uuid.UUID
	ContributorID uuid.UUID
	Amount        *money.Money
	Date          time.Time
	Description   string
	Tags          []string
}

type InvestmentSource interface {
	FindInvestments(ctx context.Context, projectID uuid.UUID, ids []uuid.UUID) ([]InvestmentItem, error)
	AssignToAsset(ctx context.Context, assetID uuid.UUID, investmentIDs []uuid.UUID) error
	CreateInitialInvestment(ctx context.Context, item InitialInvestment) error
}

