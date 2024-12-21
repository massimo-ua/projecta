package dal

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

// errorRow is a mock for pgx.Row
type errorRow struct {
	err error
}

// Scan is a mock for pgx.Row.Scan to prevent nil pointer dereference
func (e *errorRow) Scan(_ ...any) error {
	return e.err
}

// PgDb is an interface for working with Postgres database
type PgDb interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type PgRepository struct {
	db *PgDbConnection
}

func (r *PgRepository) getConnection(ctx context.Context) (PgDb, error) {
	db, err := r.db.GetConnection(ctx)

	if err != nil {
		return nil, exceptions.NewInternalException(err.Error(), errors.Join(core.DbFailedToGetConnectionError, err))
	}

	return db, nil
}

func (r *PgRepository) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	conn, err := r.getConnection(ctx)

	if err != nil {
		return pgconn.CommandTag{}, err
	}

	return conn.Exec(ctx, sql, arguments...)
}

func (r *PgRepository) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	conn, err := r.getConnection(ctx)

	if err != nil {
		return nil, err
	}

	return conn.Query(ctx, sql, args...)
}

func (r *PgRepository) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	conn, err := r.getConnection(ctx)

	if err != nil {
		return &errorRow{err: err}
	}

	return conn.QueryRow(ctx, sql, args...)
}
