package dal

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

var failedToObtainTransactionError = errors.New("failed to obtain transaction")

type Operation struct {
	Query string
	Args  []any
}

type PgRepository struct {
	db *pgxpool.Pool
}

func (r *PgRepository) TxCtx(ctx context.Context) (context.Context, error) {
	tx, err := r.db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, core.TxCtxKey, tx), nil
}

func Connect(uri string) (*pgxpool.Pool, error) {
	p, err := pgxpool.New(context.Background(), uri)

	if err != nil {
		return nil, err
	}

	return p, nil
}

func rollbackTx(ctx context.Context, tx pgx.Tx) {
	_ = tx.Rollback(ctx)
}

func (r *PgRepository) RollbackTxFromCtx(ctx context.Context) {
	tx, ok := ctx.Value(core.TxCtxKey).(pgx.Tx)

	if ok {
		rollbackTx(ctx, tx)
	}
}

func (r *PgRepository) CommitTxFromCtx(ctx context.Context) error {
	tx, ok := ctx.Value(core.TxCtxKey).(pgx.Tx)

	if !ok {
		return failedToObtainTransactionError
	}

	return tx.Commit(ctx)
}
