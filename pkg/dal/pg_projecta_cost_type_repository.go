package dal

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

type PgProjectaCostTypeRepository struct {
	PgRepository
}

func NewPgProjectaCostTypeRepository(pool *pgxpool.Pool) *PgProjectaCostTypeRepository {
	return &PgProjectaCostTypeRepository{
		PgRepository{db: pool},
	}
}

func (r *PgProjectaCostTypeRepository) FindOne(ctx context.Context, filter projecta.TypeFilter) (*projecta.CostType, error) {
	personID := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)

	if personID == uuid.Nil {
		return nil, core.FailedToIdentifyRequester
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_cost_types")
	qb.Select("type_id", "projecta_cost_types.project_id", "projecta_cost_types.category_id", "projecta_cost_types.name", "projecta_cost_types.description", "projecta_cost_categories.name as category_name")
	qb.Join("projecta_cost_categories", "projecta_cost_categories.category_id = projecta_cost_types.category_id")
	qb.Limit(1)

	if filter.TypeID != uuid.Nil {
		qb.Where(qb.Equal("type_id", filter.TypeID.String()))
	}

	if filter.ProjectID != uuid.Nil {
		qb.Where(qb.Equal("projecta_cost_types.project_id", filter.ProjectID.String()))
	}

	if filter.CategoryID != uuid.Nil {
		qb.Where(qb.Equal("projecta_cost_types.category_id", filter.CategoryID.String()))
	}

	if filter.Name != "" {
		qb.Where(qb.Equal("projecta_cost_types.name", filter.Name))
	}

	sql, args := qb.Build()

	var (
		typeID       string
		projectID    string
		categoryID   string
		name         string
		description  string
		categoryName string
	)

	if err := r.db.QueryRow(
		ctx,
		sql,
		args...,
	).Scan(
		&typeID,
		&projectID,
		&categoryID,
		&name,
		&description,
		&categoryName,
	); err != nil {
		return nil, err
	}

	return toCostType(typeID, projectID, name, description, categoryID, categoryName)
}

func (r *PgProjectaCostTypeRepository) Save(ctx context.Context, costType *projecta.CostType) error {
	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
	qb.InsertInto("projecta_cost_types")
	qb.Cols("type_id", "project_id", "category_id", "name", "description")
	qb.Values(
		costType.ID.String(),
		costType.ProjectID.String(),
		costType.Category.ID.String(),
		costType.Name,
		costType.Description,
	)

	sql, args := qb.Build()

	_, err := r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgProjectaCostTypeRepository) Remove(ctx context.Context, costType *projecta.CostType) error {
	// TODO: Verify that the person is the owner of the project that the cost type belongs to
	qb := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	qb.DeleteFrom("projecta_cost_types")
	qb.Where(qb.Equal("type_id", costType.ID.String()))
	qb.Where(qb.Equal("project_id", costType.ProjectID.String()))

	sql, args := qb.Build()

	_, err := r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgProjectaCostTypeRepository) Find(ctx context.Context, filter projecta.TypeCollectionFilter) ([]*projecta.CostType, error) {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return nil, err
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_cost_types")
	qb.Select(
		"type_id",
		"projecta_projects.project_id",
		"projecta_cost_types.category_id",
		"projecta_cost_categories.name as category_name",
		"projecta_cost_types.name",
		"projecta_cost_types.description",
	)
	qb.Join("projecta_projects", "projecta_projects.project_id = projecta_cost_types.project_id")
	qb.Join("projecta_cost_categories", "projecta_cost_categories.category_id = projecta_cost_types.category_id")
	qb.Where(qb.Equal("projecta_projects.owner_id", personID.String()))
	qb.Limit(filter.Limit)
	qb.Offset(filter.Offset)

	if filter.ProjectID != uuid.Nil {
		qb.Where(qb.Equal("projecta_cost_types.project_id", filter.ProjectID.String()))
	}

	if filter.Name != "" {
		qb.Where(qb.Like("projecta_cost_types.name", fmt.Sprintf("%s%%", filter.Name)))
	}

	if filter.CategoryID != uuid.Nil {
		qb.Where(qb.Equal("projecta_cost_types.category_id", filter.CategoryID.String()))
	}

	var costTypes []*projecta.CostType = make([]*projecta.CostType, 0)

	sql, args := qb.Build()

	rows, err := r.db.Query(ctx, sql, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var (
			typeID       string
			projectID    string
			categoryID   string
			categoryName string
			name         string
			description  string
		)

		if err := rows.Scan(
			&typeID,
			&projectID,
			&categoryID,
			&categoryName,
			&name,
			&description,
		); err != nil {
			return nil, err
		}

		costType, err := toCostType(typeID, projectID, name, description, categoryID, categoryName)

		if err != nil {
			return nil, err
		}

		costTypes = append(costTypes, costType)
	}

	return costTypes, nil
}

func toCostType(typeID, projectID, name, description, categoryID, categoryName string) (*projecta.CostType, error) {
	return &projecta.CostType{
		ID:        uuid.MustParse(typeID),
		ProjectID: uuid.MustParse(projectID),
		Category: &projecta.CostCategory{
			ID:   uuid.MustParse(categoryID),
			Name: categoryName,
		},
		Name:        name,
		Description: description,
	}, nil
}
