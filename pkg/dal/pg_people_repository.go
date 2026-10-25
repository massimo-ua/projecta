package dal

import (
	"context"
	types "database/sql"
	"errors"

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

type personRecord struct {
	id          string
	firstName   string
	lastName    string
	displayName types.NullString
}

func (r *personRecord) toPerson(roles []people.Role, credentials []people.Credentials) (*people.Person, error) {
	pID, err := core.ParsePersonID(r.id)
	if err != nil {
		return nil, exceptions.NewInternalException("invalid person id", err)
	}

	dn, err := people.NewDisplayName(r.displayName.String)
	if err != nil {
		return nil, exceptions.NewInternalException("invalid display name", err)
	}

	return people.ReconstitutePerson(pID, r.firstName, r.lastName, dn, credentials, roles), nil
}

func (r *PgPeopleRepository) Save(ctx context.Context, person *people.Person) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO people (person_id, first_name, last_name, display_name)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (person_id) DO UPDATE
		 SET first_name = EXCLUDED.first_name,
		     last_name = EXCLUDED.last_name,
		     display_name = EXCLUDED.display_name,
		     updated_at = now()`,
		person.ID().String(),
		person.FirstName(),
		person.LastName(),
		person.DisplayNameVO().String(),
	)
	if err != nil {
		return exceptions.NewInternalException(failedToRegisterPersonError, err)
	}

	if err := r.saveCredentials(ctx, person.ID(), person.Identities()); err != nil {
		return exceptions.NewInternalException(failedToRegisterPersonError, err)
	}

	if err := r.saveRoles(ctx, person.ID(), person.Roles()); err != nil {
		return exceptions.NewInternalException(failedToRegisterPersonError, err)
	}

	return nil
}

func (r *PgPeopleRepository) FindCredentials(
	ctx context.Context,
	provider people.IdentityProvider,
	registrationID string,
) (core.PersonID, string, error) {
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
		return core.NilPersonID, "", exceptions.NewNotFoundException("credentials not found", err)
	}

	personTyped, err := core.ParsePersonID(personID)
	if err != nil {
		return core.NilPersonID, "", exceptions.NewInternalException("failed to fetch person id", err)
	}

	return personTyped, identity, nil
}

func (r *PgPeopleRepository) saveCredentials(
	ctx context.Context,
	personID core.PersonID,
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
	sql += ` ON CONFLICT (provider, person_id) DO UPDATE SET identity = EXCLUDED.identity, registration_id = EXCLUDED.registration_id`

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (r *PgPeopleRepository) fetchRoles(ctx context.Context, personID core.PersonID) ([]people.Role, error) {
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *PgPeopleRepository) fetchCredentials(ctx context.Context, personID core.PersonID) ([]people.Credentials, error) {
	rows, err := r.db.Query(ctx, `SELECT provider, identity, registration_id FROM credentials WHERE person_id = $1`, personID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []people.Credentials
	for rows.Next() {
		var provider, identity, registrationID string
		if err := rows.Scan(&provider, &identity, &registrationID); err != nil {
			return nil, err
		}
		c, err := people.NewCredentials(provider, registrationID, identity)
		if err == nil {
			creds = append(creds, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return creds, nil
}

func (r *PgPeopleRepository) FindByID(ctx context.Context, personID core.PersonID) (*people.Person, error) {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("people")
	qb.Select("first_name", "last_name", "display_name")
	qb.Where(qb.Equal("person_id", personID.String()))

	sql, args := qb.Build()

	rec := personRecord{id: personID.String()}

	if err := r.db.QueryRow(
		ctx,
		sql,
		args...,
	).Scan(
		&rec.firstName,
		&rec.lastName,
		&rec.displayName,
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

	creds, err := r.fetchCredentials(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to fetch person credentials", err)
	}

	return rec.toPerson(roles, creds)
}

func (r *PgPeopleRepository) saveRoles(ctx context.Context, personID core.PersonID, roles []people.Role) error {
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
		var rec personRecord
		var personID string
		if err := rows.Scan(&personID, &rec.firstName, &rec.lastName, &rec.displayName); err != nil {
			return nil, 0, exceptions.NewInternalException("failed to scan person", err)
		}
		rec.id = personID
		pID, err := core.ParsePersonID(personID)
		if err != nil {
			return nil, 0, exceptions.NewInternalException("invalid person id", err)
		}
		roles, err := r.fetchRoles(ctx, pID)
		if err != nil {
			return nil, 0, exceptions.NewInternalException("failed to fetch person roles", err)
		}
		creds, err := r.fetchCredentials(ctx, pID)
		if err != nil {
			return nil, 0, exceptions.NewInternalException("failed to fetch person credentials", err)
		}
		p, err := rec.toPerson(roles, creds)
		if err != nil {
			return nil, 0, exceptions.NewInternalException("failed to build person", err)
		}
		list = append(list, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, exceptions.NewInternalException("failed to iterate people", err)
	}
	return list, total, nil
}

func (r *PgPeopleRepository) DeletePerson(ctx context.Context, personID core.PersonID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM people WHERE person_id = $1`, personID.String())
	return err
}
