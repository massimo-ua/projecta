package dal

import (
    "context"
    "errors"
    "github.com/huandu/go-sqlbuilder"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "gitlab.com/massimo-ua/projecta/internal/core"
    "gitlab.com/massimo-ua/projecta/internal/projecta"
)

type PgProjectaExpenseRepository struct {
    PgRepository
}

func NewPgProjectaExpenseRepository(pool *pgxpool.Pool) *PgProjectaExpenseRepository {
    return &PgProjectaExpenseRepository{
        PgRepository{db: pool},
    }
}

func (r *PgProjectaExpenseRepository) FindOne(ctx context.Context, filter projecta.ExpenseFilter) (*projecta.Expense, error) {
    return nil, nil
}

func (r *PgProjectaExpenseRepository) Save(ctx context.Context, expense *projecta.Expense) error {
    personID, err := core.AuthGuard(ctx)

    if err != nil {
        return err
    }

    qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
    qb.From("projecta_expenses")
    qb.Select(
        "expense_id",
        "project_id",
        "category_id",
        "type_id",
        "amount",
        "currency",
        "description",
        "created_at",
        "updated_at",
    )

    qb.Where(qb.Equal("expense_id", expense.ID.String()))
    qb.Where(qb.Equal("owner_id", personID.String()))

    sql, args := qb.Build()

    _, err = r.db.Exec(ctx, sql, args...)

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return r.create(ctx, expense)
        }

        return err
    }

    return r.update(ctx, expense)
}

func (r *PgProjectaExpenseRepository) create(ctx context.Context, expense *projecta.Expense) error {
    qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
    qb.InsertInto("projecta_expenses")
    qb.Cols(
        "expense_id",
        "project_id",
        "category_id",
        "type_id",
        "amount",
        "currency",
        "description",
        "owner_id",
    )

    qb.Values(
        expense.ID.String(),
        expense.Project.ProjectID.String(),
        expense.Category.ID.String(),
        expense.Type.ID.String(),
        expense.Amount.Amount(),
        expense.Amount.Currency().Code,
        expense.Description,
        expense.Owner.PersonID.String(),
    )

    sql, args := qb.Build()

    _, err := r.db.Exec(ctx, sql, args...)

    return err
}

func (r *PgProjectaExpenseRepository) update(ctx context.Context, expense *projecta.Expense) error {
    qb := sqlbuilder.PostgreSQL.NewUpdateBuilder()
    qb.Update("projecta_expenses")
    qb.Set(
        qb.Assign("project_id", expense.Project.ProjectID.String()),
        qb.Assign("category_id", expense.Category.ID.String()),
        qb.Assign("type_id", expense.Type.ID.String()),
        qb.Assign("amount", expense.Amount.Amount()),
        qb.Assign("currency", expense.Amount.Currency().Code),
        qb.Assign("description", expense.Description),
    )

    qb.Where(qb.Equal("expense_id", expense.ID.String()))

    sql, args := qb.Build()

    _, err := r.db.Exec(ctx, sql, args...)

    return err
}

func (r *PgProjectaExpenseRepository) Remove(ctx context.Context, expense *projecta.Expense) error {
    personID, err := core.AuthGuard(ctx)

    if err != nil {
        return core.FailedToIdentifyRequester
    }

    qb := sqlbuilder.PostgreSQL.NewDeleteBuilder()
    qb.DeleteFrom("projecta_expenses")
    qb.Where(qb.Equal("expense_id", expense.ID.String()))
    qb.Where(qb.Equal("owner_id", personID.String()))
    qb.Where(qb.Equal("project_id", expense.Project.ProjectID.String()))

    sql, args := qb.Build()

    _, err = r.db.Exec(ctx, sql, args...)

    return err
}
