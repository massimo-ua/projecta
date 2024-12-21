package core

import (
	"context"
	"errors"
)

var (
	DbConfigurationError             = errors.New("configuration error")
	DbConnectionFailedError          = errors.New("connection failed error")
	DbTransactionStartFailedError    = errors.New("transaction start failed error")
	DbTransactionCommitFailedError   = errors.New("transaction commit failed error")
	DbTransactionRollbackFailedError = errors.New("transaction rollback failed error")
	DbFailedToGetConnectionError     = errors.New("failed to get connection error")
)

type DbConnection interface {
	Tx(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error)
	Close()
	Ping(ctx context.Context) error
}
