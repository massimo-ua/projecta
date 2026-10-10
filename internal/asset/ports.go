package asset

import (
	"context"

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
