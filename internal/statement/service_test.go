package statement_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/statement"
)

type mockProjectValidator struct {
	err error
}

func (m *mockProjectValidator) ValidateProject(_ context.Context, _ core.ProjectID) error {
	return m.err
}

type mockFinder struct {
	txs []statement.ExistingTransaction
	err error
}

func (m *mockFinder) FindExisting(_ context.Context, _ core.ProjectID) ([]statement.ExistingTransaction, error) {
	return m.txs, m.err
}

type mockParser struct {
	stmt *statement.ParsedStatement
	err  error
}

func (m *mockParser) Parse(_ []byte) (*statement.ParsedStatement, error) {
	return m.stmt, m.err
}

func TestStatementService(t *testing.T) {
	ctx := context.Background()
	projID := core.NewProjectID()
	now := time.Now()

	t.Run("Validation: missing project id", func(t *testing.T) {
		svc := statement.NewService(&mockProjectValidator{}, nil)
		_, err := svc.ParseStatement(ctx, core.NilProjectID, []byte("data"))
		if err == nil {
			t.Errorf("expected validation error for nil project id")
		}
	})

	t.Run("Project not found error", func(t *testing.T) {
		svc := statement.NewService(&mockProjectValidator{err: errors.New("not found")}, nil)
		_, err := svc.ParseStatement(ctx, projID, []byte("data"))
		if err == nil {
			t.Errorf("expected error when project is not found")
		}
	})

	t.Run("Parser error", func(t *testing.T) {
		p := &mockParser{err: errors.New("corrupt file")}
		svc := statement.NewService(&mockProjectValidator{}, nil, p)
		_, err := svc.ParseStatement(ctx, projID, []byte("corrupt"))
		if err == nil {
			t.Errorf("expected parser error")
		}
	})

	t.Run("Success with duplicate detection and polymorphic parser", func(t *testing.T) {
		p := &mockParser{
			stmt: &statement.ParsedStatement{
				Account:  "ACC123",
				Currency: "UAH",
				Period:   "2026-09",
				Transactions: []statement.ParsedStatementTransaction{
					{
						Date:        now,
						DocNum:      "DOC1",
						Description: "Office Supplies",
						AmountMinor: 50000,
						Currency:    "UAH",
					},
					{
						Date:        now,
						DocNum:      "DOC2",
						Description: "Coffee Beans",
						AmountMinor: 12000,
						Currency:    "UAH",
					},
				},
			},
		}

		finder := &mockFinder{
			txs: []statement.ExistingTransaction{
				{
					Date:        now,
					Amount:      50000,
					Description: "Office Supplies",
				},
			},
		}

		svc := statement.NewService(&mockProjectValidator{}, finder, p)
		res, err := svc.ParseStatement(ctx, projID, []byte("data"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.Account != "ACC123" || res.Currency != "UAH" {
			t.Errorf("unexpected statement metadata: %+v", res)
		}
		if len(res.Transactions) != 2 {
			t.Fatalf("expected 2 transactions, got %d", len(res.Transactions))
		}

		// DOC1 must be marked duplicate
		if !res.Transactions[0].IsDuplicate {
			t.Errorf("expected DOC1 to be marked as duplicate")
		}
		// DOC2 must not be duplicate
		if res.Transactions[1].IsDuplicate {
			t.Errorf("expected DOC2 not to be marked as duplicate")
		}
	})

	t.Run("Finder error", func(t *testing.T) {
		p := &mockParser{
			stmt: &statement.ParsedStatement{
				Account:  "ACC123",
				Currency: "UAH",
				Transactions: []statement.ParsedStatementTransaction{
					{Date: now, DocNum: "DOC1", AmountMinor: 50000, Currency: "UAH"},
				},
			},
		}
		finder := &mockFinder{err: errors.New("db error")}
		svc := statement.NewService(&mockProjectValidator{}, finder, p)
		_, err := svc.ParseStatement(ctx, projID, []byte("data"))
		if err == nil {
			t.Errorf("expected error when finder fails")
		}
	})
}
