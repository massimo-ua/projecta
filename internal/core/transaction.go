package core

import "context"

// TxCtx represents an abstract context for transactional operations
type TxCtx interface {
	Context() context.Context
}

// UnitOfWork defines a generic unit of work interface
type UnitOfWork interface {
	// Execute runs the work function in a transaction and returns the result
	Execute(ctx context.Context, work func(tc TxCtx) (any, error)) (any, error)
}

// TypedUnitOfWork provides a type-safe wrapper around UnitOfWork
type TypedUnitOfWork[T any] struct {
	uow UnitOfWork
}

// NewTypedUnitOfWork creates a new typed unit of work
func NewTypedUnitOfWork[T any](uow UnitOfWork) *TypedUnitOfWork[T] {
	return &TypedUnitOfWork[T]{uow: uow}
}

// Execute runs the work function with type safety
func (t *TypedUnitOfWork[T]) Execute(ctx context.Context, work func(txCtx TxCtx) (T, error)) (T, error) {
	result, err := t.uow.Execute(ctx, func(tc TxCtx) (interface{}, error) {
		return work(tc)
	})

	if err != nil {
		var zero T
		return zero, err
	}

	return result.(T), nil
}
