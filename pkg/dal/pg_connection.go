package dal

import (
	"context"
	"errors"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

// txKey is a context key for storing transaction reference
type txKey struct{}

type PgDbConnection struct {
	pool *pgxpool.Pool
	mu   sync.RWMutex
}

func NewPgDbConnection(connectionString string) (*PgDbConnection, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, exceptions.NewInternalException("parsing pg config failed", errors.Join(core.DbConfigurationError, err))
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, exceptions.NewInternalException("creating pg pool failed", errors.Join(core.DbConnectionFailedError, err))
	}

	return &PgDbConnection{
		pool: pool,
	}, nil
}

// GetConnection returns either active transaction, mock, or pool
func (p *PgDbConnection) GetConnection(ctx context.Context) (PgDb, error) {
	if p == nil {
		return nil, errors.New("connection is nil")
	}
	p.mu.RLock()
	defer p.mu.RUnlock()

	if db, ok := ctx.Value(txKey{}).(PgDb); ok && db != nil {
		return db, nil
	}
	if p.pool == nil {
		return nil, errors.New("connection pool is nil")
	}
	return p.pool, nil
}

// finishTransaction commits tx when err is nil, or rolls back tx when err is non-nil.
// Rollback errors are joined with the original error so neither is lost.
func finishTransaction(ctx context.Context, tx pgx.Tx, err error) error {
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return exceptions.NewInternalException(
				"transaction rollback failed",
				errors.Join(core.DbTransactionRollbackFailedError, err, rollbackErr),
			)
		}
		return err
	}
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return exceptions.NewInternalException(
			"failed to commit transaction",
			errors.Join(core.DbTransactionCommitFailedError, commitErr),
		)
	}
	return nil
}

// Tx starts a transaction and executes the provided function with defer-based commit/rollback safety.
func (p *PgDbConnection) Tx(ctx context.Context, fn func(ctx context.Context) (any, error)) (res any, err error) {
	if p == nil {
		return nil, errors.New("connection pool is nil")
	}

	// Reuse active transaction or mock db present in context.
	if db, ok := ctx.Value(txKey{}).(PgDb); ok && db != nil {
		return fn(ctx)
	}

	if p.pool == nil {
		return nil, errors.New("connection pool is nil")
	}

	p.mu.Lock()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		p.mu.Unlock()
		return nil, exceptions.NewInternalException("failed to begin transaction", errors.Join(core.DbTransactionStartFailedError, err))
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)
	p.mu.Unlock()

	committed := false
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback(ctx)
			panic(r)
		}
		if !committed {
			err = finishTransaction(ctx, tx, err)
			if err != nil {
				res = nil
			}
		}
	}()

	res, err = fn(txCtx)
	err = finishTransaction(ctx, tx, err)
	committed = true
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Close closes the connection pool
func (p *PgDbConnection) Close() {
	if p != nil && p.pool != nil {
		p.pool.Close()
	}
}

// Ping checks if the connection is alive
func (p *PgDbConnection) Ping(ctx context.Context) error {
	if p == nil || p.pool == nil {
		return errors.New("connection pool is nil")
	}
	return p.pool.Ping(ctx)
}
