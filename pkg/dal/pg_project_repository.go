package dal

import (
	"context"
	types "database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

type PgProjectRepository struct {
	db *PgRepository
}

func NewPgProjectRepository(db *PgDbConnection) *PgProjectRepository {
	return &PgProjectRepository{
		db: &PgRepository{db},
	}
}

func (r *PgProjectRepository) FindOne(ctx context.Context, filter projecta.ProjectFilter) (*projecta.Project, error) {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return nil, err
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_projects")
	qb.Select(
		"project_id",
		"name",
		"description",
		"owner_id",
		"main_currency",
		"started_at",
		"ended_at",
		"people.first_name",
		"people.last_name",
		"people.display_name",
	)
	qb.Join("people", "people.person_id = projecta_projects.owner_id")

	if filter.ProjectID != uuid.Nil {
		qb.Where(qb.Equal("project_id", filter.ProjectID.String()))
	}

	if filter.Name != "" {
		qb.Where(qb.Equal("name", filter.Name))
	}

	qb.Where(qb.Equal("owner_id", personID.String()))

	sql, args := qb.Build()

	var (
		projectID    string
		name         string
		description  string
		ownerID      string
		mainCurrency string
		startedAt    time.Time
		endedAt      time.Time
		firstName    string
		lastName     string
		displayName  types.NullString
	)

	if err := r.db.QueryRow(
		ctx,
		sql,
		args...,
	).Scan(
		&projectID,
		&name,
		&description,
		&ownerID,
		&mainCurrency,
		&startedAt,
		&endedAt,
		&firstName,
		&lastName,
		&displayName,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, exceptions.NewNotFoundException("project not found", err)
		}

		return nil, err
	}

	return toProject(projectID, name, description, ownerID, mainCurrency, firstName, lastName, displayName.String, startedAt, endedAt)
}

func (r *PgProjectRepository) Create(ctx context.Context, project *projecta.Project) error {
	mainCurrency := project.MainCurrency
	if mainCurrency == "" {
		mainCurrency = "UAH"
	}

	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
	qb.InsertInto("projecta_projects")
	qb.Cols(
		"project_id",
		"name",
		"description",
		"owner_id",
		"main_currency",
		"started_at",
		"ended_at",
	)
	qb.Values(
		project.ProjectID.String(),
		project.Name,
		project.Description,
		project.Owner.PersonID.String(),
		mainCurrency,
		project.StartDate,
		project.EndDate,
	)

	sql, args := qb.Build()

	_, err := r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgProjectRepository) Update(ctx context.Context, project *projecta.Project) error {
	mainCurrency := project.MainCurrency
	if mainCurrency == "" {
		mainCurrency = "UAH"
	}

	qb := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	qb.Update("projecta_projects")
	qb.Set(
		qb.Assign("name", project.Name),
		qb.Assign("description", project.Description),
		qb.Assign("owner_id", project.Owner.PersonID.String()),
		qb.Assign("main_currency", mainCurrency),
		qb.Assign("started_at", project.StartDate),
		qb.Assign("ended_at", project.EndDate),
	)
	qb.Where(qb.Equal("project_id", project.ProjectID.String()))
	qb.Where(qb.Equal("owner_id", project.Owner.PersonID.String()))

	sql, args := qb.Build()

	_, err := r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgProjectRepository) Remove(ctx context.Context, project *projecta.Project) error {
	qb := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	qb.DeleteFrom("projecta_projects")
	qb.Where(qb.Equal("project_id", project.ProjectID.String()))
	qb.Where(qb.Equal("owner_id", project.Owner.PersonID.String()))

	sql, args := qb.Build()

	_, err := r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgProjectRepository) Find(ctx context.Context, filter projecta.ProjectCollectionFilter) ([]*projecta.Project, error) {
	personID, err := core.AuthGuard(ctx)

	if err != nil {
		return nil, err
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_projects")
	qb.Select(
		"project_id",
		"name",
		"description",
		"owner_id",
		"main_currency",
		"started_at",
		"ended_at",
		"people.first_name",
		"people.last_name",
		"people.display_name",
	)
	qb.Join("people", "people.person_id = projecta_projects.owner_id")

	if filter.Name != "" {
		qb.Where(qb.Like("name", filter.Name))
	}

	qb.Where(qb.Equal("owner_id", personID.String()))

	qb.Limit(filter.Limit)
	qb.Offset(filter.Offset)

	sql, args := qb.Build()

	rows, err := r.db.Query(ctx, sql, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var projects []*projecta.Project

	for rows.Next() {
		var (
			projectID    string
			name         string
			description  string
			ownerID      string
			mainCurrency string
			startedAt    time.Time
			endedAt      time.Time
			firstName    string
			lastName     string
			displayName  types.NullString
		)

		if err = rows.Scan(
			&projectID,
			&name,
			&description,
			&ownerID,
			&mainCurrency,
			&startedAt,
			&endedAt,
			&firstName,
			&lastName,
			&displayName,
		); err != nil {
			return nil, err
		}

		p, err := toProject(projectID, name, description, ownerID, mainCurrency, firstName, lastName, displayName.String, startedAt, endedAt)

		if err != nil {
			return nil, err
		}

		projects = append(projects, p)
	}

	return projects, nil
}

func toProject(projectID, name, description, ownerID, mainCurrency, firstName, lastName, displayName string, startedAt, enddedAt time.Time) (*projecta.Project, error) {
	person, err := people.NewPerson(
		uuid.MustParse(ownerID),
		firstName,
		lastName,
		displayName,
		nil,
	)

	if err != nil {
		return nil, err
	}

	p, err := projecta.NewProject(
		uuid.MustParse(projectID),
		name,
		description,
		&projecta.Owner{
			PersonID:    person.ID(),
			DisplayName: person.DisplayName(),
		},
		startedAt,
		enddedAt,
	)
	if err != nil {
		return nil, err
	}

	if mainCurrency != "" {
		p.MainCurrency = mainCurrency
	}

	return p, nil
}
