package dal

import (
    "context"
    db "database/sql"
    "errors"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "gitlab.com/massimo-ua/projecta/internal/core"
    "time"
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

func toNullTime(t time.Time) db.NullTime {
    if !t.IsZero() {
        return db.NullTime{
            Time:  t,
            Valid: true,
        }
    }

    return db.NullTime{}
}

func toTime(t db.NullTime) time.Time {
    if t.Valid {
        return t.Time
    }

    return time.Time{}
}
