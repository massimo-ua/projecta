package dal

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"sync"
)

// txKey is a context key for storing transaction reference
type txKey struct{}

type PgDbConnection struct {
	pool *pgxpool.Pool
	// Remove tx field as we'll store it in context
	mu sync.RWMutex
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

// GetConnection returns either active transaction or pool
func (p *PgDbConnection) GetConnection(ctx context.Context) (PgDb, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Check if there's a transaction in the context
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if ok && tx != nil {
		return tx, nil
	}
	return p.pool, nil
}

// Tx starts a transaction and executes the provided function
func (p *PgDbConnection) Tx(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	p.mu.Lock()
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		p.mu.Unlock()
		return nil, exceptions.NewInternalException(fmt.Sprintf("beginning transaction: %s", err.Error()), errors.Join(core.DbTransactionStartFailedError, err))
	}

	// Create new context with transaction
	txCtx := context.WithValue(ctx, txKey{}, tx)
	p.mu.Unlock()

	// Execute the function with transaction context
	res, err := fn(txCtx)
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return nil, exceptions.NewInternalException(
				fmt.Sprintf(
					"transaction rollback failed err: %s, rb err: %s",
					err.Error(),
					rollbackErr.Error(),
				), errors.Join(core.DbTransactionRollbackFailedError, err, rollbackErr),
			)
		}

		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, exceptions.NewInternalException(
			fmt.Sprintf("committing transaction: %s", err.Error()),
			errors.Join(core.DbTransactionCommitFailedError, err),
		)
	}

	return res, nil
}

// Close closes the connection pool
func (p *PgDbConnection) Close() {
	p.pool.Close()
}
