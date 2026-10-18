package dal

import (
	"context"
	types "database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

type projectRecord struct {
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
	mainCurrency  types.NullString
}

func (r *projectRecord) scanFields() []any {
	return []any{
		&r.projectID,
		&r.name,
		&r.description,
		&r.ownerID,
		&r.startedAt,
		&r.endedAt,
		&r.firstName,
		&r.lastName,
		&r.displayName,
		&r.shareTokenStr,
		&r.mainCurrency,
	}
}

func (r *projectRecord) toProject(isShared bool, sharedParticipants []project.Participant) (*project.Project, error) {
	pID, err := core.ParseProjectID(r.projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project id: %w", err)
	}
	oID, err := core.ParsePersonID(r.ownerID)
	if err != nil {
		return nil, fmt.Errorf("invalid owner id: %w", err)
	}

	dispName := r.displayName.String
	if dispName == "" {
		dispName = strings.TrimSpace(r.firstName + " " + r.lastName)
	}

	owner := &project.Owner{
		PersonID:    oID,
		DisplayName: dispName,
	}

	var participants []project.Participant
	if dispName != "" {
		if op, err := project.NewParticipant(dispName); err == nil {
			participants = append(participants, op)
		}
	}
	for _, sp := range sharedParticipants {
		duplicate := false
		for _, ep := range participants {
			if ep.DisplayName() == sp.DisplayName() {
				duplicate = true
				break
			}
		}
		if !duplicate {
			participants = append(participants, sp)
		}
	}

	shareToken, _ := uuid.Parse(r.shareTokenStr)
	curr := r.mainCurrency.String
	if curr == "" {
		curr = "UAH"
	}

	return project.ReconstituteProject(
		pID,
		r.name,
		r.description,
		owner,
		r.startedAt,
		r.endedAt,
		shareToken,
		isShared,
		curr,
		participants,
	), nil
}

type PgProjectRepository struct {
	db *PgRepository
}

func NewPgProjectRepository(db *PgDbConnection) *PgProjectRepository {
	return &PgProjectRepository{
		db: &PgRepository{db},
	}
}

func (r *PgProjectRepository) FindOne(ctx context.Context, filter project.ProjectFilter) (*project.Project, error) {
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
		"projecta_projects.main_currency",
	)
	qb.Join("people", "people.person_id = projecta_projects.owner_id")

	if !filter.ProjectID.IsNil() {
		qb.Where(qb.Equal("projecta_projects.project_id", filter.ProjectID.String()))
	}

	if filter.Name != "" {
		qb.Where(qb.Equal("name", filter.Name))
	}

	if !filter.RequesterID.IsNil() {
		qb.Where(fmt.Sprintf("(projecta_projects.owner_id = %s OR projecta_projects.project_id IN (SELECT project_id FROM projecta_project_shares WHERE person_id = %s))", qb.Var(filter.RequesterID.String()), qb.Var(filter.RequesterID.String())))
	}

	sql, args := qb.Build()

	var rec projectRecord
	if err := r.db.QueryRow(ctx, sql, args...).Scan(rec.scanFields()...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, exceptions.NewNotFoundException("project not found", err)
		}

		return nil, err
	}

	pID, err := core.ParseProjectID(rec.projectID)
	if err != nil {
		return nil, err
	}

	sharedParts, err := r.fetchParticipants(ctx, pID)
	if err != nil {
		return nil, err
	}

	isShared := (!filter.RequesterID.IsNil() && rec.ownerID != filter.RequesterID.String())
	return rec.toProject(isShared, sharedParts)
}

func (r *PgProjectRepository) FindByShareToken(ctx context.Context, token uuid.UUID) (*project.Project, error) {
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
		"projecta_projects.main_currency",
	)
	qb.Join("people", "people.person_id = projecta_projects.owner_id")
	qb.Where(qb.Equal("share_token", token.String()))

	sql, args := qb.Build()

	var rec projectRecord
	if err := r.db.QueryRow(ctx, sql, args...).Scan(rec.scanFields()...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, exceptions.NewNotFoundException("project not found", err)
		}
		return nil, err
	}

	pID, err := core.ParseProjectID(rec.projectID)
	if err != nil {
		return nil, err
	}

	sharedParts, err := r.fetchParticipants(ctx, pID)
	if err != nil {
		return nil, err
	}

	return rec.toProject(false, sharedParts)
}

func (r *PgProjectRepository) fetchParticipants(ctx context.Context, projectID core.ProjectID) ([]project.Participant, error) {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_project_shares")
	qb.Select(
		"people.first_name",
		"people.last_name",
		"people.display_name",
	)
	qb.Join("people", "people.person_id = projecta_project_shares.person_id")
	qb.Where(qb.Equal("projecta_project_shares.project_id", projectID.String()))
	qb.OrderBy("projecta_project_shares.created_at").Asc()

	sql, args := qb.Build()

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []project.Participant
	for rows.Next() {
		var (
			firstName   string
			lastName    string
			displayName types.NullString
		)

		if err := rows.Scan(&firstName, &lastName, &displayName); err != nil {
			return nil, err
		}

		name := displayName.String
		if name == "" {
			name = strings.TrimSpace(firstName + " " + lastName)
		}

		if name != "" {
			participant, err := project.NewParticipant(name)
			if err == nil {
				participants = append(participants, participant)
			}
		}
	}

	return participants, rows.Err()
}

func (r *PgProjectRepository) CreateShareRecord(ctx context.Context, projectID core.ProjectID, personID core.PersonID) error {
	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
	qb.InsertInto("projecta_project_shares")
	qb.Cols("share_id", "project_id", "person_id")
	qb.Values(uuid.New().String(), projectID.String(), personID.String())

	sql, args := qb.Build()
	sql += " ON CONFLICT (project_id, person_id) DO NOTHING"

	_, err := r.db.Exec(ctx, sql, args...)
	return err
}

func (r *PgProjectRepository) Create(ctx context.Context, project *project.Project) error {
	mainCurrency := project.MainCurrency()
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
		"started_at",
		"ended_at",
		"share_token",
		"main_currency",
	)
	qb.Values(
		project.ID().String(),
		project.Name(),
		project.Description(),
		project.OwnerID().String(),
		project.StartDate(),
		project.EndDate(),
		project.ShareToken().String(),
		mainCurrency,
	)

	sql, args := qb.Build()

	_, err := r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgProjectRepository) Update(ctx context.Context, project *project.Project) error {
	mainCurrency := project.MainCurrency()
	if mainCurrency == "" {
		mainCurrency = "UAH"
	}

	qb := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	qb.Update("projecta_projects")
	qb.Set(
		qb.Assign("name", project.Name()),
		qb.Assign("description", project.Description()),
		qb.Assign("owner_id", project.OwnerID().String()),
		qb.Assign("started_at", project.StartDate()),
		qb.Assign("ended_at", project.EndDate()),
		qb.Assign("main_currency", mainCurrency),
	)
	qb.Where(qb.Equal("project_id", project.ID().String()))
	qb.Where(qb.Equal("owner_id", project.OwnerID().String()))

	sql, args := qb.Build()

	_, err := r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgProjectRepository) Remove(ctx context.Context, project *project.Project) error {
	qb := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	qb.DeleteFrom("projecta_projects")
	qb.Where(qb.Equal("project_id", project.ID().String()))
	qb.Where(qb.Equal("owner_id", project.OwnerID().String()))

	sql, args := qb.Build()

	_, err := r.db.Exec(ctx, sql, args...)

	return err
}

func (r *PgProjectRepository) Find(ctx context.Context, filter project.ProjectCollectionFilter) ([]*project.Project, error) {
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
		"projecta_projects.main_currency",
	)
	qb.Join("people", "people.person_id = projecta_projects.owner_id")

	if filter.Name != "" {
		qb.Where(qb.Like("name", filter.Name))
	}

	if !filter.RequesterID.IsNil() {
		qb.Where(fmt.Sprintf("(projecta_projects.owner_id = %s OR projecta_projects.project_id IN (SELECT project_id FROM projecta_project_shares WHERE person_id = %s))", qb.Var(filter.RequesterID.String()), qb.Var(filter.RequesterID.String())))
	}

	qb.Limit(filter.Limit)
	qb.Offset(filter.Offset)

	sql, args := qb.Build()

	rows, err := r.db.Query(ctx, sql, args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var projects []*project.Project

	for rows.Next() {
		var rec projectRecord

		if err = rows.Scan(rec.scanFields()...); err != nil {
			return nil, err
		}

		isShared := (!filter.RequesterID.IsNil() && rec.ownerID != filter.RequesterID.String())
		p, err := rec.toProject(isShared, nil)
		if err != nil {
			return nil, err
		}

		projects = append(projects, p)
	}

	return projects, nil
}
