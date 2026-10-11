package asset_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
)

type mockDb struct{}

func (m *mockDb) Tx(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	return fn(ctx)
}
func (m *mockDb) Close()                        {}
func (m *mockDb) Ping(ctx context.Context) error { return nil }

type mockAssetRepo struct {
	saveErr    error
	removeErr  error
	findErr    error
	findOneErr error
	asset      *asset.Asset
	col        *asset.Collection
}

func (m *mockAssetRepo) Save(ctx context.Context, a *asset.Asset) error { return m.saveErr }
func (m *mockAssetRepo) Remove(ctx context.Context, a *asset.Asset) error {
	return m.removeErr
}
func (m *mockAssetRepo) Find(ctx context.Context, filter asset.CollectionFilter) (*asset.Collection, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.col != nil {
		return m.col, nil
	}
	return asset.NewCollection(0), nil
}
func (m *mockAssetRepo) FindOne(ctx context.Context, filter asset.Filter) (*asset.Asset, error) {
	if m.findOneErr != nil {
		return nil, m.findOneErr
	}
	return m.asset, nil
}
func (m *mockAssetRepo) AddChild(ctx context.Context, parentID uuid.UUID, childID uuid.UUID, sharePercentage float64) error {
	return nil
}
func (m *mockAssetRepo) RemoveChild(ctx context.Context, parentID uuid.UUID, childID uuid.UUID) error {
	return nil
}
func (m *mockAssetRepo) FindChildren(ctx context.Context, parentID uuid.UUID) ([]asset.ChildAssetLink, error) {
	return nil, nil
}
func (m *mockAssetRepo) FindParents(ctx context.Context, childID uuid.UUID) ([]asset.ParentAssetLink, error) {
	return nil, nil
}
func (m *mockAssetRepo) FindAncestors(ctx context.Context, assetID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

type mockPeopleService struct {
	owner *projecta.Owner
	err   error
}

func (m *mockPeopleService) FindOwner(ctx context.Context, id uuid.UUID) (*projecta.Owner, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.owner, nil
}

type mockProjectRepo struct {
	project *projecta.Project
	err     error
}

func (m *mockProjectRepo) Create(ctx context.Context, p *projecta.Project) error { return nil }
func (m *mockProjectRepo) Update(ctx context.Context, p *projecta.Project) error { return nil }
func (m *mockProjectRepo) Remove(ctx context.Context, p *projecta.Project) error { return nil }
func (m *mockProjectRepo) Find(ctx context.Context, filter projecta.ProjectCollectionFilter) ([]*projecta.Project, error) {
	return nil, nil
}
func (m *mockProjectRepo) FindOne(ctx context.Context, filter projecta.ProjectFilter) (*projecta.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.project, nil
}
func (m *mockProjectRepo) FindByShareToken(ctx context.Context, token uuid.UUID) (*projecta.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.project, nil
}
func (m *mockProjectRepo) CreateShareRecord(ctx context.Context, projectID uuid.UUID, personID uuid.UUID) error {
	return m.err
}

type mockInvestmentSource struct {
	items         []asset.InvestmentItem
	findErr       error
	assignErr     error
	assignedAsset uuid.UUID
	assignedIDs   []uuid.UUID
	createErr     error
	createdItem   asset.InitialInvestment
}

func (m *mockInvestmentSource) FindInvestments(ctx context.Context, projectID uuid.UUID, ids []uuid.UUID) ([]asset.InvestmentItem, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var matched []asset.InvestmentItem
	for _, id := range ids {
		for _, item := range m.items {
			if item.ID == id {
				matched = append(matched, item)
				break
			}
		}
	}
	return matched, nil
}

func (m *mockInvestmentSource) AssignToAsset(ctx context.Context, assetID uuid.UUID, investmentIDs []uuid.UUID) error {
	if m.assignErr != nil {
		return m.assignErr
	}
	m.assignedAsset = assetID
	m.assignedIDs = investmentIDs
	return nil
}

func (m *mockInvestmentSource) CreateInitialInvestment(ctx context.Context, item asset.InitialInvestment) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.createdItem = item
	return nil
}

type mockRateProvider struct {
	err error
}

func (m *mockRateProvider) Convert(a currency.Currency, b currency.Currency) (currency.Currency, error) {
	if m.err != nil {
		return currency.Currency{}, m.err
	}
	if a.Code == "USD" && b.Code == "UAH" {
		return currency.Currency{Amount: a.Amount * 40, Code: b.Code}, nil
	}
	return currency.Currency{Amount: a.Amount, Code: b.Code}, nil
}

func newTestService(
	db core.DbConnection,
	assets asset.Repository,
	people projecta.PeopleService,
	projects projecta.ProjectRepository,
	inv asset.InvestmentSource,
	rp ...currency.CurrencyRateProvider,
) *asset.ServiceImpl {
	return asset.NewService(db, assets, people, projects, inv, rp...)
}

func TestAssetService(t *testing.T) {
	requesterID := uuid.New()
	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, requesterID)

	now := time.Now()
	owner := &projecta.Owner{PersonID: requesterID, DisplayName: "John Doe"}
	project, _ := projecta.NewProject(uuid.New(), "Project 1", "Desc", owner, now, now)
	costType, _ := projecta.NewCostType(project.ProjectID, nil, "Type 1", "Desc")
	existingAsset := asset.NewAsset(uuid.New(), "Laptop", "Work Laptop", project, costType, money.New(1000, money.USD), now, owner)

	t.Run("Find unauthorized and success", func(t *testing.T) {
		svc := newTestService(&mockDb{}, &mockAssetRepo{}, &mockPeopleService{}, &mockProjectRepo{}, nil)

		_, err := svc.Find(context.Background(), asset.CollectionFilter{})
		if err == nil {
			t.Errorf("expected unauthorized error")
		}

		res, err := svc.Find(authedCtx, asset.CollectionFilter{})
		if err != nil || res == nil {
			t.Errorf("expected find success, got err: %v", err)
		}

		svcErr := newTestService(&mockDb{}, &mockAssetRepo{findErr: errors.New("db error")}, &mockPeopleService{}, &mockProjectRepo{}, nil)
		_, err = svcErr.Find(authedCtx, asset.CollectionFilter{})
		if err == nil {
			t.Errorf("expected find error")
		}
	})

	t.Run("FindOne unauthorized and success", func(t *testing.T) {
		repo := &mockAssetRepo{asset: existingAsset}
		svc := newTestService(&mockDb{}, repo, &mockPeopleService{}, &mockProjectRepo{}, nil)

		_, err := svc.FindOne(context.Background(), asset.Filter{})
		if err == nil {
			t.Errorf("expected unauthorized error")
		}

		res, err := svc.FindOne(authedCtx, asset.Filter{})
		if err != nil || res != existingAsset {
			t.Errorf("expected findone success")
		}

		repoErr := &mockAssetRepo{findOneErr: errors.New("not found")}
		svcErr := newTestService(&mockDb{}, repoErr, &mockPeopleService{}, &mockProjectRepo{}, nil)
		_, err = svcErr.FindOne(authedCtx, asset.Filter{})
		if err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("Create success without and with payment", func(t *testing.T) {
		assetRepo := &mockAssetRepo{}
		peopleSvc := &mockPeopleService{owner: owner}
		projRepo := &mockProjectRepo{project: project}

		svc := newTestService(&mockDb{}, assetRepo, peopleSvc, projRepo, nil)

		cmd := asset.CreateAssetCommand{
			Name:        "Server",
			Description: "Rack Server",
			ProjectID:   project.ProjectID,
			Price:       money.New(5000, money.USD),
			WithPayment: false,
		}
		a, err := svc.Create(authedCtx, cmd)
		if err != nil || a == nil {
			t.Fatalf("unexpected error creating asset: %v", err)
		}

		// Create without TypeID (uuid.Nil)
		cmdNoType := cmd
		cmdNoType.TypeID = uuid.Nil
		aNoType, err := svc.Create(authedCtx, cmdNoType)
		if err != nil || aNoType == nil {
			t.Fatalf("unexpected error creating asset without type: %v", err)
		}

		cmdWithPayment := cmd
		cmdWithPayment.WithPayment = true
		cmdWithPayment.Description = ""
		aPay, err := svc.Create(authedCtx, cmdWithPayment)
		if err != nil || aPay == nil {
			t.Fatalf("unexpected error creating asset with payment: %v", err)
		}

		// Create with InvestmentSource
		mockInv := &mockInvestmentSource{}
		svcWithInv := newTestService(&mockDb{}, assetRepo, peopleSvc, projRepo, mockInv)
		aInv, err := svcWithInv.Create(authedCtx, cmdWithPayment)
		if err != nil || aInv == nil {
			t.Fatalf("unexpected error creating asset with investment: %v", err)
		}
		if mockInv.createdItem.AssetID != aInv.ID() {
			t.Fatalf("expected investment to be linked to asset %v, got %v", aInv.ID(), mockInv.createdItem.AssetID)
		}
	})

	t.Run("Create error branches", func(t *testing.T) {
		peopleSvc := &mockPeopleService{owner: owner}
		projRepo := &mockProjectRepo{project: project}
		cmd := asset.CreateAssetCommand{}

		svc := newTestService(&mockDb{}, &mockAssetRepo{}, peopleSvc, projRepo, nil)
		_, err := svc.Create(context.Background(), cmd)
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		svc = newTestService(&mockDb{}, &mockAssetRepo{}, &mockPeopleService{err: errors.New("err")}, projRepo, nil)
		_, err = svc.Create(authedCtx, cmd)
		if err == nil {
			t.Errorf("expected people service error")
		}

		svc = newTestService(&mockDb{}, &mockAssetRepo{}, peopleSvc, &mockProjectRepo{err: errors.New("err")}, nil)
		_, err = svc.Create(authedCtx, cmd)
		if err == nil {
			t.Errorf("expected project repo error")
		}

		// Investment source error
		mockInvErr := &mockInvestmentSource{createErr: errors.New("inv create err")}
		svcInvErr := newTestService(&mockDb{}, &mockAssetRepo{}, peopleSvc, projRepo, mockInvErr)
		_, err = svcInvErr.Create(authedCtx, asset.CreateAssetCommand{WithPayment: true, ProjectID: project.ProjectID})
		if err == nil {
			t.Errorf("expected investment source create error")
		}

		svc = newTestService(&mockDb{}, &mockAssetRepo{saveErr: errors.New("asset save err")}, peopleSvc, projRepo, nil)
		_, err = svc.Create(authedCtx, asset.CreateAssetCommand{WithPayment: true, ProjectID: project.ProjectID})
		if err == nil {
			t.Errorf("expected asset save error")
		}
	})

	t.Run("Remove success and errors", func(t *testing.T) {
		assetRepo := &mockAssetRepo{asset: existingAsset}
		svc := newTestService(&mockDb{}, assetRepo, &mockPeopleService{}, &mockProjectRepo{}, nil)

		err := svc.Remove(context.Background(), asset.RemoveAssetCommand{})
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		svcErr := newTestService(&mockDb{}, &mockAssetRepo{findOneErr: errors.New("not found")}, &mockPeopleService{}, &mockProjectRepo{}, nil)
		err = svcErr.Remove(authedCtx, asset.RemoveAssetCommand{})
		if err == nil {
			t.Errorf("expected findone error")
		}

		err = svc.Remove(authedCtx, asset.RemoveAssetCommand{AssetID: existingAsset.ID()})
		if err != nil {
			t.Errorf("unexpected remove error: %v", err)
		}
	})

	t.Run("Update success and errors", func(t *testing.T) {
		assetRepo := &mockAssetRepo{asset: existingAsset}
		projRepo := &mockProjectRepo{project: project}

		svc := newTestService(&mockDb{}, assetRepo, &mockPeopleService{}, projRepo, nil)

		updCmd := asset.UpdateAssetCommand{
			AssetID:     existingAsset.ID(),
			ProjectID:   project.ProjectID,
			Name:        "Updated Name",
			Description: "Updated Desc",
			Price:       money.New(3000, money.USD),
			AcquiredAt:  time.Now(),
		}

		err := svc.Update(context.Background(), updCmd)
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		svcProjErr := newTestService(&mockDb{}, assetRepo, &mockPeopleService{}, &mockProjectRepo{err: errors.New("proj err")}, nil)
		err = svcProjErr.Update(authedCtx, updCmd)
		if err == nil {
			t.Errorf("expected project error")
		}

		svcAssetErr := newTestService(&mockDb{}, &mockAssetRepo{findOneErr: errors.New("asset err")}, &mockPeopleService{}, projRepo, nil)
		err = svcAssetErr.Update(authedCtx, updCmd)
		if err == nil {
			t.Errorf("expected asset err")
		}

		err = svc.Update(authedCtx, updCmd)
		if err != nil {
			t.Errorf("unexpected update error: %v", err)
		}
	})

	t.Run("CreateFromPayments", func(t *testing.T) {
		inv1ID := uuid.New()
		inv2ID := uuid.New()
		inv3ID := uuid.New()

		d1 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
		d2 := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)

		rateProv := &mockRateProvider{}
		mockInv := &mockInvestmentSource{
			items: []asset.InvestmentItem{
				{ID: inv1ID, Amount: money.New(1000, "USD"), Date: d1, Tags: []string{"tag-a", "tag-b"}},
				{ID: inv2ID, Amount: money.New(2000, "USD"), Date: d2, Tags: []string{"tag-b", "tag-c"}},
				{ID: inv3ID, Amount: money.New(500, "UAH"), Date: d1, Tags: []string{"tag-d"}},
			},
		}

		svc := newTestService(&mockDb{}, &mockAssetRepo{}, &mockPeopleService{owner: owner}, &mockProjectRepo{project: project}, mockInv, rateProv)

		// 1. Unauthorized
		_, err := svc.CreateFromPayments(context.Background(), asset.CreateAssetFromPaymentsCommand{ProjectID: project.ProjectID})
		if err == nil {
			t.Errorf("expected unauthorized error")
		}

		// 2. Missing project id
		_, err = svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{Name: "Asset", PaymentIDs: []uuid.UUID{inv1ID}})
		if err == nil {
			t.Errorf("expected missing project id error")
		}

		// 3. Empty payment ids
		_, err = svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{ProjectID: project.ProjectID, Name: "Asset"})
		if err == nil {
			t.Errorf("expected empty payment ids error")
		}

		// 4. Empty name
		_, err = svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{ProjectID: project.ProjectID, PaymentIDs: []uuid.UUID{inv1ID}})
		if err == nil {
			t.Errorf("expected empty name error")
		}

		// 5. Payment not found
		_, err = svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{
			ProjectID:  project.ProjectID,
			PaymentIDs: []uuid.UUID{uuid.New()},
			Name:       "Asset",
		})
		if err == nil {
			t.Errorf("expected payment not found error")
		}

		// 6. Single payment success (inherits date, amount, currency)
		a, err := svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{
			ProjectID:  project.ProjectID,
			PaymentIDs: []uuid.UUID{inv1ID},
			Name:       "Single Asset",
		})
		if err != nil {
			t.Fatalf("unexpected error creating asset from single payment: %v", err)
		}
		if a.Price().Amount() != 1000 || a.Price().Currency().Code != "USD" {
			t.Errorf("expected price 1000 USD, got %d %s", a.Price().Amount(), a.Price().Currency().Code)
		}
		if !a.AcquiredAt().Equal(d1) {
			t.Errorf("expected date %v, got %v", d1, a.AcquiredAt())
		}

		// 7. Multiple payments same currency (sums amounts, latest date)
		a2, err := svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{
			ProjectID:  project.ProjectID,
			PaymentIDs: []uuid.UUID{inv1ID, inv2ID},
			Name:       "Multi Asset",
		})
		if err != nil {
			t.Fatalf("unexpected error creating asset from multi payments: %v", err)
		}
		if a2.Price().Amount() != 3000 || a2.Price().Currency().Code != "USD" {
			t.Errorf("expected price 3000 USD, got %d %s", a2.Price().Amount(), a2.Price().Currency().Code)
		}
		if !a2.AcquiredAt().Equal(d2) {
			t.Errorf("expected latest date %v, got %v", d2, a2.AcquiredAt())
		}

		// 8. Mixed currencies (USD + UAH, project MainCurrency is UAH)
		// USD 1000 * 40 = 40000 UAH + 500 UAH = 40500 UAH
		a3, err := svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{
			ProjectID:  project.ProjectID,
			PaymentIDs: []uuid.UUID{inv1ID, inv3ID},
			Name:       "Diff Types With Specified Type",
		})
		if err != nil {
			t.Fatalf("unexpected error with mixed currencies: %v", err)
		}
		if a3.Price().Currency().Code != "UAH" || a3.Price().Amount() != 40500 {
			t.Errorf("expected price 40500 UAH, got %d %s", a3.Price().Amount(), a3.Price().Currency().Code)
		}

		// 9. Rate provider error
		svcRateErr := newTestService(&mockDb{}, &mockAssetRepo{}, &mockPeopleService{owner: owner}, &mockProjectRepo{project: project}, mockInv, &mockRateProvider{err: errors.New("rate err")})
		_, err = svcRateErr.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{
			ProjectID:  project.ProjectID,
			PaymentIDs: []uuid.UUID{inv1ID, inv3ID},
			Name:       "Rate Err",
		})
		if err == nil {
			t.Errorf("expected rate provider error")
		}

		// 10. Investment assign error
		mockInvSrcErr := &mockInvestmentSource{
			items: []asset.InvestmentItem{
				{ID: inv1ID, Amount: money.New(100, "UAH"), Date: d1},
			},
			assignErr: errors.New("assign failed"),
		}
		svcAssignErr := newTestService(&mockDb{}, &mockAssetRepo{}, &mockPeopleService{owner: owner}, &mockProjectRepo{project: project}, mockInvSrcErr)
		_, err = svcAssignErr.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{
			ProjectID:  project.ProjectID,
			PaymentIDs: []uuid.UUID{inv1ID},
			Name:       "Fail Assign",
		})
		if err == nil {
			t.Errorf("expected error on assign failure")
		}
	})

	t.Run("LinkChild and UnlinkChild tests", func(t *testing.T) {
		pID := uuid.New()
		cID := uuid.New()
		svc := newTestService(&mockDb{}, &mockAssetRepo{asset: existingAsset}, &mockPeopleService{owner: owner}, &mockProjectRepo{project: project}, nil)

		// Unauthorized
		err := svc.LinkChild(context.Background(), asset.LinkChildCommand{ParentID: pID, ChildID: cID, SharePercentage: 50})
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		// Self link
		err = svc.LinkChild(authedCtx, asset.LinkChildCommand{ParentID: pID, ChildID: pID, SharePercentage: 50})
		if err == nil {
			t.Errorf("expected error for self-link")
		}

		// Invalid share
		err = svc.LinkChild(authedCtx, asset.LinkChildCommand{ParentID: pID, ChildID: cID, SharePercentage: 0})
		if err == nil {
			t.Errorf("expected error for 0 share")
		}
		err = svc.LinkChild(authedCtx, asset.LinkChildCommand{ParentID: pID, ChildID: cID, SharePercentage: 101})
		if err == nil {
			t.Errorf("expected error for >100 share")
		}

		// Cycle detection
		mockSvcCycle := newTestService(&mockDb{}, &mockAssetRepoWithAncestors{mockAssetRepo: mockAssetRepo{asset: existingAsset}, ancestors: []uuid.UUID{cID}}, &mockPeopleService{owner: owner}, &mockProjectRepo{project: project}, nil)
		err = mockSvcCycle.LinkChild(authedCtx, asset.LinkChildCommand{ParentID: pID, ChildID: cID, SharePercentage: 50})
		if err == nil {
			t.Errorf("expected circular dependency error")
		}

		// Success link
		err = svc.LinkChild(authedCtx, asset.LinkChildCommand{ParentID: pID, ChildID: cID, SharePercentage: 50})
		if err != nil {
			t.Errorf("unexpected link error: %v", err)
		}

		// Success unlink
		err = svc.UnlinkChild(authedCtx, asset.UnlinkChildCommand{ParentID: pID, ChildID: cID})
		if err != nil {
			t.Errorf("unexpected unlink error: %v", err)
		}
	})

	t.Run("AssignInvestments", func(t *testing.T) {
		astID := uuid.New()
		invID1 := uuid.New()
		invID2 := uuid.New()
		targetAsset := asset.NewAsset(astID, "Car", "Vehicle", project, nil, money.New(100000, "UAH"), time.Now(), owner)

		mockInv := &mockInvestmentSource{
			items: []asset.InvestmentItem{
				{ID: invID1, Amount: money.New(50000, "UAH"), Date: time.Now()},
				{ID: invID2, Amount: money.New(50000, "UAH"), Date: time.Now()},
			},
		}

		svc := newTestService(
			&mockDb{},
			&mockAssetRepo{asset: targetAsset},
			&mockPeopleService{owner: owner},
			&mockProjectRepo{project: project},
			mockInv,
		)

		// Unauthenticated
		err := svc.AssignInvestments(context.Background(), asset.AssignInvestmentsCommand{
			AssetID:       astID,
			ProjectID:     project.ProjectID,
			InvestmentIDs: []uuid.UUID{invID1, invID2},
		})
		if err == nil {
			t.Errorf("expected unauthenticated error")
		}

		// Validation errors
		err = svc.AssignInvestments(authedCtx, asset.AssignInvestmentsCommand{
			AssetID:       uuid.Nil,
			ProjectID:     project.ProjectID,
			InvestmentIDs: []uuid.UUID{invID1},
		})
		if err == nil {
			t.Errorf("expected error for nil asset id")
		}

		err = svc.AssignInvestments(authedCtx, asset.AssignInvestmentsCommand{
			AssetID:       astID,
			ProjectID:     project.ProjectID,
			InvestmentIDs: nil,
		})
		if err == nil {
			t.Errorf("expected error for empty investment ids")
		}

		// Success
		err = svc.AssignInvestments(authedCtx, asset.AssignInvestmentsCommand{
			AssetID:       astID,
			ProjectID:     project.ProjectID,
			InvestmentIDs: []uuid.UUID{invID1, invID2},
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if mockInv.assignedAsset != astID {
			t.Errorf("expected assigned asset to be %v, got %v", astID, mockInv.assignedAsset)
		}
		if len(mockInv.assignedIDs) != 2 {
			t.Errorf("expected 2 assigned investments, got %d", len(mockInv.assignedIDs))
		}
	})

	t.Run("Group assets success and validation", func(t *testing.T) {
		child1 := asset.NewAsset(uuid.New(), "Child 1", "", project, nil, money.New(100, "UAH"), now, owner)
		child2 := asset.NewAsset(uuid.New(), "Child 2", "", project, nil, money.New(200, "UAH"), now, owner)

		assetRepo := &mockAssetRepo{asset: child1}
		peopleSvc := &mockPeopleService{owner: owner}
		projRepo := &mockProjectRepo{project: project}
		svc := newTestService(&mockDb{}, assetRepo, peopleSvc, projRepo, nil)

		// Validation errors
		_, err := svc.Group(context.Background(), asset.GroupAssetsCommand{})
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		_, err = svc.Group(authedCtx, asset.GroupAssetsCommand{
			ProjectID: uuid.Nil,
		})
		if err == nil {
			t.Errorf("expected validation error for project id")
		}

		_, err = svc.Group(authedCtx, asset.GroupAssetsCommand{
			ProjectID: project.ProjectID,
			Name:      "",
		})
		if err == nil {
			t.Errorf("expected validation error for name")
		}

		_, err = svc.Group(authedCtx, asset.GroupAssetsCommand{
			ProjectID:     project.ProjectID,
			Name:          "Group 1",
			ChildAssetIDs: nil,
		})
		if err == nil {
			t.Errorf("expected validation error for empty child ids")
		}

		// Success
		grouped, err := svc.Group(authedCtx, asset.GroupAssetsCommand{
			ProjectID:     project.ProjectID,
			Name:          "Rent Aggregate",
			Description:   "All rent",
			ChildAssetIDs: []uuid.UUID{child1.ID(), child2.ID()},
			Tags:          []string{"rent"},
		})
		if err != nil || grouped == nil {
			t.Fatalf("unexpected error grouping assets: %v", err)
		}

		// LinkChildren success
		otherChild := asset.NewAsset(uuid.New(), "Other Child", "", project, nil, money.New(50, "UAH"), now, owner)
		err = svc.LinkChildren(authedCtx, asset.LinkChildrenCommand{
			ProjectID:     project.ProjectID,
			ParentID:      grouped.ID(),
			ChildAssetIDs: []uuid.UUID{otherChild.ID()},
		})
		if err != nil {
			t.Errorf("unexpected error linking children: %v", err)
		}
	})
}

type mockAssetRepoWithAncestors struct {
	mockAssetRepo
	ancestors []uuid.UUID
}

func (m *mockAssetRepoWithAncestors) FindAncestors(ctx context.Context, assetID uuid.UUID) ([]uuid.UUID, error) {
	return m.ancestors, nil
}

