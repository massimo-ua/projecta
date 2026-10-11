package projecta_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

func TestDomainEntities(t *testing.T) {
	ownerID := uuid.New()
	owner := &projecta.Owner{PersonID: ownerID, DisplayName: "John Doe"}
	now := time.Now()

	// Project
	projID := uuid.New()
	proj, err := projecta.NewProject(projID, "Test Project", "Description", owner, now, now)
	if err != nil {
		t.Fatalf("unexpected project error: %v", err)
	}
	if !proj.IsOwnedBy(owner) {
		t.Errorf("IsOwnedBy mismatch")
	}
	otherOwner := &projecta.Owner{PersonID: uuid.New()}
	if proj.IsOwnedBy(otherOwner) {
		t.Errorf("IsOwnedBy should be false for other owner")
	}

	// Invalid project name
	_, err = projecta.NewProject(projID, "A", "Desc", owner, now, now)
	if err == nil {
		t.Errorf("expected error for short project name")
	}
}

// Mock Repositories for projecta services
type mockPeopleService struct {
	owner *projecta.Owner
	err   error
}

func (m *mockPeopleService) FindOwner(ctx context.Context, personID uuid.UUID) (*projecta.Owner, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.owner, nil
}

type mockProjectRepo struct {
	project   *projecta.Project
	projects  []*projecta.Project
	findErr   error
	createErr error
	updateErr error
	removeErr error
	shareErr  error
}

func (m *mockProjectRepo) Find(ctx context.Context, filter projecta.ProjectCollectionFilter) ([]*projecta.Project, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.projects != nil {
		return m.projects, nil
	}
	if m.project != nil {
		return []*projecta.Project{m.project}, nil
	}
	return []*projecta.Project{}, nil
}

func (m *mockProjectRepo) FindOne(ctx context.Context, filter projecta.ProjectFilter) (*projecta.Project, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.project != nil {
		return m.project, nil
	}
	return nil, exceptions.NotFoundError
}

func (m *mockProjectRepo) FindByShareToken(ctx context.Context, token uuid.UUID) (*projecta.Project, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.project != nil {
		return m.project, nil
	}
	return nil, exceptions.NotFoundError
}

func (m *mockProjectRepo) Create(ctx context.Context, project *projecta.Project) error {
	return m.createErr
}

func (m *mockProjectRepo) Update(ctx context.Context, project *projecta.Project) error {
	return m.updateErr
}

func (m *mockProjectRepo) Remove(ctx context.Context, project *projecta.Project) error {
	return m.removeErr
}

func (m *mockProjectRepo) CreateShareRecord(ctx context.Context, projectID uuid.UUID, personID uuid.UUID) error {
	return m.shareErr
}

type mockPeopleRepo struct {
	person *people.Person
	err    error
}

func (m *mockPeopleRepo) FindByID(ctx context.Context, id uuid.UUID) (*people.Person, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.person, nil
}

func TestProjectService(t *testing.T) {
	ownerID := uuid.New()
	owner := &projecta.Owner{PersonID: ownerID, DisplayName: "John", CanHaveProjects: true}
	peopleSvc := &mockPeopleService{owner: owner}

	proj, _ := projecta.NewProject(uuid.New(), "Project 1", "Desc", owner, time.Now(), time.Now())
	projRepo := &mockProjectRepo{project: proj}

	svc := projecta.NewProjectService(projRepo, peopleSvc)

	// FindOne
	p, err := svc.FindOne(context.Background(), projecta.ProjectFilter{})
	if err != nil || p == nil {
		t.Fatalf("FindOne error: %v", err)
	}

	// Find
	list, err := svc.Find(context.Background(), projecta.ProjectCollectionFilter{})
	if err != nil || len(list) != 1 {
		t.Fatalf("Find error: %v", err)
	}

	// Create existing project (returns existing)
	pExisting, err := svc.Create(context.Background(), projecta.CreateProjectCommand{PersonID: owner.PersonID, Name: "Project 1"})
	if err != nil || pExisting == nil {
		t.Fatalf("Create existing project error: %v", err)
	}

	// Create new project
	projRepoNotFound := &mockProjectRepo{findErr: exceptions.NotFoundError}
	svcNew := projecta.NewProjectService(projRepoNotFound, peopleSvc)
	pNew, err := svcNew.Create(context.Background(), projecta.CreateProjectCommand{PersonID: owner.PersonID, Name: "New Project", Description: "Desc"})
	if err != nil || pNew == nil {
		t.Fatalf("Create new project error: %v", err)
	}

	// Create error branches
	svcPeopleErr := projecta.NewProjectService(projRepo, &mockPeopleService{err: errors.New("err")})
	_, err = svcPeopleErr.Create(context.Background(), projecta.CreateProjectCommand{Name: "New Project"})
	if err == nil {
		t.Errorf("expected error when FindOwner fails")
	}

	svcCreateErr := projecta.NewProjectService(&mockProjectRepo{findErr: exceptions.NotFoundError, createErr: errors.New("err")}, peopleSvc)
	_, err = svcCreateErr.Create(context.Background(), projecta.CreateProjectCommand{Name: "New Project"})
	if err == nil {
		t.Errorf("expected error when repo Create fails")
	}

	svcUnknownErr := projecta.NewProjectService(&mockProjectRepo{findErr: errors.New("unknown")}, peopleSvc)
	_, err = svcUnknownErr.Create(context.Background(), projecta.CreateProjectCommand{Name: "New Project"})
	if err == nil {
		t.Errorf("expected error when repo FindOne returns unknown error")
	}

	// Update project
	pUpdated, err := svc.Update(context.Background(), projecta.UpdateProjectCommand{
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
	_, err = svcNew.Update(context.Background(), projecta.UpdateProjectCommand{PersonID: owner.PersonID, ProjectID: uuid.New()})
	if err == nil {
		t.Errorf("expected error when FindOne returns error")
	}

	svcSaveErr := projecta.NewProjectService(&mockProjectRepo{project: proj, updateErr: errors.New("err")}, peopleSvc)
	_, err = svcSaveErr.Update(context.Background(), projecta.UpdateProjectCommand{PersonID: owner.PersonID, ProjectID: proj.ID()})
	if err == nil {
		t.Errorf("expected error when repo Update fails")
	}
}

func TestUnimplementedPanics(t *testing.T) {
	svc := projecta.NewProjectService(&mockProjectRepo{}, &mockPeopleService{})

	t.Run("ProjectService Remove panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		_ = svc.Remove(context.Background(), projecta.RemoveProjectCommand{})
	})
}

func TestProjectSharing(t *testing.T) {
	ownerID := uuid.New()
	owner := &projecta.Owner{PersonID: ownerID, DisplayName: "Owner"}
	proj, _ := projecta.NewProject(uuid.New(), "Shared Project", "Desc", owner, time.Now(), time.Now())
	projRepo := &mockProjectRepo{project: proj}
	svc := projecta.NewProjectService(projRepo, &mockPeopleService{owner: owner})

	// AcceptShare by owner
	p, err := svc.AcceptShare(context.Background(), proj.ShareToken(), ownerID)
	if err != nil || p == nil {
		t.Fatalf("AcceptShare owner error: %v", err)
	}

	// AcceptShare by non-owner user
	recipientID := uuid.New()
	pShared, err := svc.AcceptShare(context.Background(), proj.ShareToken(), recipientID)
	if err != nil || pShared == nil {
		t.Fatalf("AcceptShare recipient error: %v", err)
	}
	if !pShared.IsShared() {
		t.Errorf("expected IsShared to be true for recipient")
	}

	// AcceptShare with error
	errRepo := &mockProjectRepo{findErr: errors.New("not found")}
	svcErr := projecta.NewProjectService(errRepo, &mockPeopleService{owner: owner})
	_, err = svcErr.AcceptShare(context.Background(), uuid.New(), recipientID)
	if err == nil {
		t.Errorf("expected error when share token not found")
	}
}

func TestPeopleService(t *testing.T) {
	pID := uuid.New()
	cred, _ := people.NewCredentials("LOCAL", "user@example.com", "secret")
	p, _ := people.NewPerson(pID, "John", "Doe", "J.D.", []people.Credentials{cred})

	svc := projecta.NewPeopleService(&mockPeopleRepo{person: p})
	owner, err := svc.FindOwner(context.Background(), pID)
	if err != nil || owner == nil {
		t.Fatalf("FindOwner error: %v", err)
	}
	if owner.PersonID != pID || owner.DisplayName != "J.D." {
		t.Errorf("owner fields mismatch")
	}

	svcErr := projecta.NewPeopleService(&mockPeopleRepo{err: errors.New("err")})
	_, err = svcErr.FindOwner(context.Background(), pID)
	if err == nil {
		t.Errorf("expected error when FindByID fails")
	}
}

func TestStatementService(t *testing.T) {
	projID := uuid.New()
	owner := &projecta.Owner{PersonID: uuid.New(), DisplayName: "Owner"}
	proj, _ := projecta.NewProject(projID, "Test Project", "Desc", owner, time.Now(), time.Now())
	projRepo := &mockProjectRepo{project: proj}
	projSvc := projecta.NewProjectService(projRepo, &mockPeopleService{owner: owner})

	svc := projecta.NewStatementService(projSvc)

	t.Run("nil project id", func(t *testing.T) {
		_, err := svc.ParseStatement(context.Background(), uuid.Nil, []byte("data"))
		if err == nil {
			t.Errorf("expected error for nil project id")
		}
	})

	t.Run("project not found", func(t *testing.T) {
		notFoundSvc := projecta.NewStatementService(projecta.NewProjectService(&mockProjectRepo{findErr: exceptions.NotFoundError}, &mockPeopleService{owner: owner}))
		_, err := notFoundSvc.ParseStatement(context.Background(), projID, []byte("data"))
		if err == nil {
			t.Errorf("expected error for project not found")
		}
	})

	t.Run("invalid PDF data", func(t *testing.T) {
		_, err := svc.ParseStatement(context.Background(), projID, []byte("not a pdf"))
		if err == nil {
			t.Errorf("expected error for invalid PDF data")
		}
	})
}
