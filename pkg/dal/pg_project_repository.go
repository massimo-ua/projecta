package dal

import (
	"context"
	types "database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"time"
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
		"projecta_projects.project_id",
		"name",
		"description",
		"owner_id",
		"started_at",
		"ended_at",
		"people.first_name",
		"people.last_name",
		"people.display_name",
		"projecta_projects.share_token",
	)
	qb.Join("people", "people.person_id = projecta_projects.owner_id")

	if filter.ProjectID != uuid.Nil {
		qb.Where(qb.Equal("projecta_projects.project_id", filter.ProjectID.String()))
	}

	if filter.Name != "" {
		qb.Where(qb.Equal("name", filter.Name))
	}

	qb.Where(fmt.Sprintf("(projecta_projects.owner_id = %s OR projecta_projects.project_id IN (SELECT project_id FROM projecta_project_shares WHERE person_id = %s))", qb.Var(personID.String()), qb.Var(personID.String())))

	sql, args := qb.Build()

	var (
		projectID     string
		name          string
		description   string
		ownerID       string
		startedAt     time.Time
		endedAt       time.Time
		firstName     string
		lastName      string
		displayName   types.NullString
		shareTokenStr string
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
		&startedAt,
		&endedAt,
		&firstName,
		&lastName,
		&displayName,
		&shareTokenStr,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, exceptions.NewNotFoundException("project not found", err)
		}

		return nil, err
	}

	p, err := toProject(projectID, name, description, ownerID, firstName, lastName, displayName.String, startedAt, endedAt, shareTokenStr)
	if err != nil {
		return nil, err
	}
	p.IsShared = (p.Owner.PersonID != personID)
	return p, nil
}

func (r *PgProjectRepository) FindByShareToken(ctx context.Context, token uuid.UUID) (*projecta.Project, error) {
	if token == uuid.Nil {
		return nil, exceptions.NewValidationException("invalid share token", nil)
	}

	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_projects")
	qb.Select(
		"projecta_projects.project_id",
		"name",
		"description",
		"owner_id",
		"started_at",
		"ended_at",
		"people.first_name",
		"people.last_name",
		"people.display_name",
		"projecta_projects.share_token",
	)
	qb.Join("people", "people.person_id = projecta_projects.owner_id")
	qb.Where(qb.Equal("share_token", token.String()))

	sql, args := qb.Build()

	var (
		projectID     string
		name          string
		description   string
		ownerID       string
		startedAt     time.Time
		endedAt       time.Time
		firstName     string
		lastName      string
		displayName   types.NullString
		shareTokenStr string
	)

	if err := r.db.QueryRow(ctx, sql, args...).Scan(
		&projectID,
		&name,
		&description,
		&ownerID,
		&startedAt,
		&endedAt,
		&firstName,
		&lastName,
		&displayName,
		&shareTokenStr,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, exceptions.NewNotFoundException("project not found", err)
		}
		return nil, err
	}

	return toProject(projectID, name, description, ownerID, firstName, lastName, displayName.String, startedAt, endedAt, shareTokenStr)
}

func (r *PgProjectRepository) CreateShareRecord(ctx context.Context, projectID uuid.UUID, personID uuid.UUID) error {
	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
	qb.InsertInto("projecta_project_shares")
	qb.Cols("share_id", "project_id", "person_id")
	qb.Values(uuid.New().String(), projectID.String(), personID.String())

	sql, args := qb.Build()
	sql += " ON CONFLICT (project_id, person_id) DO NOTHING"

	_, err := r.db.Exec(ctx, sql, args...)
	return err
}

func (r *PgProjectRepository) Create(ctx context.Context, project *projecta.Project) error {
	if project.ShareToken == uuid.Nil {
		project.ShareToken = uuid.New()
	}

	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
	qb.InsertInto("projecta_projects")
	qb.Cols(
		"project_id",
		"name",
		"description",
		"owner_id",
		"started_at",
		"ended_at",
		"share_token",
	)
	qb.Values(
		project.ProjectID.String(),
		project.Name,
		project.Description,
		project.Owner.PersonID.String(),
		project.StartDate,
		project.EndDate,
		project.ShareToken.String(),
	)

	sql, args := qb.Build()

	_, err := r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgProjectRepository) Update(ctx context.Context, project *projecta.Project) error {
	qb := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	qb.Update("projecta_projects")
	qb.Set(
		qb.Assign("name", project.Name),
		qb.Assign("description", project.Description),
		qb.Assign("owner_id", project.Owner.PersonID.String()),
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
		"projecta_projects.project_id",
		"name",
		"description",
		"owner_id",
		"started_at",
		"ended_at",
		"people.first_name",
		"people.last_name",
		"people.display_name",
		"projecta_projects.share_token",
	)
	qb.Join("people", "people.person_id = projecta_projects.owner_id")

	if filter.Name != "" {
		qb.Where(qb.Like("name", filter.Name))
	}

	qb.Where(fmt.Sprintf("(projecta_projects.owner_id = %s OR projecta_projects.project_id IN (SELECT project_id FROM projecta_project_shares WHERE person_id = %s))", qb.Var(personID.String()), qb.Var(personID.String())))

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
			projectID     string
			name          string
			description   string
			ownerID       string
			startedAt     time.Time
			endedAt       time.Time
			firstName     string
			lastName      string
			displayName   types.NullString
			shareTokenStr string
		)

		if err = rows.Scan(
			&projectID,
			&name,
			&description,
			&ownerID,
			&startedAt,
			&endedAt,
			&firstName,
			&lastName,
			&displayName,
			&shareTokenStr,
		); err != nil {
			return nil, err
		}

		p, err := toProject(projectID, name, description, ownerID, firstName, lastName, displayName.String, startedAt, endedAt, shareTokenStr)

		if err != nil {
			return nil, err
		}

		p.IsShared = (p.Owner.PersonID != personID)

		projects = append(projects, p)
	}

	return projects, nil
}

func toProject(projectID, name, description, ownerID, firstName, lastName, displayName string, startedAt, enddedAt time.Time, shareToken ...string) (*projecta.Project, error) {
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

	if len(shareToken) > 0 && shareToken[0] != "" {
		if token, err := uuid.Parse(shareToken[0]); err == nil {
			p.ShareToken = token
		}
	}

	return p, nil
}
