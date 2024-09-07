package projecta

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type CategoryService interface {
	Find(ctx context.Context, filter CategoryCollectionFilter) (*CostCategoryCollection, error)
	Create(ctx context.Context, command CreateCategoryCommand) (*CostCategory, error)
	Update(ctx context.Context, command UpdateCategoryCommand) error
	Remove(ctx context.Context, command RemoveCategoryCommand) error
}

type PeopleService interface {
	FindOwner(ctx context.Context, personID uuid.UUID) (*Owner, error)
}

type ProjectService interface {
	Find(ctx context.Context, filter ProjectCollectionFilter) ([]*Project, error)
	FindOne(ctx context.Context, filter ProjectFilter) (*Project, error)
	Create(ctx context.Context, command CreateProjectCommand) (*Project, error)
	Remove(ctx context.Context, command RemoveProjectCommand) error
	Update(ctx context.Context, command UpdateProjectCommand) error
}

type TypeService interface {
	Find(ctx context.Context, filter TypeCollectionFilter) (*CostTypeCollection, error)
	FindOne(ctx context.Context, filter TypeFilter) (*CostType, error)
	Create(ctx context.Context, command CreateTypeCommand) (*CostType, error)
	Remove(ctx context.Context, command RemoveProjectResourceCommand) error
	Update(ctx context.Context, command UpdateTypeCommand) error
}

type PaymentService interface {
	FindOne(ctx context.Context, filter PaymentFilter) (*Payment, error)
	Find(ctx context.Context, filter PaymentCollectionFilter) (*PaymentCollection, error)
	Create(ctx context.Context, command CreatePaymentCommand) (*Payment, error)
	Update(ctx context.Context, command UpdatePaymentCommand) error
	Remove(ctx context.Context, command RemovePaymentCommand) error
}

type CategoryRepository interface {
	Find(ctx context.Context, filter CategoryCollectionFilter) (*CostCategoryCollection, error)
	FindOne(ctx context.Context, filter CategoryFilter) (*CostCategory, error)
	Save(ctx context.Context, category *CostCategory) error
	Remove(ctx context.Context, category *CostCategory) error
}

type ProjectRepository interface {
	Find(ctx context.Context, filter ProjectCollectionFilter) ([]*Project, error)
	FindOne(ctx context.Context, filter ProjectFilter) (*Project, error)
	Create(ctx context.Context, project *Project) error
	Update(ctx context.Context, project *Project) error
	Remove(ctx context.Context, project *Project) error
}

type TypeRepository interface {
	Find(ctx context.Context, filter TypeCollectionFilter) (*CostTypeCollection, error)
	FindOne(ctx context.Context, filter TypeFilter) (*CostType, error)
	Save(ctx context.Context, costType *CostType) error
	Remove(ctx context.Context, costType *CostType) error
}

type PaymentRepository interface {
	core.BaseRepository
	Find(ctx context.Context, filter PaymentCollectionFilter) (*PaymentCollection, error)
	FindOne(ctx context.Context, filter PaymentFilter) (*Payment, error)
	Save(ctx context.Context, payment *Payment) error
	Remove(ctx context.Context, payment *Payment) error
}
