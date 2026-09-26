package dal

import (
	"context"
	types "database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/people"
)

type PgPeopleRepository struct {
	db *PgRepository
}

var failedToRegisterPersonError = "failed to register person"

func NewPgPeopleRepository(db *PgDbConnection) *PgPeopleRepository {
	return &PgPeopleRepository{
		db: &PgRepository{db},
	}
}

func (r *PgPeopleRepository) Register(ctx context.Context, person *people.Person) error {
	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
	qb.InsertInto("people")
	qb.Cols("person_id", "first_name", "last_name", "display_name")
	qb.Values(person.ID().String(), person.FirstName, person.LastName, person.DisplayName())

	sql, args := qb.Build()

	if _, err := r.db.Exec(
		ctx,
		sql,
		args...,
	); err != nil {
		return exceptions.NewInternalException(failedToRegisterPersonError, err)
	}

	if err := r.setCredentials(ctx, person.ID(), person.Identities()); err != nil {
		return exceptions.NewInternalException(failedToRegisterPersonError, err)
	}

	if len(person.Roles()) > 0 {
		if err := r.SaveRoles(ctx, person.ID(), person.Roles()); err != nil {
			return exceptions.NewInternalException(failedToRegisterPersonError, err)
		}
	}

	return nil
}

func (r *PgPeopleRepository) FindCredentials(
	ctx context.Context,
	provider people.IdentityProvider,
	registrationID string,
) (uuid.UUID, string, error) {
	var personID string
	var identity string

	err := r.db.QueryRow(
		ctx,
		`SELECT
				"person_id", "identity"
				FROM "credentials"
				WHERE "provider" = $1 AND "registration_id" = $2`,
		provider,
		registrationID,
	).Scan(
		&personID,
		&identity)

	if err != nil {
		return uuid.Nil, "", exceptions.NewNotFoundException("credentials not found", err)
	}

	personUUID, err := uuid.Parse(personID)

	if err != nil {
		return uuid.Nil, "", exceptions.NewInternalException("failed to fetch person id", err)
	}

	return personUUID, identity, nil
}

func (r *PgPeopleRepository) setCredentials(
	ctx context.Context,
	personID uuid.UUID,
	credentials []people.Credentials,
) error {
	if len(credentials) == 0 {
		return nil
	}

	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
	qb.InsertInto("credentials")
	qb.Cols("person_id", "provider", "identity", "registration_id")

	for _, i := range credentials {
		qb.Values(personID.String(), i.Provider(), i.Identifier(), i.RegistrationID())
	}

	sql, args := qb.Build()

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (r *PgPeopleRepository) FindByID(ctx context.Context, personID uuid.UUID) (*people.Person, error) {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("people")
	qb.Select("first_name", "last_name", "display_name")
	qb.Where(qb.Equal("person_id", personID.String()))

	sql, args := qb.Build()

	var (
		firstName   string
		lastName    string
		displayName types.NullString
	)

	if err := r.db.QueryRow(
		ctx,
		sql,
		args...,
	).Scan(
		&firstName,
		&lastName,
		&displayName,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, exceptions.NewNotFoundException("person not found", err)
		}

		return nil, err
	}

	roles, err := r.fetchRoles(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to fetch person roles", err)
	}

	person, err := toPersonFromPg(personID.String(), firstName, lastName, displayName.String, roles...)

	if err != nil {
		return nil, exceptions.NewInternalException("failed to fetch person information", err)
	}

	return &person, nil
}

func (r *PgPeopleRepository) fetchRoles(ctx context.Context, personID uuid.UUID) ([]people.Role, error) {
	rows, err := r.db.Query(ctx, `SELECT role FROM person_roles WHERE person_id = $1`, personID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []people.Role
	for rows.Next() {
		var roleStr string
		if err := rows.Scan(&roleStr); err != nil {
			return nil, err
		}
		if role, err := people.ToRole(roleStr); err == nil {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

func (r *PgPeopleRepository) SaveRoles(ctx context.Context, personID uuid.UUID, roles []people.Role) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM person_roles WHERE person_id = $1`, personID.String()); err != nil {
		return err
	}

	if len(roles) == 0 {
		return nil
	}

	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()
	qb.InsertInto("person_roles")
	qb.Cols("person_id", "role")
	for _, role := range roles {
		qb.Values(personID.String(), role.String())
	}
	sql, args := qb.Build()
	sql += " ON CONFLICT DO NOTHING"

	_, err := r.db.Exec(ctx, sql, args...)
	return err
}

func (r *PgPeopleRepository) FindAll(ctx context.Context, pagination core.Pagination) ([]*people.Person, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM people WHERE deleted_at IS NULL`).Scan(&total)
	if err != nil {
		return nil, 0, exceptions.NewInternalException("failed to count people", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT person_id, first_name, last_name, display_name
		FROM people
		WHERE deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1 OFFSET $2
	`, pagination.Limit, pagination.Offset)
	if err != nil {
		return nil, 0, exceptions.NewInternalException("failed to fetch people", err)
	}
	defer rows.Close()

	var list []*people.Person
	for rows.Next() {
		var (
			personID    string
			firstName   string
			lastName    string
			displayName types.NullString
		)
		if err := rows.Scan(&personID, &firstName, &lastName, &displayName); err != nil {
			return nil, 0, exceptions.NewInternalException("failed to scan person", err)
		}
		pID, err := uuid.Parse(personID)
		if err != nil {
			return nil, 0, exceptions.NewInternalException("invalid person id", err)
		}
		roles, err := r.fetchRoles(ctx, pID)
		if err != nil {
			return nil, 0, exceptions.NewInternalException("failed to fetch person roles", err)
		}
		p, err := toPersonFromPg(personID, firstName, lastName, displayName.String, roles...)
		if err != nil {
			return nil, 0, exceptions.NewInternalException("failed to build person", err)
		}
		list = append(list, &p)
	}
	return list, total, nil
}

func toPersonFromPg(personID string, personFirstName string, personLastName string, personDisplayName string, roles ...people.Role) (people.Person, error) {
	p, err := people.NewPerson(uuid.MustParse(personID), personFirstName, personLastName, personDisplayName, nil, roles...)

	if err != nil {
		return people.Person{}, err
	}
	return *p, nil
}

func (r *PgPeopleRepository) DeletePerson(ctx context.Context, personID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM people WHERE person_id = $1`, personID.String())
	return err
}

func (r *PgPeopleRepository) UpdateProfile(ctx context.Context, personID uuid.UUID, firstName string, lastName string, displayName string) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE people SET first_name = $1, last_name = $2, display_name = $3, updated_at = now() WHERE person_id = $4`,
		firstName,
		lastName,
		displayName,
		personID.String(),
	)
	return err
}

func (r *PgPeopleRepository) SaveCredentials(ctx context.Context, personID uuid.UUID, cred people.Credentials) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO credentials (person_id, provider, identity, registration_id)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (provider, person_id) DO UPDATE
		 SET identity = EXCLUDED.identity, registration_id = EXCLUDED.registration_id`,
		personID.String(),
		cred.Provider(),
		cred.Identifier(),
		cred.RegistrationID(),
	)
	return err
}

