package dal

import (
	"context"
	"fmt"
	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"time"
)

type PgPaymentRepository struct {
	db *PgRepository
}

func NewPgPaymentRepository(db *PgDbConnection) *PgPaymentRepository {
	return &PgPaymentRepository{
		db: &PgRepository{db},
	}
}

func (r *PgPaymentRepository) FindOne(ctx context.Context, filter projecta.PaymentFilter) (*projecta.Payment, error) {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return nil, core.FailedToIdentifyRequester
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()

	qb.From("projecta_payments")
	qb.Join("projecta_projects", "projecta_projects.project_id = projecta_payments.project_id")
	qb.Join("projecta_cost_types", "projecta_cost_types.type_id = projecta_payments.type_id")
	qb.Join("people", "people.person_id = projecta_payments.owner_id")
	qb.Join("projecta_cost_categories", "projecta_cost_categories.category_id = projecta_cost_types.category_id")

	qb.Select(
		"projecta_payments.payment_id",
		"projecta_payments.project_id",
		"projecta_projects.name as project_name",
		"projecta_cost_categories.category_id",
		"projecta_cost_categories.name as category_name",
		"projecta_cost_types.type_id",
		"projecta_cost_types.name as type_name",
		"projecta_payments.amount",
		"projecta_payments.currency",
		"projecta_payments.description",
		"projecta_payments.owner_id",
		"people.first_name",
		"COALESCE(people.display_name, '') display_name",
		"COALESCE(projecta_payments.payment_date, projecta_payments.created_at) payment_date",
		"projecta_payments.kind",
	)

	if filter.ProjectID != uuid.Nil {
		qb.Where(qb.Equal("projecta_payments.project_id", filter.ProjectID.String()))
	}

	if filter.PaymentID != uuid.Nil {
		qb.Where(qb.Equal("payment_id", filter.PaymentID.String()))
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
		expenseKind  string
	)

	if err = r.db.QueryRow(
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
		&expenseKind,
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
		expenseKind,
	), nil
}

func (r *PgPaymentRepository) Save(ctx context.Context, expense *projecta.Payment) error {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return err
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_payments")
	qb.Select(
		"payment_id",
		"project_id",
		"type_id",
		"amount",
		"currency",
		"description",
		"created_at",
		"updated_at",
	)

	qb.Where(qb.Equal("payment_id", expense.ID.String()))
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

func (r *PgPaymentRepository) create(ctx context.Context, expense *projecta.Payment) error {
	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
	qb.InsertInto("projecta_payments")
	qb.Cols(
		"payment_id",
		"project_id",
		"type_id",
		"amount",
		"currency",
		"description",
		"owner_id",
		"payment_date",
		"kind",
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
		expense.Kind.String(),
	)

	sql, args := qb.Build()

	_, err := r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgPaymentRepository) update(ctx context.Context, payment *projecta.Payment) error {
	qb := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	qb.Update("projecta_payments")
	qb.Set(
		qb.Assign("type_id", payment.Type.ID.String()),
		qb.Assign("amount", payment.Amount.Amount()),
		qb.Assign("currency", payment.Amount.Currency().Code),
		qb.Assign("description", payment.Description),
		qb.Assign("payment_date", payment.Date),
		qb.Assign("kind", payment.Kind.String()),
	)

	qb.Where(qb.Equal("payment_id", payment.ID.String()))
	qb.Where(qb.Equal("project_id", payment.Project.ProjectID.String()))
	qb.Where(qb.Equal("owner_id", payment.Owner.PersonID.String()))

	sql, args := qb.Build()

	_, err := r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgPaymentRepository) Remove(ctx context.Context, expense *projecta.Payment) error {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return core.FailedToIdentifyRequester
	}

	qb := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	qb.DeleteFrom("projecta_payments")
	qb.Where(qb.Equal("payment_id", expense.ID.String()))
	qb.Where(qb.Equal("owner_id", personID.String()))
	qb.Where(qb.Equal("project_id", expense.Project.ProjectID.String()))

	sql, args := qb.Build()

	_, err = r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgPaymentRepository) Find(ctx context.Context, filter projecta.PaymentCollectionFilter) (*projecta.PaymentCollection, error) {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return nil, core.FailedToIdentifyRequester
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_payments")
	qb.Join("projecta_projects", "projecta_projects.project_id = projecta_payments.project_id")
	qb.Join("projecta_cost_types", "projecta_cost_types.type_id = projecta_payments.type_id")
	qb.Join("people", "people.person_id = projecta_payments.owner_id")
	qb.Join("projecta_cost_categories", "projecta_cost_categories.category_id = projecta_cost_types.category_id")

	qb.Where(qb.Equal("projecta_payments.owner_id", personID.String()))
	qb.Where(qb.Equal("projecta_payments.project_id", filter.ProjectID.String()))

	if filter.CategoryID != uuid.Nil {
		qb.Where(qb.Equal("projecta_payments.category_id", filter.CategoryID.String()))
	}

	if filter.TypeID != uuid.Nil {
		qb.Where(qb.Equal("projecta_payments.type_id", filter.TypeID.String()))
	}

	if filter.Kind != "" {
		qb.Where(qb.Equal("projecta_payments.kind", filter.Kind.String()))
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
		qb.OrderBy(fmt.Sprintf("projecta_payments.%s %s", filter.OrderBy, filter.Order.String()))
	} else {
		qb.OrderBy("projecta_payments.payment_date DESC")
	}

	qb.Select(
		"projecta_payments.payment_id",
		"projecta_payments.project_id",
		"projecta_projects.name as project_name",
		"projecta_cost_categories.category_id",
		"projecta_cost_categories.name as category_name",
		"projecta_cost_types.type_id",
		"projecta_cost_types.name as type_name",
		"projecta_payments.amount",
		"projecta_payments.currency",
		"projecta_payments.description",
		"projecta_payments.owner_id",
		"people.first_name",
		"COALESCE(people.display_name, '') display_name",
		"COALESCE(projecta_payments.payment_date, projecta_payments.created_at) payment_date",
		"projecta_payments.kind",
	)

	sql, args = qb.Build()

	rows, err := r.db.Query(ctx, sql, args...)

	if err != nil {
		return nil, err
	}

	collection := projecta.NewPaymentCollection(total)

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
			expenseKind  string
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
			&expenseKind,
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
			expenseKind,
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
	expenseKind string,
) *projecta.Payment {
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

	kind, _ := projecta.ToPaymentKind(expenseKind)

	expense := projecta.NewPayment(
		expenseUUID,
		project,
		person,
		costType,
		description,
		amountMoney,
		expenseDate,
		kind,
	)

	return expense
}
