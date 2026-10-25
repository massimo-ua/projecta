package project_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

func TestDomainEntities(t *testing.T) {
	ownerID := core.NewPersonID()
	owner := &project.Owner{PersonID: ownerID, DisplayName: "John Doe"}
	now := time.Now()

	// Project
	projID := core.NewProjectID()
	proj, err := project.NewProject(projID, "Test Project", "Description", owner, now, now)
	if err != nil {
		t.Fatalf("unexpected project error: %v", err)
	}
	if !proj.IsOwnedBy(owner) {
		t.Errorf("IsOwnedBy mismatch")
	}
	otherOwner := &project.Owner{PersonID: core.NewPersonID()}
	if proj.IsOwnedBy(otherOwner) {
		t.Errorf("IsOwnedBy should be false for other owner")
	}

	// Invalid project name
	_, err = project.NewProject(projID, "A", "Desc", owner, now, now)
	if err == nil {
		t.Errorf("expected error for short project name")
	}
}

// Mock Repositories for projecta services
type mockPeopleService struct {
	owner *project.Owner
	err   error
}

func (m *mockPeopleService) FindOwner(ctx context.Context, personID core.PersonID) (*project.Owner, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.owner, nil
}

type mockProjectRepo struct {
	project   *project.Project
	projects  []*project.Project
	findErr   error
	createErr error
	updateErr error
	removeErr error
	shareErr  error
}

func (m *mockProjectRepo) Find(ctx context.Context, filter project.ProjectCollectionFilter) ([]*project.Project, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.projects != nil {
		return m.projects, nil
	}
	if m.project != nil {
		return []*project.Project{m.project}, nil
	}
	return []*project.Project{}, nil
}

func (m *mockProjectRepo) FindOne(ctx context.Context, filter project.ProjectFilter) (*project.Project, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.project != nil {
		return m.project, nil
	}
	return nil, exceptions.NotFoundError
}

func (m *mockProjectRepo) FindByShareToken(ctx context.Context, token uuid.UUID) (*project.Project, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.project != nil {
		return m.project, nil
	}
	return nil, exceptions.NotFoundError
}

func (m *mockProjectRepo) Create(ctx context.Context, project *project.Project) error {
	return m.createErr
}

func (m *mockProjectRepo) Update(ctx context.Context, project *project.Project) error {
	return m.updateErr
}

func (m *mockProjectRepo) Remove(ctx context.Context, project *project.Project) error {
	return m.removeErr
}

func (m *mockProjectRepo) CreateShareRecord(ctx context.Context, projectID core.ProjectID, personID core.PersonID) error {
	return m.shareErr
}

func TestProjectService(t *testing.T) {
	ownerID := core.NewPersonID()
	owner := &project.Owner{PersonID: ownerID, DisplayName: "John", CanHaveProjects: true}
	peopleSvc := &mockPeopleService{owner: owner}

	proj, _ := project.NewProject(core.NewProjectID(), "Project 1", "Desc", owner, time.Now(), time.Now())
	projRepo := &mockProjectRepo{project: proj}

	svc := project.NewProjectService(projRepo, peopleSvc)

	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, ownerID)

	// FindOne unauthenticated error
	if _, err := svc.FindOne(context.Background(), project.ProjectFilter{}); err == nil {
		t.Fatalf("expected unauthenticated error on FindOne")
	}

	// Find unauthenticated error
	if _, err := svc.Find(context.Background(), project.ProjectCollectionFilter{}); err == nil {
		t.Fatalf("expected unauthenticated error on Find")
	}

	// FindOne
	p, err := svc.FindOne(authedCtx, project.ProjectFilter{})
	if err != nil || p == nil {
		t.Fatalf("FindOne error: %v", err)
	}

	// Find
	list, err := svc.Find(authedCtx, project.ProjectCollectionFilter{})
	if err != nil || len(list) != 1 {
		t.Fatalf("Find error: %v", err)
	}

	// Create existing project (returns existing)
	pExisting, err := svc.Create(context.Background(), project.CreateProjectCommand{PersonID: owner.PersonID, Name: "Project 1"})
	if err != nil || pExisting == nil {
		t.Fatalf("Create existing project error: %v", err)
	}

	// Create new project with tag budgets
	projRepoNotFound := &mockProjectRepo{findErr: exceptions.NotFoundError}
	svcNew := project.NewProjectService(projRepoNotFound, peopleSvc)
	pNew, err := svcNew.Create(context.Background(), project.CreateProjectCommand{
		PersonID:    owner.PersonID,
		Name:        "New Project",
		Description: "Desc",
		StartDate:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		TagBudgets: []project.TagBudgetInput{
			{Tag: "materials", Frequency: "MONTHLY", Amount: 50000},
			{Tag: "labor", Frequency: "MONTHLY", Amount: 100000},
		},
	})
	if err != nil || pNew == nil {
		t.Fatalf("Create new project error: %v", err)
	}
	if len(pNew.TagBudgets()) != 2 {
		t.Errorf("expected 2 tag budgets, got %d", len(pNew.TagBudgets()))
	}
	if pNew.AnnualPlannedBudget() != (50000*12 + 100000*12) {
		t.Errorf("expected annual planned budget %d, got %d", 50000*12+100000*12, pNew.AnnualPlannedBudget())
	}

	// Create with invalid tag budget returns error
	_, err = svcNew.Create(context.Background(), project.CreateProjectCommand{
		PersonID: owner.PersonID,
		Name:     "Invalid Tag Budget Project",
		TagBudgets: []project.TagBudgetInput{
			{Tag: "materials", Frequency: "INVALID_FREQ", Amount: 50000},
		},
	})
	if err == nil {
		t.Errorf("expected error for invalid tag budget frequency")
	}

	// UpdateBudget
	pBudgetUpdated, err := svc.UpdateBudget(context.Background(), project.UpdateProjectBudgetCommand{
		PersonID:  owner.PersonID,
		ProjectID: proj.ID(),
		TagBudgets: []project.TagBudgetInput{
			{Tag: "rent", Frequency: "MONTHLY", Amount: 80000},
		},
	})
	if err != nil || len(pBudgetUpdated.TagBudgets()) != 1 {
		t.Errorf("UpdateBudget error: %v", err)
	}

	// Create error branches
	svcPeopleErr := project.NewProjectService(projRepo, &mockPeopleService{err: errors.New("err")})
	_, err = svcPeopleErr.Create(context.Background(), project.CreateProjectCommand{Name: "New Project"})
	if err == nil {
		t.Errorf("expected error when FindOwner fails")
	}

	svcCreateErr := project.NewProjectService(&mockProjectRepo{findErr: exceptions.NotFoundError, createErr: errors.New("err")}, peopleSvc)
	_, err = svcCreateErr.Create(context.Background(), project.CreateProjectCommand{Name: "New Project"})
	if err == nil {
		t.Errorf("expected error when repo Create fails")
	}

	svcUnknownErr := project.NewProjectService(&mockProjectRepo{findErr: errors.New("unknown")}, peopleSvc)
	_, err = svcUnknownErr.Create(context.Background(), project.CreateProjectCommand{Name: "New Project"})
	if err == nil {
		t.Errorf("expected error when repo FindOne returns unknown error")
	}

	// Create with invalid domain data returns domain validation error directly
	_, err = svcNew.Create(context.Background(), project.CreateProjectCommand{PersonID: owner.PersonID, Name: "A"})
	if err == nil {
		t.Errorf("expected validation error for short project name")
	}

	// Update project
	pUpdated, err := svc.Update(context.Background(), project.UpdateProjectCommand{
		PersonID:     owner.PersonID,
		ProjectID:    proj.ID(),
		Name:         "Updated Project",
		Description:  "Updated Desc",
		MainCurrency: "USD",
	})
	if err != nil || pUpdated.MainCurrency() != "USD" || pUpdated.Name() != "Updated Project" {
		t.Errorf("Update error or unexpected fields: %v", err)
	}

	// Update error branches
	_, err = svcNew.Update(context.Background(), project.UpdateProjectCommand{PersonID: owner.PersonID, ProjectID: core.NewProjectID()})
	if err == nil {
		t.Errorf("expected error when FindOne returns error")
	}

	svcSaveErr := project.NewProjectService(&mockProjectRepo{project: proj, updateErr: errors.New("err")}, peopleSvc)
	_, err = svcSaveErr.Update(context.Background(), project.UpdateProjectCommand{PersonID: owner.PersonID, ProjectID: proj.ID()})
	if err == nil {
		t.Errorf("expected error when repo Update fails")
	}
}

type mockTotalsCalculator struct {
	totals project.ProjectTotals
	err    error
}

func (m *mockTotalsCalculator) CalculateTotals(ctx context.Context, projectID core.ProjectID, homeCurrency string) (project.ProjectTotals, error) {
	if m.err != nil {
		return project.ProjectTotals{}, m.err
	}
	return m.totals, nil
}

func TestProjectService_Remove(t *testing.T) {
	ownerID := core.NewPersonID()
	otherPersonID := core.NewPersonID()
	owner := &project.Owner{PersonID: ownerID, DisplayName: "Owner"}
	proj, _ := project.NewProject(core.NewProjectID(), "Project to Remove", "Desc", owner, time.Now(), time.Now())

	t.Run("success when owner removes project", func(t *testing.T) {
		repo := &mockProjectRepo{project: proj}
		svc := project.NewProjectService(repo, &mockPeopleService{owner: owner})
		err := svc.Remove(context.Background(), project.RemoveProjectCommand{
			ProjectID: proj.ID(),
			PersonID:  ownerID,
		})
		if err != nil {
			t.Fatalf("expected nil error on owner remove, got: %v", err)
		}
	})

	t.Run("forbidden when non-owner attempts remove", func(t *testing.T) {
		repo := &mockProjectRepo{project: proj}
		svc := project.NewProjectService(repo, &mockPeopleService{owner: owner})
		err := svc.Remove(context.Background(), project.RemoveProjectCommand{
			ProjectID: proj.ID(),
			PersonID:  otherPersonID,
		})
		if err == nil {
			t.Fatal("expected forbidden error, got nil")
		}
		var exc exceptions.Exception
		if !errors.As(err, &exc) || exc.Code != exceptions.Forbidden {
			t.Errorf("expected Forbidden error type, got: %v", err)
		}
	})

	t.Run("not found when project does not exist", func(t *testing.T) {
		repo := &mockProjectRepo{findErr: exceptions.NotFoundError}
		svc := project.NewProjectService(repo, &mockPeopleService{owner: owner})
		err := svc.Remove(context.Background(), project.RemoveProjectCommand{
			ProjectID: core.NewProjectID(),
			PersonID:  ownerID,
		})
		if err == nil {
			t.Fatal("expected not found error, got nil")
		}
		var exc exceptions.Exception
		if !errors.As(err, &exc) || exc.Code != exceptions.NotFound {
			t.Errorf("expected NotFound error type, got: %v", err)
		}
	})
}

func TestProjectService_GetTotals(t *testing.T) {
	ownerID := core.NewPersonID()
	owner := &project.Owner{PersonID: ownerID, DisplayName: "Owner"}
	proj, _ := project.NewProject(core.NewProjectID(), "Project Totals", "Desc", owner, time.Now(), time.Now(), "USD")
	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, ownerID)

	t.Run("unauthenticated error", func(t *testing.T) {
		repo := &mockProjectRepo{project: proj}
		svc := project.NewProjectService(repo, &mockPeopleService{owner: owner})
		if _, err := svc.GetTotals(context.Background(), proj.ID()); err == nil {
			t.Fatal("expected unauthenticated error")
		}
	})

	t.Run("delegates to totals calculator", func(t *testing.T) {
		repo := &mockProjectRepo{project: proj}
		expectedTotals := project.ProjectTotals{
			TotalInvested: 50000,
			Balance:       30000,
			HasActivity:   true,
			Currency:      "USD",
		}
		calc := &mockTotalsCalculator{totals: expectedTotals}
		svc := project.NewProjectService(repo, &mockPeopleService{owner: owner}, calc)

		totals, err := svc.GetTotals(authedCtx, proj.ID())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if totals != expectedTotals {
			t.Errorf("expected %+v, got %+v", expectedTotals, totals)
		}
	})

	t.Run("returns empty totals with currency when calculator is nil", func(t *testing.T) {
		repo := &mockProjectRepo{project: proj}
		svc := project.NewProjectService(repo, &mockPeopleService{owner: owner})

		totals, err := svc.GetTotals(authedCtx, proj.ID())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if totals.Currency != "USD" || totals.TotalInvested != 0 {
			t.Errorf("unexpected totals: %+v", totals)
		}
	})

	t.Run("returns not found when project does not exist", func(t *testing.T) {
		repo := &mockProjectRepo{findErr: exceptions.NotFoundError}
		svc := project.NewProjectService(repo, &mockPeopleService{owner: owner})

		_, err := svc.GetTotals(authedCtx, core.NewProjectID())
		if err == nil {
			t.Fatal("expected not found error")
		}
		var exc exceptions.Exception
		if !errors.As(err, &exc) || exc.Code != exceptions.NotFound {
			t.Errorf("expected NotFound error type, got: %v", err)
		}
	})
}

func TestProjectSharing(t *testing.T) {
	ownerID := core.NewPersonID()
	owner := &project.Owner{PersonID: ownerID, DisplayName: "Owner"}
	proj, _ := project.NewProject(core.NewProjectID(), "Shared Project", "Desc", owner, time.Now(), time.Now())
	projRepo := &mockProjectRepo{project: proj}
	svc := project.NewProjectService(projRepo, &mockPeopleService{owner: owner})

	// AcceptShare by owner
	p, err := svc.AcceptShare(context.Background(), proj.ShareToken(), ownerID)
	if err != nil || p == nil {
		t.Fatalf("AcceptShare owner error: %v", err)
	}

	// AcceptShare by non-owner user
	recipientID := core.NewPersonID()
	pShared, err := svc.AcceptShare(context.Background(), proj.ShareToken(), recipientID)
	if err != nil || pShared == nil {
		t.Fatalf("AcceptShare recipient error: %v", err)
	}
	if !pShared.IsShared() {
		t.Errorf("expected IsShared to be true for recipient")
	}

	// AcceptShare with error
	errRepo := &mockProjectRepo{findErr: errors.New("not found")}
	svcErr := project.NewProjectService(errRepo, &mockPeopleService{owner: owner})
	_, err = svcErr.AcceptShare(context.Background(), uuid.New(), recipientID)
	if err == nil {
		t.Errorf("expected error when share token not found")
	}
}
