package core

import (
	"context"
)

var TxCtxKey = "txCtx"

type BaseRepository interface {
	TxCtx(ctx context.Context) (context.Context, error)
	RollbackTxFromCtx(ctx context.Context)
	CommitTxFromCtx(ctx context.Context) error
}
