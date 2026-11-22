package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
	projectapp "gitlab.com/massimo-ua/projecta/internal/project/app"
)

type ownerReaderStub struct {
	owner *project.Owner
	err   error
}

func (s ownerReaderStub) FindOwner(_ context.Context, _ core.PersonID) (*project.Owner, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.owner, nil
}

type totalsReadModelStub struct {
	totals project.ProjectTotals
	err    error
}

func (s totalsReadModelStub) CalculateTotals(_ context.Context, _ core.ProjectID, _ core.PersonID, homeCurrency string) (project.ProjectTotals, error) {
	if s.err != nil {
		return project.ProjectTotals{}, s.err
	}
	res := s.totals
	if res.Currency == "" {
		res.Currency = homeCurrency
	}
	return res, nil
}

type converterStub struct {
	rate float64
}

func (c converterStub) ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error) {
	if amount == nil {
		return money.New(0, targetCurrency), nil
	}
	if amount.Currency().Code == targetCurrency {
		return amount, nil
	}
	return money.New(int64(float64(amount.Amount())*c.rate), targetCurrency), nil
}

func TestNewApplication_FailFastOnMissingDependencies(t *testing.T) {
	repo := project.NewMemoryProjectRepository()
	ownerStub := ownerReaderStub{}

	t.Run("panics on nil repo", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on nil repository")
			}
		}()
		projectapp.NewApplication(nil, ownerStub, nil, nil, nil)
	})

	t.Run("panics on nil ownerReader", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on nil ownerReader")
			}
		}()
		projectapp.NewApplication(repo, nil, nil, nil, nil)
	})
}

func TestProjectApplication_CRUD_Auth_And_Idempotency(t *testing.T) {
	ctx := context.Background()
	ownerID := core.NewPersonID()
	otherID := core.NewPersonID()
	sharedID := core.NewPersonID()

	repo := project.NewMemoryProjectRepository()
	ownerStub := ownerReaderStub{
		owner: &project.Owner{
			PersonID:        ownerID,
			DisplayName:     "Alice Owner",
			CanHaveProjects: true,
		},
	}
	totalsStub := totalsReadModelStub{
		totals: project.ProjectTotals{
			TotalInvested: 10000,
			Balance:       4000,
			HasActivity:   true,
		},
	}

	app := projectapp.NewApplication(repo, ownerStub, totalsStub, converterStub{rate: 40}, nil)

	// 1. Create with pre-generated ProjectID
	projID := core.NewProjectID()
	err := app.Commands.CreateProject.Handle(ctx, projectapp.CreateProject{
		ProjectID:    projID,
		RequesterID:  ownerID,
		Name:         "Solar Home",
		Description:  "Off-grid solar build",
		MainCurrency: "UAH",
		StartDate:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		TagBudgets: []project.TagBudgetInput{
			{Tag: "panels", Frequency: "MONTHLY", ResourceType: "MONEY", Amount: 10000, Currency: "USD"},
		},
	})
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// 2. Idempotent Create with duplicate name returns nil and keeps original project
	dupID := core.NewProjectID()
	if err := app.Commands.CreateProject.Handle(ctx, projectapp.CreateProject{
		ProjectID:   dupID,
		RequesterID: ownerID,
		Name:        "Solar Home",
	}); err != nil {
		t.Fatalf("duplicate CreateProject failed: %v", err)
	}

	got, err := app.Queries.GetProject.Handle(ctx, projectapp.GetProject{
		ProjectID:   projID,
		RequesterID: ownerID,
	})
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if got.Name() != "Solar Home" || got.TagBudgets()[0].HomeAmount() != 400000 {
		t.Fatalf("unexpected project or evaluated home amount: name=%q home=%d", got.Name(), got.TagBudgets()[0].HomeAmount())
	}

	// 3. Unauthorized outsider cannot view or update project
	if _, err := app.Queries.GetProject.Handle(ctx, projectapp.GetProject{
		ProjectID:   projID,
		RequesterID: otherID,
	}); err == nil {
		t.Fatal("expected forbidden/not found error for outsider GetProject")
	}

	// 4. Share acceptance grants read access, but NOT owner management access
	if err := app.Commands.AcceptProjectShare.Handle(ctx, projectapp.AcceptProjectShare{
		ShareToken:  got.ShareToken(),
		RequesterID: sharedID,
	}); err != nil {
		t.Fatalf("AcceptProjectShare failed: %v", err)
	}

	sharedView, err := app.Queries.GetProject.Handle(ctx, projectapp.GetProject{
		ProjectID:   projID,
		RequesterID: sharedID,
	})
	if err != nil || !sharedView.IsShared() {
		t.Fatalf("expected shared participant to view project with isShared=true, err=%v", err)
	}

	err = app.Commands.UpdateProjectBudget.Handle(ctx, projectapp.UpdateProjectBudget{
		ProjectID:   projID,
		RequesterID: sharedID,
		TagBudgets:  []project.TagBudgetInput{},
	})
	var exc exceptions.Exception
	if !errors.As(err, &exc) || exc.Code != exceptions.Forbidden {
		t.Fatalf("expected Forbidden for non-owner UpdateProjectBudget, got %v", err)
	}

	// 5. Owner updates details & budget atomically via UpdateProject closure
	if err := app.Commands.UpdateProject.Handle(ctx, projectapp.UpdateProject{
		ProjectID:   projID,
		RequesterID: ownerID,
		Name:        "Solar Home v2",
		Description: "Updated desc",
	}); err != nil {
		t.Fatalf("UpdateProject failed: %v", err)
	}

	updated, err := app.Queries.GetProject.Handle(ctx, projectapp.GetProject{
		ProjectID:   projID,
		RequesterID: ownerID,
	})
	if err != nil || updated.Name() != "Solar Home v2" {
		t.Fatalf("expected updated name, got %v err=%v", updated, err)
	}

	// 6. GetProjectTotals
	totals, err := app.Queries.GetProjectTotals.Handle(ctx, projectapp.GetProjectTotals{
		ProjectID:   projID,
		RequesterID: ownerID,
	})
	if err != nil || totals.TotalInvested != 10000 || totals.Currency != "UAH" {
		t.Fatalf("unexpected totals: %+v err=%v", totals, err)
	}

	// 7. Invalid update rolls back cleanly in value-copy MemoryProjectRepository
	if err := app.Commands.UpdateProject.Handle(ctx, projectapp.UpdateProject{
		ProjectID:   projID,
		RequesterID: ownerID,
		Name:        "X", // too short
	}); err == nil {
		t.Fatal("expected validation error for short name")
	}
	afterFailed, _ := app.Queries.GetProject.Handle(ctx, projectapp.GetProject{
		ProjectID:   projID,
		RequesterID: ownerID,
	})
	if afterFailed.Name() != "Solar Home v2" {
		t.Fatalf("failed update leaked state into repository: got %q", afterFailed.Name())
	}

	// 8. ListProjects & RemoveProject
	list, err := app.Queries.ListProjects.Handle(ctx, projectapp.ListProjects{
		Pagination:  core.Pagination{Limit: 10},
		RequesterID: ownerID,
	})
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 project in list, got %d err=%v", len(list), err)
	}

	if err := app.Commands.RemoveProject.Handle(ctx, projectapp.RemoveProject{
		ProjectID:   projID,
		RequesterID: sharedID,
	}); err == nil {
		t.Fatal("expected forbidden when non-owner removes project")
	}

	if err := app.Commands.RemoveProject.Handle(ctx, projectapp.RemoveProject{
		ProjectID:   projID,
		RequesterID: ownerID,
	}); err != nil {
		t.Fatalf("RemoveProject failed: %v", err)
	}

	if err := app.Commands.AcceptProjectShare.Handle(ctx, projectapp.AcceptProjectShare{
		ShareToken:  uuid.Nil,
		RequesterID: otherID,
	}); err == nil {
		t.Fatal("expected validation error for nil share token")
	}
}
