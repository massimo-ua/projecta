package dal

import (
	"context"
	"fmt"
	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"time"
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
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return nil, core.FailedToIdentifyRequester
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()

	qb.From("projecta_expenses")
	qb.Join("projecta_projects", "projecta_projects.project_id = projecta_expenses.project_id")
	qb.Join("projecta_cost_types", "projecta_cost_types.type_id = projecta_expenses.type_id")
	qb.Join("people", "people.person_id = projecta_expenses.owner_id")
	qb.Join("projecta_cost_categories", "projecta_cost_categories.category_id = projecta_cost_types.category_id")

	qb.Select(
		"projecta_expenses.expense_id",
		"projecta_expenses.project_id",
		"projecta_projects.name as project_name",
		"projecta_cost_categories.category_id",
		"projecta_cost_categories.name as category_name",
		"projecta_cost_types.type_id",
		"projecta_cost_types.name as type_name",
		"projecta_expenses.amount",
		"projecta_expenses.currency",
		"projecta_expenses.description",
		"projecta_expenses.owner_id",
		"people.first_name",
		"COALESCE(people.display_name, '') display_name",
		"COALESCE(projecta_expenses.expense_date, projecta_expenses.created_at) expense_date",
	)

	if filter.ProjectID != uuid.Nil {
		qb.Where(qb.Equal("projecta_expenses.project_id", filter.ProjectID.String()))
	}

	if filter.ExpenseID != uuid.Nil {
		qb.Where(qb.Equal("expense_id", filter.ExpenseID.String()))
	}

	qb.Where(qb.Equal("projecta_projects.owner_id", personID.String()))

	sql, args := qb.Build()

	var (
		expenseID    string
		projectID    string
		projectName  string
		categoryID   string
		categoryName string
		typeID       string
		typeName     string
		amount       int64
		currency     string
		description  string
		ownerID      string
		firstName    string
		displayName  string
		expenseDate  time.Time
	)

	if err := r.db.QueryRow(
		ctx,
		sql,
		args...,
	).Scan(
		&expenseID,
		&projectID,
		&projectName,
		&categoryID,
		&categoryName,
		&typeID,
		&typeName,
		&amount,
		&currency,
		&description,
		&ownerID,
		&firstName,
		&displayName,
		&expenseDate,
	); err != nil {
		return nil, err
	}

	return toExpense(
		expenseID,
		projectID,
		projectName,
		categoryID,
		categoryName,
		typeID,
		typeName,
		amount,
		currency,
		description,
		ownerID,
		firstName,
		displayName,
		expenseDate,
	), nil
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

	res, err := r.db.Exec(ctx, sql, args...)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return r.create(ctx, expense)
	} else {
		return r.update(ctx, expense)
	}
}

func (r *PgProjectaExpenseRepository) create(ctx context.Context, expense *projecta.Expense) error {
	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
	qb.InsertInto("projecta_expenses")
	qb.Cols(
		"expense_id",
		"project_id",
		"type_id",
		"amount",
		"currency",
		"description",
		"owner_id",
		"expense_date",
	)

	qb.Values(
		expense.ID.String(),
		expense.Project.ProjectID.String(),
		expense.Type.ID.String(),
		expense.Amount.Amount(),
		expense.Amount.Currency().Code,
		expense.Description,
		expense.Owner.PersonID.String(),
		expense.Date,
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
		qb.Assign("type_id", expense.Type.ID.String()),
		qb.Assign("amount", expense.Amount.Amount()),
		qb.Assign("currency", expense.Amount.Currency().Code),
		qb.Assign("description", expense.Description),
		qb.Assign("expense_date", expense.Date),
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

func (r *PgProjectaExpenseRepository) Find(ctx context.Context, filter projecta.ExpenseCollectionFilter) (*projecta.ExpenseCollection, error) {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return nil, core.FailedToIdentifyRequester
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_expenses")
	qb.Join("projecta_projects", "projecta_projects.project_id = projecta_expenses.project_id")
	qb.Join("projecta_cost_types", "projecta_cost_types.type_id = projecta_expenses.type_id")
	qb.Join("people", "people.person_id = projecta_expenses.owner_id")
	qb.Join("projecta_cost_categories", "projecta_cost_categories.category_id = projecta_cost_types.category_id")

	qb.Where(qb.Equal("projecta_expenses.owner_id", personID.String()))
	qb.Where(qb.Equal("projecta_expenses.project_id", filter.ProjectID.String()))

	if filter.CategoryID != uuid.Nil {
		qb.Where(qb.Equal("projecta_expenses.category_id", filter.CategoryID.String()))
	}

	if filter.TypeID != uuid.Nil {
		qb.Where(qb.Equal("projecta_expenses.type_id", filter.TypeID.String()))
	}

	qb.Select(qb.As("COUNT(*)", "total"))

	sql, args := qb.Build()

	var total int

	if err = r.db.QueryRow(ctx, sql, args...).Scan(&total); err != nil {
		return nil, err
	}

	qb.Select() // reset select

	if filter.Limit == 0 {
		filter.Limit = core.DefaultLimit
	}

	qb.Limit(filter.Limit)
	qb.Offset(filter.Offset)

	if filter.OrderBy != "" && filter.Order != "" {
		qb.OrderBy(fmt.Sprintf("projecta_expenses.%s %s", filter.OrderBy, filter.Order.String()))
	} else {
		qb.OrderBy("projecta_expenses.expense_date DESC")
	}

	qb.Select(
		"projecta_expenses.expense_id",
		"projecta_expenses.project_id",
		"projecta_projects.name as project_name",
		"projecta_cost_categories.category_id",
		"projecta_cost_categories.name as category_name",
		"projecta_cost_types.type_id",
		"projecta_cost_types.name as type_name",
		"projecta_expenses.amount",
		"projecta_expenses.currency",
		"projecta_expenses.description",
		"projecta_expenses.owner_id",
		"people.first_name",
		"COALESCE(people.display_name, '') display_name",
		"COALESCE(projecta_expenses.expense_date, projecta_expenses.created_at) expense_date",
	)

	sql, args = qb.Build()

	rows, err := r.db.Query(ctx, sql, args...)

	if err != nil {
		return nil, err
	}

	collection := projecta.NewExpenseCollection(total)

	defer rows.Close()

	for rows.Next() {
		var (
			expenseID    string
			projectID    string
			projectName  string
			categoryID   string
			categoryName string
			typeID       string
			typeName     string
			amount       int64
			currency     string
			description  string
			ownerID      string
			firstName    string
			displayName  string
			expenseDate  time.Time
		)
		err = rows.Scan(
			&expenseID,
			&projectID,
			&projectName,
			&categoryID,
			&categoryName,
			&typeID,
			&typeName,
			&amount,
			&currency,
			&description,
			&ownerID,
			&firstName,
			&displayName,
			&expenseDate,
		)

		if err != nil {
			return nil, err
		}

		expense := toExpense(
			expenseID,
			projectID,
			projectName,
			categoryID,
			categoryName,
			typeID,
			typeName,
			amount,
			currency,
			description,
			ownerID,
			firstName,
			displayName,
			expenseDate,
		)

		collection.Add(expense)
	}

	return collection, nil
}

func toExpense(
	expenseID string,
	projectID string,
	projectName string,
	categoryID string,
	categoryName string,
	typeID string,
	typeName string,
	amount int64,
	currency string,
	description string,
	ownerID string,
	firstName string,
	displayName string,
	expenseDate time.Time,
) *projecta.Expense {
	projectUUID := uuid.MustParse(projectID)
	person := &projecta.Owner{
		PersonID:    uuid.MustParse(ownerID),
		FirstName:   firstName,
		DisplayName: displayName,
	}

	project, _ := projecta.NewProject(
		projectUUID,
		projectName,
		"",
		person,
		time.Time{},
		time.Time{})

	category, _ := projecta.NewCostCategory(
		uuid.MustParse(categoryID),
		projectUUID,
		categoryName,
		"",
	)

	costType := &projecta.CostType{
		ID:          uuid.MustParse(typeID),
		ProjectID:   projectUUID,
		Category:    category,
		Name:        typeName,
		Description: "",
	}

	amountMoney := money.New(amount, currency)

	expenseUUID := uuid.MustParse(expenseID)

	expense := projecta.NewExpense(
		expenseUUID,
		project,
		person,
		costType,
		description,
		amountMoney,
		expenseDate,
	)

	return expense
}
