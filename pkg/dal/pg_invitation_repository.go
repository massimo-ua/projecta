package dal

import (
	"context"
	types "database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/people"
)

type PgInvitationRepository struct {
	db *PgRepository
}

func NewPgInvitationRepository(db *PgDbConnection) *PgInvitationRepository {
	return &PgInvitationRepository{
		db: &PgRepository{db},
	}
}

func (r *PgInvitationRepository) Create(ctx context.Context, inv *people.Invitation) error {
	var personIDVal *string
	if inv.PersonID() != nil {
		pStr := inv.PersonID().String()
		personIDVal = &pStr
	}

	_, err := r.db.Exec(
		ctx,
		`INSERT INTO invitations (invitation_id, email, code_hash, expires_at, created_by, person_id, status, created_at, completed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		inv.ID().String(),
		inv.Email(),
		inv.CodeHash(),
		inv.ExpiresAt(),
		inv.CreatedBy().String(),
		personIDVal,
		string(inv.Status()),
		inv.CreatedAt(),
		inv.CompletedAt(),
	)
	if err != nil {
		return exceptions.NewInternalException("failed to save invitation", err)
	}
	return nil
}

func (r *PgInvitationRepository) scanInvitation(row pgx.Row) (*people.Invitation, error) {
	var (
		idStr       string
		email       string
		codeHash    string
		expiresAt   time.Time
		createdBy   string
		personID    types.NullString
		status      string
		createdAt   time.Time
		completedAt types.NullTime
	)

	err := row.Scan(
		&idStr,
		&email,
		&codeHash,
		&expiresAt,
		&createdBy,
		&personID,
		&status,
		&createdAt,
		&completedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, exceptions.NewNotFoundException("invitation not found", err)
		}
		return nil, exceptions.NewInternalException("failed to scan invitation", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, exceptions.NewInternalException("invalid invitation id", err)
	}

	createdByID, err := uuid.Parse(createdBy)
	if err != nil {
		return nil, exceptions.NewInternalException("invalid created_by id", err)
	}

	var pID *uuid.UUID
	if personID.Valid && personID.String != "" {
		parsedPID, err := uuid.Parse(personID.String)
		if err == nil {
			pID = &parsedPID
		}
	}

	var compAt *time.Time
	if completedAt.Valid {
		compAt = &completedAt.Time
	}

	return people.NewInvitation(
		id,
		email,
		codeHash,
		expiresAt,
		createdByID,
		pID,
		people.InvitationStatus(status),
		createdAt,
		compAt,
	)
}

func (r *PgInvitationRepository) FindByID(ctx context.Context, id uuid.UUID) (*people.Invitation, error) {
	row := r.db.QueryRow(
		ctx,
		`SELECT invitation_id, email, code_hash, expires_at, created_by, person_id, status, created_at, completed_at
		 FROM invitations
		 WHERE invitation_id = $1`,
		id.String(),
	)
	return r.scanInvitation(row)
}

func (r *PgInvitationRepository) FindByCodeHash(ctx context.Context, codeHash string) (*people.Invitation, error) {
	row := r.db.QueryRow(
		ctx,
		`SELECT invitation_id, email, code_hash, expires_at, created_by, person_id, status, created_at, completed_at
		 FROM invitations
		 WHERE code_hash = $1`,
		codeHash,
	)
	return r.scanInvitation(row)
}

func (r *PgInvitationRepository) FindByCreator(ctx context.Context, creatorID uuid.UUID) ([]*people.Invitation, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT invitation_id, email, code_hash, expires_at, created_by, person_id, status, created_at, completed_at
		 FROM invitations
		 WHERE created_by = $1
		 ORDER BY created_at DESC`,
		creatorID.String(),
	)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to fetch invitations", err)
	}
	defer rows.Close()

	var list []*people.Invitation
	for rows.Next() {
		var (
			idStr       string
			email       string
			codeHash    string
			expiresAt   time.Time
			createdBy   string
			personID    types.NullString
			status      string
			createdAt   time.Time
			completedAt types.NullTime
		)

		if err := rows.Scan(
			&idStr,
			&email,
			&codeHash,
			&expiresAt,
			&createdBy,
			&personID,
			&status,
			&createdAt,
			&completedAt,
		); err != nil {
			return nil, exceptions.NewInternalException("failed to scan invitation", err)
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}
		cByID, err := uuid.Parse(createdBy)
		if err != nil {
			continue
		}

		var pID *uuid.UUID
		if personID.Valid && personID.String != "" {
			if parsedPID, err := uuid.Parse(personID.String); err == nil {
				pID = &parsedPID
			}
		}

		var compAt *time.Time
		if completedAt.Valid {
			compAt = &completedAt.Time
		}

		inv, err := people.NewInvitation(
			id,
			email,
			codeHash,
			expiresAt,
			cByID,
			pID,
			people.InvitationStatus(status),
			createdAt,
			compAt,
		)
		if err == nil {
			list = append(list, inv)
		}
	}

	return list, nil
}

func (r *PgInvitationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM invitations WHERE invitation_id = $1`, id.String())
	if err != nil {
		return exceptions.NewInternalException("failed to delete invitation", err)
	}
	return nil
}

func (r *PgInvitationRepository) Complete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE invitations SET status = $1, completed_at = now() WHERE invitation_id = $2`,
		string(people.InvitationStatusCompleted),
		id.String(),
	)
	if err != nil {
		return exceptions.NewInternalException("failed to complete invitation", err)
	}
	return nil
}

func (r *PgInvitationRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		ctx,
		`SELECT (
			EXISTS (SELECT 1 FROM credentials WHERE LOWER(registration_id) = LOWER($1) OR LOWER(identity) = LOWER($1))
			OR EXISTS (SELECT 1 FROM invitations WHERE LOWER(email) = LOWER($1) AND status IN ('PENDING', 'COMPLETED'))
			OR EXISTS (SELECT 1 FROM people WHERE LOWER(display_name) = LOWER($1) AND deleted_at IS NULL)
		)`,
		email,
	).Scan(&exists)
	if err != nil {
		return false, exceptions.NewInternalException("failed to check if email exists", err)
	}
	return exists, nil
}
