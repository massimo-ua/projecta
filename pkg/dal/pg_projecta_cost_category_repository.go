package dal

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

type PgProjectaCostCategoryRepository struct {
	PgRepository
}

func NewPgProjectaCategoryRepository(pool *pgxpool.Pool) *PgProjectaCostCategoryRepository {
	return &PgProjectaCostCategoryRepository{
		PgRepository{db: pool},
	}
}

func (r *PgProjectaCostCategoryRepository) Find(ctx context.Context, filter projecta.CategoryCollectionFilter) ([]*projecta.CostCategory, error) {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return nil, err
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_cost_categories")
	qb.Select(
		"projecta_cost_categories.category_id",
		"projecta_cost_categories.project_id",
		"projecta_cost_categories.name",
		"projecta_cost_categories.description",
	)
	qb.Join("projecta_projects", "projecta_projects.project_id = projecta_cost_categories.project_id")
	qb.Where(qb.Equal("projecta_projects.owner_id", personID.String()))
	qb.Limit(filter.Limit)
	qb.Offset(filter.Offset)

	if filter.Name != "" {
		qb.Where(qb.Like("projecta_cost_categories.name", fmt.Sprintf("%s%%", filter.Name)))
	}

	if filter.ProjectID != uuid.Nil {
		qb.Where(qb.Equal("projecta_cost_categories.project_id", filter.ProjectID.String()))
	}

	sql, args := qb.Build()

	var categories []*projecta.CostCategory = make([]*projecta.CostCategory, 0)

	rows, err := r.db.Query(ctx, sql, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			categoryID  string
			projectID   string
			name        string
			description string
		)

		if err := rows.Scan(&categoryID, &projectID, &name, &description); err != nil {
			return nil, err
		}

		category, err := toCostCategory(categoryID, projectID, name, description)

		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func (r *PgProjectaCostCategoryRepository) FindOne(ctx context.Context, filter projecta.CategoryFilter) (*projecta.CostCategory, error) {
	personID := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)

	if personID == uuid.Nil {
		return nil, core.FailedToIdentifyRequester
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_cost_categories")
	qb.Select("category_id", "project_id", "name", "description")

	if filter.CategoryID != uuid.Nil {
		qb.Where(qb.Equal("category_id", filter.CategoryID.String()))
	}

	if filter.ProjectID != uuid.Nil {
		qb.Where(qb.Equal("project_id", filter.ProjectID.String()))
	}

	if filter.Name != "" {
		qb.Where(qb.Equal("name", filter.Name))
	}

	sql, args := qb.Build()

	var (
		categoryID  string
		projectID   string
		name        string
		description string
	)

	if err := r.db.QueryRow(
		ctx,
		sql,
		args...,
	).Scan(
		&categoryID,
		&projectID,
		&name,
		&description,
	); err != nil {
		return nil, err
	}
	return toCostCategory(categoryID, projectID, name, description)
}

func (r *PgProjectaCostCategoryRepository) Save(ctx context.Context, category *projecta.CostCategory) error {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_cost_categories")
	qb.Select("1 as exists")
	qb.Where(qb.Equal("category_id", category.ID.String()))
	qb.Where(qb.Equal("project_id", category.ProjectID.String()))

	sql, args := qb.Build()
	err := r.db.QueryRow(ctx, sql, args...).Scan()

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return r.create(ctx, category)
		}

		return err
	}

	return r.update(ctx, category)
}

func (r *PgProjectaCostCategoryRepository) Remove(ctx context.Context, category *projecta.CostCategory) error {
	qb := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	qb.DeleteFrom("projecta_cost_categories")
	qb.Where(qb.Equal("category_id", category.ID.String()))

	sql, args := qb.Build()
	res, err := r.db.Exec(ctx, sql, args...)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return errors.New("failed to save category")
	}

	return nil
}

func (r *PgProjectaCostCategoryRepository) create(ctx context.Context, category *projecta.CostCategory) error {
	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
	qb.InsertInto("projecta_cost_categories")
	qb.Cols("category_id", "project_id", "name", "description")
	qb.Values(category.ID.String(), category.ProjectID.String(), category.Name, category.Description)

	sql, args := qb.Build()
	_, err := r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgProjectaCostCategoryRepository) update(ctx context.Context, category *projecta.CostCategory) error {
	qb := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	qb.Update("projecta_cost_categories")
	qb.Set(
		qb.Assign("name", category.Name),
		qb.Assign("description", category.Description),
	)
	qb.Where(qb.Equal("category_id", category.ID.String()))

	sql, args := qb.Build()
	res, err := r.db.Exec(ctx, sql, args...)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return errors.New("failed to update category")
	}

	return nil
}

func toCostCategory(
	categoryID string,
	projectID string,
	name string,
	description string,
) (*projecta.CostCategory, error) {
	categoryUUID, err := uuid.Parse(categoryID)

	if err != nil {
		return nil, err
	}

	projectUUID, err := uuid.Parse(projectID)

	if err != nil {
		return nil, err
	}

	return projecta.NewCostCategory(categoryUUID, projectUUID, name, description)
}
