package projecta_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
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

	// CostCategory
	catID := uuid.New()
	cat, err := projecta.NewCostCategory(catID, projID, "Category 1", "Desc")
	if err != nil {
		t.Fatalf("unexpected category error: %v", err)
	}
	if cat.ID != catID || cat.ProjectID != projID || cat.Name != "Category 1" {
		t.Errorf("category fields mismatch")
	}

	_, err = projecta.NewCostCategory(catID, projID, "AB", "Desc")
	if err == nil {
		t.Errorf("expected error for short category name")
	}

	catCol := projecta.NewCategoryCollection(10)
	if catCol.Total() != 10 {
		t.Errorf("category collection total mismatch")
	}

	// CostType
	costType, err := projecta.NewCostType(projID, cat, "Type 1", "Desc")
	if err != nil {
		t.Fatalf("unexpected cost type error: %v", err)
	}
	if costType.ProjectID != projID || costType.Category != cat {
		t.Errorf("cost type fields mismatch")
	}

	typeCol := projecta.NewCostTypeCollection(5)
	if typeCol.Total() != 5 {
		t.Errorf("cost type collection total mismatch")
	}

	// PaymentKind & Payment
	pk, err := projecta.ToPaymentKind("DOWN_PAYMENT")
	if err != nil || pk != projecta.DownPayment || pk.String() != "DOWN_PAYMENT" {
		t.Errorf("payment kind DOWN_PAYMENT mismatch")
	}
	pkUp, _ := projecta.ToPaymentKind("UPON_COMPLETION")
	if pkUp != projecta.UponCompletionPayment {
		t.Errorf("payment kind UPON_COMPLETION mismatch")
	}
	pkCred, _ := projecta.ToPaymentKind("CREDIT_PAYMENT")
	if pkCred != projecta.CreditPayment {
		t.Errorf("payment kind CREDIT_PAYMENT mismatch")
	}

	_, err = projecta.ToPaymentKind("INVALID")
	if err == nil {
		t.Errorf("expected error for invalid payment kind")
	}

	payID := uuid.New()
	amt := money.New(100, money.USD)
	pay := projecta.NewPayment(payID, proj, owner, costType, "Pay desc", amt, now, projecta.DownPayment)
	if pay.ID != payID || pay.Amount != amt || pay.Kind != projecta.DownPayment {
		t.Errorf("payment fields mismatch")
	}

	payCol := projecta.NewPaymentCollection(3)
	if payCol.Total() != 3 {
		t.Errorf("payment collection total mismatch")
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
	findErr   error
	createErr error
}

func (m *mockProjectRepo) Find(ctx context.Context, filter projecta.ProjectCollectionFilter) ([]*projecta.Project, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return []*projecta.Project{m.project}, nil
}
func (m *mockProjectRepo) FindOne(ctx context.Context, filter projecta.ProjectFilter) (*projecta.Project, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return m.project, nil
}
func (m *mockProjectRepo) Create(ctx context.Context, p *projecta.Project) error {
	return m.createErr
}
func (m *mockProjectRepo) Update(ctx context.Context, p *projecta.Project) error { return nil }
func (m *mockProjectRepo) Remove(ctx context.Context, p *projecta.Project) error { return nil }

type mockCategoryRepo struct {
	cat        *projecta.CostCategory
	findErr    error
	findOneErr error
	saveErr    error
	removeErr  error
}

func (m *mockCategoryRepo) Find(ctx context.Context, filter projecta.CategoryCollectionFilter) (*projecta.CostCategoryCollection, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return projecta.NewCategoryCollection(1), nil
}
func (m *mockCategoryRepo) FindOne(ctx context.Context, filter projecta.CategoryFilter) (*projecta.CostCategory, error) {
	if m.findOneErr != nil {
		return nil, m.findOneErr
	}
	return m.cat, nil
}
func (m *mockCategoryRepo) Save(ctx context.Context, c *projecta.CostCategory) error {
	return m.saveErr
}
func (m *mockCategoryRepo) Remove(ctx context.Context, c *projecta.CostCategory) error {
	return m.removeErr
}

type mockTypeRepo struct {
	costType   *projecta.CostType
	findErr    error
	findOneErr error
	saveErr    error
	removeErr  error
}

func (m *mockTypeRepo) Find(ctx context.Context, filter projecta.TypeCollectionFilter) (*projecta.CostTypeCollection, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return projecta.NewCostTypeCollection(1), nil
}
func (m *mockTypeRepo) FindOne(ctx context.Context, filter projecta.TypeFilter) (*projecta.CostType, error) {
	if m.findOneErr != nil {
		return nil, m.findOneErr
	}
	return m.costType, nil
}
func (m *mockTypeRepo) Save(ctx context.Context, t *projecta.CostType) error { return m.saveErr }
func (m *mockTypeRepo) Remove(ctx context.Context, t *projecta.CostType) error {
	return m.removeErr
}

type mockPaymentRepo struct {
	pay        *projecta.Payment
	findErr    error
	findOneErr error
	saveErr    error
	removeErr  error
}

func (m *mockPaymentRepo) Find(ctx context.Context, filter projecta.PaymentCollectionFilter) (*projecta.PaymentCollection, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	return projecta.NewPaymentCollection(1), nil
}
func (m *mockPaymentRepo) FindOne(ctx context.Context, filter projecta.PaymentFilter) (*projecta.Payment, error) {
	if m.findOneErr != nil {
		return nil, m.findOneErr
	}
	return m.pay, nil
}
func (m *mockPaymentRepo) Save(ctx context.Context, p *projecta.Payment) error { return m.saveErr }
func (m *mockPaymentRepo) Remove(ctx context.Context, p *projecta.Payment) error {
	return m.removeErr
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
func (m *mockPeopleRepo) Register(ctx context.Context, p *people.Person) error { return nil }
func (m *mockPeopleRepo) FindCredentials(ctx context.Context, provider people.IdentityProvider, regID string) (uuid.UUID, string, error) {
	return uuid.Nil, "", nil
}

func TestProjectService(t *testing.T) {
	owner := &projecta.Owner{PersonID: uuid.New(), DisplayName: "John"}
	proj, _ := projecta.NewProject(uuid.New(), "Project Alpha", "Desc", owner, time.Now(), time.Now())

	peopleSvc := &mockPeopleService{owner: owner}
	projRepo := &mockProjectRepo{project: proj}

	svc := projecta.NewProjectService(projRepo, peopleSvc)

	// Find & FindOne
	pList, err := svc.Find(context.Background(), projecta.ProjectCollectionFilter{})
	if err != nil || len(pList) != 1 {
		t.Errorf("Find error: %v", err)
	}

	pOne, err := svc.FindOne(context.Background(), projecta.ProjectFilter{})
	if err != nil || pOne != proj {
		t.Errorf("FindOne error: %v", err)
	}

	// Create when project already exists
	pExist, err := svc.Create(context.Background(), projecta.CreateProjectCommand{PersonID: owner.PersonID, Name: "Project Alpha"})
	if err != nil || pExist != proj {
		t.Errorf("Create existing project mismatch")
	}

	// Create new project (FindOne returns NotFoundError)
	projRepoNotFound := &mockProjectRepo{findErr: exceptions.NotFoundError}
	svcNew := projecta.NewProjectService(projRepoNotFound, peopleSvc)
	pNew, err := svcNew.Create(context.Background(), projecta.CreateProjectCommand{PersonID: owner.PersonID, Name: "New Project", Description: "Desc"})
	if err != nil || pNew == nil {
		t.Fatalf("Create new project error: %v", err)
	}

	// Create error branches
	svcPeopleErr := projecta.NewProjectService(projRepo, &mockPeopleService{err: errors.New("err")})
	_, err = svcPeopleErr.Create(context.Background(), projecta.CreateProjectCommand{})
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

	// Test unimplemented method panics
	defer func() { _ = recover() }()
	_ = svc.Save(context.Background(), nil)
}

func TestUnimplementedPanics(t *testing.T) {
	svc := projecta.NewProjectService(&mockProjectRepo{}, &mockPeopleService{})
	typeSvc := projecta.NewTypeService(&mockTypeRepo{}, &mockCategoryRepo{}, &mockProjectRepo{})

	t.Run("ProjectService Remove panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		_ = svc.Remove(context.Background(), projecta.RemoveProjectCommand{})
	})

	t.Run("ProjectService Update panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		_ = svc.Update(context.Background(), projecta.UpdateProjectCommand{})
	})

	t.Run("TypeService Update panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		_ = typeSvc.Update(context.Background(), projecta.UpdateTypeCommand{})
	})
}

func TestCategoryService(t *testing.T) {
	owner := &projecta.Owner{PersonID: uuid.New()}
	proj, _ := projecta.NewProject(uuid.New(), "Project", "Desc", owner, time.Now(), time.Now())
	cat, _ := projecta.NewCostCategory(uuid.New(), proj.ProjectID, "Category", "Desc")

	catRepo := &mockCategoryRepo{cat: cat}
	projSvc := projecta.NewProjectService(&mockProjectRepo{project: proj}, &mockPeopleService{owner: owner})
	svc := projecta.NewCategoryService(catRepo, projSvc)

	// Find
	cols, err := svc.Find(context.Background(), projecta.CategoryCollectionFilter{})
	if err != nil || cols == nil {
		t.Errorf("Find error: %v", err)
	}

	svcFindErr := projecta.NewCategoryService(&mockCategoryRepo{findErr: errors.New("err")}, projSvc)
	_, err = svcFindErr.Find(context.Background(), projecta.CategoryCollectionFilter{})
	if err == nil {
		t.Errorf("expected Find error")
	}

	// Create
	newCat, err := svc.Create(context.Background(), projecta.CreateCategoryCommand{ProjectID: proj.ProjectID, Name: "New Cat", Description: "Desc"})
	if err != nil || newCat == nil {
		t.Fatalf("Create error: %v", err)
	}

	svcSaveErr := projecta.NewCategoryService(&mockCategoryRepo{saveErr: errors.New("err")}, projSvc)
	_, err = svcSaveErr.Create(context.Background(), projecta.CreateCategoryCommand{ProjectID: proj.ProjectID, Name: "New Cat"})
	if err == nil {
		t.Errorf("expected Save error")
	}

	// Create error branches
	projSvcErr := projecta.NewCategoryService(catRepo, projecta.NewProjectService(&mockProjectRepo{findErr: errors.New("err")}, &mockPeopleService{}))
	_, err = projSvcErr.Create(context.Background(), projecta.CreateCategoryCommand{ProjectID: proj.ProjectID})
	if err == nil {
		t.Errorf("expected error when project FindOne fails")
	}

	projSvcNil := projecta.NewCategoryService(catRepo, projecta.NewProjectService(&mockProjectRepo{project: nil}, &mockPeopleService{}))
	_, err = projSvcNil.Create(context.Background(), projecta.CreateCategoryCommand{ProjectID: proj.ProjectID})
	if err == nil {
		t.Errorf("expected error when project is nil")
	}

	// Update & Remove
	err = svc.Update(context.Background(), projecta.UpdateCategoryCommand{ID: cat.ID, ProjectID: proj.ProjectID, Name: "Updated", Description: "Updated"})
	if err != nil {
		t.Errorf("Update error: %v", err)
	}

	svcUpdFindErr := projecta.NewCategoryService(&mockCategoryRepo{findOneErr: errors.New("err")}, projSvc)
	err = svcUpdFindErr.Update(context.Background(), projecta.UpdateCategoryCommand{})
	if err == nil {
		t.Errorf("expected error on Update FindOne")
	}

	svcUpdSaveErr := projecta.NewCategoryService(&mockCategoryRepo{cat: cat, saveErr: errors.New("err")}, projSvc)
	err = svcUpdSaveErr.Update(context.Background(), projecta.UpdateCategoryCommand{})
	if err == nil {
		t.Errorf("expected error on Update Save")
	}

	err = svc.Remove(context.Background(), projecta.RemoveCategoryCommand{ID: cat.ID, ProjectID: proj.ProjectID})
	if err != nil {
		t.Errorf("Remove error: %v", err)
	}

	svcRemFindErr := projecta.NewCategoryService(&mockCategoryRepo{findOneErr: errors.New("err")}, projSvc)
	err = svcRemFindErr.Remove(context.Background(), projecta.RemoveCategoryCommand{})
	if err == nil {
		t.Errorf("expected error on Remove FindOne")
	}

	svcRemErr := projecta.NewCategoryService(&mockCategoryRepo{cat: cat, removeErr: errors.New("err")}, projSvc)
	err = svcRemErr.Remove(context.Background(), projecta.RemoveCategoryCommand{})
	if err == nil {
		t.Errorf("expected error on Remove")
	}
}

func TestTypeService(t *testing.T) {
	owner := &projecta.Owner{PersonID: uuid.New()}
	proj, _ := projecta.NewProject(uuid.New(), "Project", "Desc", owner, time.Now(), time.Now())
	cat, _ := projecta.NewCostCategory(uuid.New(), proj.ProjectID, "Cat", "Desc")
	costType, _ := projecta.NewCostType(proj.ProjectID, cat, "Type", "Desc")

	typeRepo := &mockTypeRepo{costType: costType}
	catRepo := &mockCategoryRepo{cat: cat}
	projRepo := &mockProjectRepo{project: proj}

	svc := projecta.NewTypeService(typeRepo, catRepo, projRepo)

	// Find & FindOne
	_, err := svc.FindOne(context.Background(), projecta.TypeFilter{})
	if err != nil {
		t.Errorf("FindOne error: %v", err)
	}

	_, err = svc.Find(context.Background(), projecta.TypeCollectionFilter{})
	if err != nil {
		t.Errorf("Find error: %v", err)
	}

	// Create
	createdType, err := svc.Create(context.Background(), projecta.CreateTypeCommand{ProjectID: proj.ProjectID, CategoryID: cat.ID, Name: "New Type", Description: "Desc"})
	if err != nil || createdType == nil {
		t.Fatalf("Create error: %v", err)
	}

	// Create error branches
	svcProjErr := projecta.NewTypeService(typeRepo, catRepo, &mockProjectRepo{findErr: errors.New("err")})
	_, err = svcProjErr.Create(context.Background(), projecta.CreateTypeCommand{})
	if err == nil {
		t.Errorf("expected error on project FindOne")
	}

	svcCatErr := projecta.NewTypeService(typeRepo, &mockCategoryRepo{findOneErr: errors.New("err")}, projRepo)
	_, err = svcCatErr.Create(context.Background(), projecta.CreateTypeCommand{})
	if err == nil {
		t.Errorf("expected error on category FindOne")
	}

	svcSaveErr := projecta.NewTypeService(&mockTypeRepo{saveErr: errors.New("err")}, catRepo, projRepo)
	_, err = svcSaveErr.Create(context.Background(), projecta.CreateTypeCommand{Name: "New Type"})
	if err == nil {
		t.Errorf("expected error on type Save")
	}

	// Remove success, not found, and error
	err = svc.Remove(context.Background(), projecta.RemoveProjectResourceCommand{ResourceID: costType.ID, ProjectID: proj.ProjectID})
	if err != nil {
		t.Errorf("Remove error: %v", err)
	}

	svcNotFound := projecta.NewTypeService(&mockTypeRepo{findOneErr: exceptions.NotFoundError}, catRepo, projRepo)
	err = svcNotFound.Remove(context.Background(), projecta.RemoveProjectResourceCommand{})
	if err == nil {
		t.Errorf("expected not found error")
	}

	svcErr := projecta.NewTypeService(&mockTypeRepo{findOneErr: errors.New("err")}, catRepo, projRepo)
	err = svcErr.Remove(context.Background(), projecta.RemoveProjectResourceCommand{})
	if err == nil {
		t.Errorf("expected internal error")
	}

	svcRemErr := projecta.NewTypeService(&mockTypeRepo{costType: costType, removeErr: errors.New("err")}, catRepo, projRepo)
	err = svcRemErr.Remove(context.Background(), projecta.RemoveProjectResourceCommand{})
	if err == nil {
		t.Errorf("expected remove error")
	}
}

func TestPaymentService(t *testing.T) {
	requesterID := uuid.New()
	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, requesterID)

	owner := &projecta.Owner{PersonID: requesterID, DisplayName: "John"}
	proj, _ := projecta.NewProject(uuid.New(), "Project", "Desc", owner, time.Now(), time.Now())
	cat, _ := projecta.NewCostCategory(uuid.New(), proj.ProjectID, "Cat", "Desc")
	costType, _ := projecta.NewCostType(proj.ProjectID, cat, "Type", "Desc")
	pay := projecta.NewPayment(uuid.New(), proj, owner, costType, "Payment", money.New(100, money.USD), time.Now(), projecta.DownPayment)

	payRepo := &mockPaymentRepo{pay: pay}
	typeRepo := &mockTypeRepo{costType: costType}
	projRepo := &mockProjectRepo{project: proj}
	peopleSvc := &mockPeopleService{owner: owner}

	svc := projecta.NewPaymentService(payRepo, typeRepo, projRepo, peopleSvc)

	// Create
	createdPay, err := svc.Create(authedCtx, projecta.CreatePaymentCommand{
		ProjectID:   proj.ProjectID,
		TypeID:      costType.ID,
		Description: "New Pay",
		Amount:      money.New(500, money.USD),
		Kind:        projecta.DownPayment,
	})
	if err != nil || createdPay == nil {
		t.Fatalf("Create error: %v", err)
	}

	// Create error branches
	svcTypeErr := projecta.NewPaymentService(payRepo, &mockTypeRepo{findOneErr: errors.New("err")}, projRepo, peopleSvc)
	_, err = svcTypeErr.Create(authedCtx, projecta.CreatePaymentCommand{})
	if err == nil {
		t.Errorf("expected error on type FindOne")
	}

	svcProjErr := projecta.NewPaymentService(payRepo, typeRepo, &mockProjectRepo{findErr: errors.New("err")}, peopleSvc)
	_, err = svcProjErr.Create(authedCtx, projecta.CreatePaymentCommand{})
	if err == nil {
		t.Errorf("expected error on project FindOne")
	}

	svcSaveErr := projecta.NewPaymentService(&mockPaymentRepo{saveErr: errors.New("err")}, typeRepo, projRepo, peopleSvc)
	_, err = svcSaveErr.Create(authedCtx, projecta.CreatePaymentCommand{})
	if err == nil {
		t.Errorf("expected error on payment Save")
	}

	// Create unauthorized (nil requesterID)
	_, err = svc.Create(context.WithValue(context.Background(), core.RequesterIDContextKey, uuid.Nil), projecta.CreatePaymentCommand{})
	if err == nil {
		t.Errorf("expected unauthorized error")
	}

	// Find & FindOne
	_, err = svc.Find(authedCtx, projecta.PaymentCollectionFilter{})
	if err != nil {
		t.Errorf("Find error: %v", err)
	}

	svcFindErr := projecta.NewPaymentService(&mockPaymentRepo{findErr: errors.New("err")}, typeRepo, projRepo, peopleSvc)
	_, err = svcFindErr.Find(authedCtx, projecta.PaymentCollectionFilter{})
	if err == nil {
		t.Errorf("expected Find error")
	}

	_, err = svc.FindOne(authedCtx, projecta.PaymentFilter{})
	if err != nil {
		t.Errorf("FindOne error: %v", err)
	}

	svcPayNotFound := projecta.NewPaymentService(&mockPaymentRepo{findOneErr: exceptions.NotFoundError}, typeRepo, projRepo, peopleSvc)
	_, err = svcPayNotFound.FindOne(authedCtx, projecta.PaymentFilter{})
	if err == nil {
		t.Errorf("expected not found error")
	}

	svcPayFindErr := projecta.NewPaymentService(&mockPaymentRepo{findOneErr: errors.New("err")}, typeRepo, projRepo, peopleSvc)
	_, err = svcPayFindErr.FindOne(authedCtx, projecta.PaymentFilter{})
	if err == nil {
		t.Errorf("expected internal error")
	}

	// Update
	updCmd := projecta.UpdatePaymentCommand{
		ID:          pay.ID,
		ProjectID:   proj.ProjectID,
		TypeID:      costType.ID,
		Description: "Updated Pay",
		Amount:      money.New(600, money.USD),
		Kind:        projecta.CreditPayment,
	}
	err = svc.Update(authedCtx, updCmd)
	if err != nil {
		t.Errorf("Update error: %v", err)
	}

	err = svcPayNotFound.Update(authedCtx, updCmd)
	if err == nil {
		t.Errorf("expected not found error on Update")
	}

	err = svcPayFindErr.Update(authedCtx, updCmd)
	if err == nil {
		t.Errorf("expected internal error on Update FindOne")
	}

	svcUpdTypeErr := projecta.NewPaymentService(payRepo, &mockTypeRepo{findOneErr: errors.New("err")}, projRepo, peopleSvc)
	err = svcUpdTypeErr.Update(authedCtx, updCmd)
	if err == nil {
		t.Errorf("expected type FindOne error on Update")
	}

	// Remove
	err = svc.Remove(authedCtx, projecta.RemovePaymentCommand{ID: pay.ID, ProjectID: proj.ProjectID})
	if err != nil {
		t.Errorf("Remove error: %v", err)
	}

	err = svcPayNotFound.Remove(authedCtx, projecta.RemovePaymentCommand{})
	if err == nil {
		t.Errorf("expected not found error on Remove")
	}

	err = svcPayFindErr.Remove(authedCtx, projecta.RemovePaymentCommand{})
	if err == nil {
		t.Errorf("expected internal error on Remove")
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
