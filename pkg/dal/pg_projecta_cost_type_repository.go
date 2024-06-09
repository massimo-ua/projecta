package dal

import (
    "context"
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
    qb.Select("type_id", "project_id", "name", "description")
    qb.Limit(1)

    if filter.TypeID != uuid.Nil {
        qb.Where(qb.Equal("type_id", filter.TypeID.String()))
    }

    if filter.ProjectID != uuid.Nil {
        qb.Where(qb.Equal("project_id", filter.ProjectID.String()))
    }

    if filter.Name != "" {
        qb.Where(qb.Equal("name", filter.Name))
    }

    sql, args := qb.Build()

    var (
        typeID      string
        projectID   string
        name        string
        description string
    )

    if err := r.db.QueryRow(
        ctx,
        sql,
        args...,
    ).Scan(
        &typeID,
        &projectID,
        &name,
        &description,
    ); err != nil {
        return nil, err
    }

    return toCostType(typeID, projectID, name, description)
}

func (r *PgProjectaCostTypeRepository) Save(ctx context.Context, costType *projecta.CostType) error {
    qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
    qb.InsertInto("projecta_cost_types")
    qb.Cols("type_id", "project_id", "name", "description")
    qb.Values(
        costType.ID.String(),
        costType.ProjectID.String(),
        costType.Name,
        costType.Description,
    )

    sql, args := qb.Build()

    _, err := r.db.Exec(ctx, sql, args...)

    return err
}

func (r *PgProjectaCostTypeRepository) Remove(ctx context.Context, costType *projecta.CostType) error {
    qb := sqlbuilder.PostgreSQL.NewDeleteBuilder()
    qb.DeleteFrom("projecta_cost_types")
    qb.Where(qb.Equal("type_id", costType.ID.String()))
    qb.Where(qb.Equal("project_id", costType.ProjectID.String()))

    sql, args := qb.Build()

    _, err := r.db.Exec(ctx, sql, args...)

    return err
}

func toCostType(typeID, projectID, name, description string) (*projecta.CostType, error) {
    return &projecta.CostType{
        ID:          uuid.MustParse(typeID),
        ProjectID:   uuid.MustParse(projectID),
        Name:        name,
        Description: description,
    }, nil
}
