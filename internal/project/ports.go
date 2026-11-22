package project

import (
	"context"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

// CurrencyConverter is a narrow domain interface for converting monetary amounts to a target currency code.
type CurrencyConverter interface {
	ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error)
}

// ProjectTotals represents aggregated investment and balance totals for a project.
type ProjectTotals struct {
	TotalInvested int64
	Balance       int64
	HasActivity   bool
	Currency      string
}

// ProjectRepository defines persistence and secure-by-design transactional access for Project aggregates.
type ProjectRepository interface {
	Find(ctx context.Context, filter ProjectCollectionFilter) ([]*Project, error)
	FindOne(ctx context.Context, filter ProjectFilter) (*Project, error)
	GetProject(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID) (*Project, error)
	FindByShareToken(ctx context.Context, token uuid.UUID) (*Project, error)
	Create(ctx context.Context, project *Project) error
	Update(ctx context.Context, project *Project) error
	UpdateProject(
		ctx context.Context,
		projectID core.ProjectID,
		requesterID core.PersonID,
		updateFn func(ctx context.Context, p *Project) (*Project, error),
	) error
	Remove(ctx context.Context, project *Project) error
	CreateShareRecord(ctx context.Context, projectID core.ProjectID, personID core.PersonID) error
}
