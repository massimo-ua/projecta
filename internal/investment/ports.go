package investment

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, command CreateInvestmentCommand) (*Investment, error)
	CreateBatch(ctx context.Context, command CreateBatchInvestmentsCommand) ([]*Investment, error)
	Update(ctx context.Context, command UpdateInvestmentCommand) error
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
