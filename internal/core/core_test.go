package core_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

func TestCollection(t *testing.T) {
	col := core.NewPaginatedCollection[string](10)
	if col.Total() != 10 {
		t.Errorf("expected total 10, got %d", col.Total())
	}
	if len(col.Elements()) != 0 {
		t.Errorf("expected empty elements")
	}

	col.Add("item1", "item2")
	if len(col.Elements()) != 2 {
		t.Errorf("expected 2 elements, got %d", len(col.Elements()))
	}
	if col.Elements()[0] != "item1" || col.Elements()[1] != "item2" {
		t.Errorf("unexpected elements content")
	}
}

func TestDates(t *testing.T) {
	t.Run("zero time returns non-zero time", func(t *testing.T) {
		var zero time.Time
		got := core.DateOrNow(zero)
		if got.IsZero() {
			t.Errorf("expected non-zero time")
		}
	})

	t.Run("non-zero time returns original time", func(t *testing.T) {
		now := time.Now()
		got := core.DateOrNow(now)
		if !got.Equal(now) {
			t.Errorf("expected %v, got %v", now, got)
		}
	})
}

func TestContext(t *testing.T) {
	t.Run("AuthGuard with valid UUID context value", func(t *testing.T) {
		expectedID := uuid.New()
		ctx := context.WithValue(context.Background(), core.RequesterIDContextKey, expectedID)

		gotID, err := core.AuthGuard(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotID != expectedID {
			t.Errorf("expected %v, got %v", expectedID, gotID)
		}
	})

	t.Run("AuthGuard with missing or wrong type context value", func(t *testing.T) {
		ctx := context.Background()
		_, err := core.AuthGuard(ctx)
		if !errors.Is(err, core.FailedToIdentifyRequester) {
			t.Errorf("expected FailedToIdentifyRequester, got %v", err)
		}

		ctxWrong := context.WithValue(context.Background(), core.RequesterIDContextKey, "not-a-uuid")
		_, err = core.AuthGuard(ctxWrong)
		if !errors.Is(err, core.FailedToIdentifyRequester) {
			t.Errorf("expected FailedToIdentifyRequester, got %v", err)
		}
	})
}

func TestFilter(t *testing.T) {
	if core.ASC.String() != "ASC" {
		t.Errorf("expected ASC.String() to be ASC")
	}
	if core.DESC.String() != "DESC" {
		t.Errorf("expected DESC.String() to be DESC")
	}

	if core.ToOrder("ASC") != core.ASC {
		t.Errorf("expected ASC")
	}
	if core.ToOrder("DESC") != core.DESC {
		t.Errorf("expected DESC")
	}
	if core.ToOrder("UNKNOWN") != core.ASC {
		t.Errorf("expected default ASC for unknown string")
	}
}

func TestTokenRing(t *testing.T) {
	t.Run("NewTokenRing success", func(t *testing.T) {
		ring, err := core.NewTokenRing("access_123", "refresh_456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ring.AccessToken() != "access_123" {
			t.Errorf("unexpected access token: %s", ring.AccessToken())
		}
		if ring.RefreshToken() != "refresh_456" {
			t.Errorf("unexpected refresh token: %s", ring.RefreshToken())
		}
	})

	t.Run("NewTokenRing invalid parameters", func(t *testing.T) {
		_, err := core.NewTokenRing("", "refresh")
		if !errors.Is(err, core.RefreshTokenIsInvalid) {
			t.Errorf("expected RefreshTokenIsInvalid error")
		}

		_, err = core.NewTokenRing("access", "")
		if !errors.Is(err, core.RefreshTokenIsInvalid) {
			t.Errorf("expected RefreshTokenIsInvalid error")
		}
	})
}
