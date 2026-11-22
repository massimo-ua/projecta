package investment

import (
	"context"

	"gitlab.com/massimo-ua/projecta/internal/core"
)

// Repository defines persistence and transactional read-modify-write access for Investment aggregates.
type Repository interface {
	Save(ctx context.Context, inv *Investment) error
	Remove(ctx context.Context, inv *Investment) error
	FindOne(ctx context.Context, filter Filter) (*Investment, error)
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
	FindTags(ctx context.Context, projectID core.ProjectID) ([]string, error)
	UpdateInvestment(
		ctx context.Context,
		id core.InvestmentID,
		projectID core.ProjectID,
		updateFn func(ctx context.Context, inv *Investment) (*Investment, error),
	) error
}
