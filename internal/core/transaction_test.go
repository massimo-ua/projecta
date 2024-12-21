package core

import (
	"context"
	"errors"
	"testing"
)

// MockUnitOfWork implements UnitOfWork for testing
type MockUnitOfWork struct {
	ExecuteFunc func(context.Context, func(TxCtx) (any, error)) (any, error)
}

func (m *MockUnitOfWork) Execute(ctx context.Context, work func(TxCtx) (any, error)) (any, error) {
	return m.ExecuteFunc(ctx, work)
}

// MockTxCtx implements TxCtx for testing
type MockTxCtx struct {
	ctx context.Context
}

func (m *MockTxCtx) Context() context.Context {
	return m.ctx
}

func TestTypedUnitOfWork_Execute(t *testing.T) {
	t.Run("successful execution with string return type", func(t *testing.T) {
		// Arrange
		expectedResult := "success"
		mock := &MockUnitOfWork{
			ExecuteFunc: func(_ context.Context, work func(TxCtx) (any, error)) (any, error) {
				return work(&MockTxCtx{ctx: context.Background()})
			},
		}

		uow := NewTypedUnitOfWork[string](mock)

		// Act
		result, err := uow.Execute(context.Background(), func(txCtx TxCtx) (string, error) {
			return expectedResult, nil
		})

		// Assert
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != expectedResult {
			t.Errorf("expected result %v, got %v", expectedResult, result)
		}
	})

	t.Run("successful execution with struct return type", func(t *testing.T) {
		// Arrange
		type User struct {
			ID   int
			Name string
		}
		expectedResult := User{ID: 1, Name: "test"}

		mock := &MockUnitOfWork{
			ExecuteFunc: func(_ context.Context, work func(TxCtx) (any, error)) (any, error) {
				return work(&MockTxCtx{ctx: context.Background()})
			},
		}

		uow := NewTypedUnitOfWork[User](mock)

		// Act
		result, err := uow.Execute(context.Background(), func(txCtx TxCtx) (User, error) {
			return expectedResult, nil
		})

		// Assert
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != expectedResult {
			t.Errorf("expected result %v, got %v", expectedResult, result)
		}
	})

	t.Run("successful execution with pointer return type", func(t *testing.T) {
		// Arrange
		type User struct {
			ID   int
			Name string
		}
		expectedResult := &User{ID: 1, Name: "test"}

		mock := &MockUnitOfWork{
			ExecuteFunc: func(_ context.Context, work func(TxCtx) (any, error)) (any, error) {
				return work(&MockTxCtx{ctx: context.Background()})
			},
		}

		uow := NewTypedUnitOfWork[*User](mock)

		// Act
		result, err := uow.Execute(context.Background(), func(txCtx TxCtx) (*User, error) {
			return expectedResult, nil
		})

		// Assert
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != expectedResult {
			t.Errorf("expected result %v, got %v", expectedResult, result)
		}
	})

	t.Run("handles error from work function", func(t *testing.T) {
		// Arrange
		expectedError := errors.New("work function error")
		mock := &MockUnitOfWork{
			ExecuteFunc: func(_ context.Context, work func(TxCtx) (any, error)) (any, error) {
				return work(&MockTxCtx{ctx: context.Background()})
			},
		}

		uow := NewTypedUnitOfWork[string](mock)

		// Act
		result, err := uow.Execute(context.Background(), func(txCtx TxCtx) (string, error) {
			return "", expectedError
		})

		// Assert
		if !errors.Is(err, expectedError) {
			t.Errorf("expected error %v, got %v", expectedError, err)
		}
		if result != "" {
			t.Errorf("expected empty result, got %v", result)
		}
	})

	t.Run("handles error from UnitOfWork.Execute", func(t *testing.T) {
		// Arrange
		expectedError := errors.New("unit of work error")
		mock := &MockUnitOfWork{
			ExecuteFunc: func(_ context.Context, _ func(TxCtx) (any, error)) (any, error) {
				return nil, expectedError
			},
		}

		uow := NewTypedUnitOfWork[string](mock)

		// Act
		result, err := uow.Execute(context.Background(), func(txCtx TxCtx) (string, error) {
			return "success", nil
		})

		// Assert
		if !errors.Is(err, expectedError) {
			t.Errorf("expected error %v, got %v", expectedError, err)
		}
		if result != "" {
			t.Errorf("expected empty result, got %v", result)
		}
	})

	t.Run("passes context correctly", func(t *testing.T) {
		// Arrange
		ctx := context.WithValue(context.Background(), "key", "value")
		mock := &MockUnitOfWork{
			ExecuteFunc: func(execCtx context.Context, work func(TxCtx) (any, error)) (any, error) {
				return work(&MockTxCtx{ctx: execCtx})
			},
		}

		uow := NewTypedUnitOfWork[string](mock)

		// Act
		_, err := uow.Execute(ctx, func(txCtx TxCtx) (string, error) {
			// Assert
			if txCtx.Context().Value("key") != "value" {
				t.Error("context value not passed correctly")
			}
			return "success", nil
		})

		// Assert
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("handles primitive number types", func(t *testing.T) {
		// Arrange
		expectedResult := int64(42)
		mock := &MockUnitOfWork{
			ExecuteFunc: func(_ context.Context, work func(TxCtx) (any, error)) (any, error) {
				return work(&MockTxCtx{ctx: context.Background()})
			},
		}

		uow := NewTypedUnitOfWork[int64](mock)

		// Act
		result, err := uow.Execute(context.Background(), func(txCtx TxCtx) (int64, error) {
			return expectedResult, nil
		})

		// Assert
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != expectedResult {
			t.Errorf("expected result %v, got %v", expectedResult, result)
		}
	})

	t.Run("handles slice types", func(t *testing.T) {
		// Arrange
		expectedResult := []string{"a", "b", "c"}
		mock := &MockUnitOfWork{
			ExecuteFunc: func(_ context.Context, work func(TxCtx) (any, error)) (any, error) {
				return work(&MockTxCtx{ctx: context.Background()})
			},
		}

		uow := NewTypedUnitOfWork[[]string](mock)

		// Act
		result, err := uow.Execute(context.Background(), func(txCtx TxCtx) ([]string, error) {
			return expectedResult, nil
		})

		// Assert
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(result) != len(expectedResult) {
			t.Errorf("expected result length %v, got %v", len(expectedResult), len(result))
		}
		for i, v := range result {
			if v != expectedResult[i] {
				t.Errorf("expected result[%d] = %v, got %v", i, expectedResult[i], v)
			}
		}
	})
}
