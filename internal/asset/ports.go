package asset

import "context"

type Service interface {
	Create(ctx context.Context, command CreateAssetCommand) (*Asset, error)
	Remove(ctx context.Context, command RemoveAssetCommand) error
}

type Repository interface {
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
	FindOne(ctx context.Context, filter Filter) (*Asset, error)
	Save(ctx context.Context, asset *Asset) error
	Remove(ctx context.Context, asset *Asset) error
}
