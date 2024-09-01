package asset

import (
	"context"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type Service interface {
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
	FindOne(ctx context.Context, filter Filter) (*Asset, error)
	Create(ctx context.Context, command CreateAssetCommand) (*Asset, error)
	Update(ctx context.Context, command UpdateAssetCommand) error
	Remove(ctx context.Context, command RemoveAssetCommand) error
}

type Repository interface {
	core.BaseRepository
	Save(ctx context.Context, asset *Asset) error
	Remove(ctx context.Context, asset *Asset) error
	FindOne(ctx context.Context, filter Filter) (*Asset, error)
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
}
