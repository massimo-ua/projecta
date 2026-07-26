package exceptions_test

import (
	"errors"
	"fmt"
	"testing"

	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type customUnwrappableErr struct {
	inner error
}

func (c customUnwrappableErr) Error() string {
	return fmt.Sprintf("custom: %v", c.inner)
}

func (c customUnwrappableErr) Unwrap() error {
	return c.inner
}

type otherCustomErr struct {
	msg string
}

func (o otherCustomErr) Error() string {
	return o.msg
}

type unmatchingErr struct{}

func (u unmatchingErr) Error() string {
	return "unmatching"
}

func TestException(t *testing.T) {
	baseErr := errors.New("base error")

	t.Run("Error returns message", func(t *testing.T) {
		ex := exceptions.NewApplicationError("custom message", exceptions.ValidationFailed, baseErr)
		if ex.Error() != "custom message" {
			t.Errorf("expected 'custom message', got '%s'", ex.Error())
		}
		if ex.Code != exceptions.ValidationFailed {
			t.Errorf("expected ValidationFailed code, got '%s'", ex.Code)
		}
	})

	t.Run("NewApplicationError defaults", func(t *testing.T) {
		ex := exceptions.NewApplicationError("", "", baseErr)
		if ex.Code != exceptions.Internal {
			t.Errorf("expected default code Internal, got '%s'", ex.Code)
		}
		if ex.Error() != baseErr.Error() {
			t.Errorf("expected default message '%s', got '%s'", baseErr.Error(), ex.Error())
		}
	})

	t.Run("Unwrap plain error", func(t *testing.T) {
		ex := exceptions.NewApplicationError("msg", exceptions.Internal, baseErr)
		if !errors.Is(ex.Unwrap(), baseErr) {
			t.Errorf("expected unwrapped error to be baseErr")
		}
	})

	t.Run("Unwrap unwrappable error", func(t *testing.T) {
		wrapped := customUnwrappableErr{inner: baseErr}
		ex := exceptions.NewApplicationError("msg", exceptions.Internal, wrapped)
		if !errors.Is(ex.Unwrap(), baseErr) {
			t.Errorf("expected unwrapped error to unwrap inner baseErr")
		}
	})

	t.Run("Is method", func(t *testing.T) {
		ex := exceptions.NewApplicationError("msg", exceptions.Internal, baseErr)

		if !errors.Is(ex, baseErr) {
			t.Errorf("ex should match baseErr with errors.Is")
		}

		targetEx := exceptions.NewApplicationError("other", exceptions.NotFound, nil)
		if errors.Is(ex, targetEx) {
			t.Errorf("ex should not match another Exception with errors.Is")
		}
	})

	t.Run("As method", func(t *testing.T) {
		ex := exceptions.NewApplicationError("msg", exceptions.Internal, otherCustomErr{msg: "inner"})

		var targetEx exceptions.Exception
		if !ex.As(&targetEx) {
			t.Fatalf("ex.As(&Exception) should succeed")
		}
		if targetEx.Message != "msg" {
			t.Errorf("targetEx message mismatch: got '%s'", targetEx.Message)
		}

		var targetCustom otherCustomErr
		if !ex.As(&targetCustom) {
			t.Fatalf("ex.As(&otherCustomErr) should succeed")
		}
		if targetCustom.msg != "inner" {
			t.Errorf("targetCustom msg mismatch: got '%s'", targetCustom.msg)
		}

		var targetUnmatching unmatchingErr
		if ex.As(&targetUnmatching) {
			t.Errorf("ex.As(&unmatchingErr) should return false")
		}
	})

	t.Run("Helper constructors", func(t *testing.T) {
		nf := exceptions.NewNotFoundException("not found", baseErr)
		if nf.Code != exceptions.NotFound {
			t.Errorf("expected NotFound code, got '%s'", nf.Code)
		}
		if !errors.Is(nf, exceptions.NotFoundError) {
			t.Errorf("NewNotFoundException should contain NotFoundError")
		}

		internal := exceptions.NewInternalException("internal err", baseErr)
		if internal.Code != exceptions.Internal {
			t.Errorf("expected Internal code, got '%s'", internal.Code)
		}

		val := exceptions.NewValidationException("validation failed", baseErr)
		if val.Code != exceptions.ValidationFailed {
			t.Errorf("expected ValidationFailed code, got '%s'", val.Code)
		}

		unauth := exceptions.NewUnauthorizedException("unauthorized", baseErr)
		if unauth.Code != exceptions.Unauthorized {
			t.Errorf("expected Unauthorized code, got '%s'", unauth.Code)
		}
	})
}
