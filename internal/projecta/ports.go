package projecta

import (
	"context"
	"github.com/google/uuid"
)

type CategoryService interface {
	Find(ctx context.Context, filter CategoryCollectionFilter) ([]*CostCategory, error)
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
	Find(ctx context.Context, filter TypeCollectionFilter) ([]*CostType, error)
	FindOne(ctx context.Context, filter TypeFilter) (*CostType, error)
	Create(ctx context.Context, command CreateTypeCommand) (*CostType, error)
	Remove(ctx context.Context, command RemoveTypeCommand) error
	Update(ctx context.Context, command UpdateTypeCommand) error
}

type ExpenseService interface {
	Find(ctx context.Context, filter ExpenseCollectionFilter) ([]*Expense, error)
	Create(ctx context.Context, command CreateExpenseCommand) (*Expense, error)
	Update(ctx context.Context, command UpdateExpenseCommand) error
	Remove(ctx context.Context, command RemoveExpenseCommand) error
}

type CategoryRepository interface {
	Find(ctx context.Context, filter CategoryCollectionFilter) ([]*CostCategory, error)
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
	Find(ctx context.Context, filter TypeCollectionFilter) ([]*CostType, error)
	FindOne(ctx context.Context, filter TypeFilter) (*CostType, error)
	Save(ctx context.Context, costType *CostType) error
	Remove(ctx context.Context, costType *CostType) error
}

type ExpenseRepository interface {
	Find(ctx context.Context, filter ExpenseCollectionFilter) ([]*Expense, error)
	FindOne(ctx context.Context, filter ExpenseFilter) (*Expense, error)
	Save(ctx context.Context, expense *Expense) error
	Remove(ctx context.Context, expense *Expense) error
}
