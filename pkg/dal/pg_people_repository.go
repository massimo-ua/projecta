package dal

import (
	"context"
	types "database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
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
	qb.Cols("person_id", "first_name", "last_name")
	qb.Values(person.ID().String(), person.FirstName, person.LastName)

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

	person, err := toPersonFromPg(personID.String(), firstName, lastName, displayName.String)

	if err != nil {
		return nil, exceptions.NewInternalException("failed to fetch person information", err)
	}

	return &person, nil
}

func toPersonFromPg(personID string, personFirstName string, personLastName string, personDisplayName string) (people.Person, error) {
	p, err := people.NewPerson(uuid.MustParse(personID), personFirstName, personLastName, personDisplayName, nil)

	if err != nil {
		return people.Person{}, err
	}
	return *p, nil
}
