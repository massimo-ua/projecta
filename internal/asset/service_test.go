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

type mockTypeRepo struct {
	costType *projecta.CostType
	err      error
}

func (m *mockTypeRepo) Save(ctx context.Context, t *projecta.CostType) error   { return nil }
func (m *mockTypeRepo) Remove(ctx context.Context, t *projecta.CostType) error { return nil }
func (m *mockTypeRepo) Find(ctx context.Context, filter projecta.TypeCollectionFilter) (*projecta.CostTypeCollection, error) {
	return nil, nil
}
func (m *mockTypeRepo) FindOne(ctx context.Context, filter projecta.TypeFilter) (*projecta.CostType, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.costType, nil
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

type mockPaymentRepo struct {
	saveErr  error
	payments map[uuid.UUID]*projecta.Payment
	err      error
}

func (m *mockPaymentRepo) Save(ctx context.Context, p *projecta.Payment) error { return m.saveErr }
func (m *mockPaymentRepo) SaveBatch(ctx context.Context, payments []*projecta.Payment) error {
	return m.saveErr
}
func (m *mockPaymentRepo) Remove(ctx context.Context, p *projecta.Payment) error {
	return nil
}
func (m *mockPaymentRepo) Find(ctx context.Context, filter projecta.PaymentCollectionFilter) (*projecta.PaymentCollection, error) {
	return nil, nil
}
func (m *mockPaymentRepo) FindOne(ctx context.Context, filter projecta.PaymentFilter) (*projecta.Payment, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.payments != nil {
		if p, ok := m.payments[filter.PaymentID]; ok {
			return p, nil
		}
		return nil, errors.New("not found")
	}
	return nil, nil
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

func TestAssetService(t *testing.T) {
	requesterID := uuid.New()
	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, requesterID)

	now := time.Now()
	owner := &projecta.Owner{PersonID: requesterID, DisplayName: "John Doe"}
	project, _ := projecta.NewProject(uuid.New(), "Project 1", "Desc", owner, now, now)
	costType, _ := projecta.NewCostType(project.ProjectID, nil, "Type 1", "Desc")
	existingAsset := asset.NewAsset(uuid.New(), "Laptop", "Work Laptop", project, costType, money.New(1000, money.USD), now, owner)

	t.Run("Find unauthorized and success", func(t *testing.T) {
		svc := asset.NewService(&mockDb{}, &mockAssetRepo{}, &mockPeopleService{}, &mockTypeRepo{}, &mockProjectRepo{}, &mockPaymentRepo{})

		_, err := svc.Find(context.Background(), asset.CollectionFilter{})
		if err == nil {
			t.Errorf("expected unauthorized error")
		}

		res, err := svc.Find(authedCtx, asset.CollectionFilter{})
		if err != nil || res == nil {
			t.Errorf("expected find success, got err: %v", err)
		}

		svcErr := asset.NewService(&mockDb{}, &mockAssetRepo{findErr: errors.New("db error")}, &mockPeopleService{}, &mockTypeRepo{}, &mockProjectRepo{}, &mockPaymentRepo{})
		_, err = svcErr.Find(authedCtx, asset.CollectionFilter{})
		if err == nil {
			t.Errorf("expected find error")
		}
	})

	t.Run("FindOne unauthorized and success", func(t *testing.T) {
		repo := &mockAssetRepo{asset: existingAsset}
		svc := asset.NewService(&mockDb{}, repo, &mockPeopleService{}, &mockTypeRepo{}, &mockProjectRepo{}, &mockPaymentRepo{})

		_, err := svc.FindOne(context.Background(), asset.Filter{})
		if err == nil {
			t.Errorf("expected unauthorized error")
		}

		res, err := svc.FindOne(authedCtx, asset.Filter{})
		if err != nil || res != existingAsset {
			t.Errorf("expected findone success")
		}

		repoErr := &mockAssetRepo{findOneErr: errors.New("not found")}
		svcErr := asset.NewService(&mockDb{}, repoErr, &mockPeopleService{}, &mockTypeRepo{}, &mockProjectRepo{}, &mockPaymentRepo{})
		_, err = svcErr.FindOne(authedCtx, asset.Filter{})
		if err == nil {
			t.Errorf("expected error")
		}
	})

	t.Run("Create success without and with payment", func(t *testing.T) {
		assetRepo := &mockAssetRepo{}
		peopleSvc := &mockPeopleService{owner: owner}
		typeRepo := &mockTypeRepo{costType: costType}
		projRepo := &mockProjectRepo{project: project}
		payRepo := &mockPaymentRepo{}

		svc := asset.NewService(&mockDb{}, assetRepo, peopleSvc, typeRepo, projRepo, payRepo)

		cmd := asset.CreateAssetCommand{
			Name:        "Server",
			Description: "Rack Server",
			ProjectID:   project.ProjectID,
			TypeID:      costType.ID,
			Price:       money.New(5000, money.USD),
			WithPayment: false,
		}
		a, err := svc.Create(authedCtx, cmd)
		if err != nil || a == nil {
			t.Fatalf("unexpected error creating asset: %v", err)
		}

		cmdWithPayment := cmd
		cmdWithPayment.WithPayment = true
		cmdWithPayment.Description = ""
		aPay, err := svc.Create(authedCtx, cmdWithPayment)
		if err != nil || aPay == nil {
			t.Fatalf("unexpected error creating asset with payment: %v", err)
		}
	})

	t.Run("Create error branches", func(t *testing.T) {
		peopleSvc := &mockPeopleService{owner: owner}
		typeRepo := &mockTypeRepo{costType: costType}
		projRepo := &mockProjectRepo{project: project}
		cmd := asset.CreateAssetCommand{}

		svc := asset.NewService(&mockDb{}, &mockAssetRepo{}, peopleSvc, typeRepo, projRepo, &mockPaymentRepo{})
		_, err := svc.Create(context.Background(), cmd)
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		svc = asset.NewService(&mockDb{}, &mockAssetRepo{}, &mockPeopleService{err: errors.New("err")}, typeRepo, projRepo, &mockPaymentRepo{})
		_, err = svc.Create(authedCtx, cmd)
		if err == nil {
			t.Errorf("expected people service error")
		}

		svc = asset.NewService(&mockDb{}, &mockAssetRepo{}, peopleSvc, typeRepo, &mockProjectRepo{err: errors.New("err")}, &mockPaymentRepo{})
		_, err = svc.Create(authedCtx, cmd)
		if err == nil {
			t.Errorf("expected project repo error")
		}

		svc = asset.NewService(&mockDb{}, &mockAssetRepo{}, peopleSvc, &mockTypeRepo{err: errors.New("err")}, projRepo, &mockPaymentRepo{})
		_, err = svc.Create(authedCtx, cmd)
		if err == nil {
			t.Errorf("expected type repo error")
		}

		svc = asset.NewService(&mockDb{}, &mockAssetRepo{}, peopleSvc, typeRepo, projRepo, &mockPaymentRepo{saveErr: errors.New("pay save err")})
		_, err = svc.Create(authedCtx, asset.CreateAssetCommand{WithPayment: true, ProjectID: project.ProjectID, TypeID: costType.ID})
		if err == nil {
			t.Errorf("expected payment save error")
		}

		svc = asset.NewService(&mockDb{}, &mockAssetRepo{saveErr: errors.New("asset save err")}, peopleSvc, typeRepo, projRepo, &mockPaymentRepo{})
		_, err = svc.Create(authedCtx, asset.CreateAssetCommand{WithPayment: true, ProjectID: project.ProjectID, TypeID: costType.ID})
		if err == nil {
			t.Errorf("expected asset save error")
		}
	})

	t.Run("Remove success and errors", func(t *testing.T) {
		assetRepo := &mockAssetRepo{asset: existingAsset}
		svc := asset.NewService(&mockDb{}, assetRepo, &mockPeopleService{}, &mockTypeRepo{}, &mockProjectRepo{}, &mockPaymentRepo{})

		err := svc.Remove(context.Background(), asset.RemoveAssetCommand{})
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		svcErr := asset.NewService(&mockDb{}, &mockAssetRepo{findOneErr: errors.New("not found")}, &mockPeopleService{}, &mockTypeRepo{}, &mockProjectRepo{}, &mockPaymentRepo{})
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
		typeRepo := &mockTypeRepo{costType: costType}
		projRepo := &mockProjectRepo{project: project}

		svc := asset.NewService(&mockDb{}, assetRepo, &mockPeopleService{}, typeRepo, projRepo, &mockPaymentRepo{})

		updCmd := asset.UpdateAssetCommand{
			AssetID:     existingAsset.ID(),
			ProjectID:   project.ProjectID,
			TypeID:      costType.ID,
			Name:        "Updated Name",
			Description: "Updated Desc",
			Price:       money.New(3000, money.USD),
			AcquiredAt:  time.Now(),
		}

		err := svc.Update(context.Background(), updCmd)
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		svcProjErr := asset.NewService(&mockDb{}, assetRepo, &mockPeopleService{}, typeRepo, &mockProjectRepo{err: errors.New("proj err")}, &mockPaymentRepo{})
		err = svcProjErr.Update(authedCtx, updCmd)
		if err == nil {
			t.Errorf("expected project error")
		}

		svcAssetErr := asset.NewService(&mockDb{}, &mockAssetRepo{findOneErr: errors.New("asset err")}, &mockPeopleService{}, typeRepo, projRepo, &mockPaymentRepo{})
		err = svcAssetErr.Update(authedCtx, updCmd)
		if err == nil {
			t.Errorf("expected asset err")
		}

		svcTypeErr := asset.NewService(&mockDb{}, assetRepo, &mockPeopleService{}, &mockTypeRepo{err: errors.New("type err")}, projRepo, &mockPaymentRepo{})
		err = svcTypeErr.Update(authedCtx, updCmd)
		if err == nil {
			t.Errorf("expected type err")
		}

		err = svc.Update(authedCtx, updCmd)
		if err != nil {
			t.Errorf("unexpected update error: %v", err)
		}
	})

	t.Run("CreateFromPayments", func(t *testing.T) {
		p1ID := uuid.New()
		p2ID := uuid.New()
		p3ID := uuid.New()

		d1 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
		d2 := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)

		costType2, _ := projecta.NewCostType(project.ProjectID, nil, "Type 2", "Desc")

		pay1 := projecta.NewPayment(p1ID, project, owner, costType, "Pay 1", money.New(1000, "USD"), d1, projecta.UponCompletionPayment)
		pay2 := projecta.NewPayment(p2ID, project, owner, costType, "Pay 2", money.New(2000, "USD"), d2, projecta.UponCompletionPayment)
		pay3 := projecta.NewPayment(p3ID, project, owner, costType2, "Pay 3", money.New(500, "UAH"), d1, projecta.UponCompletionPayment)

		paymentsMap := map[uuid.UUID]*projecta.Payment{
			p1ID: pay1,
			p2ID: pay2,
			p3ID: pay3,
		}

		rateProv := &mockRateProvider{}
		payRepoWithPayments := &mockPaymentRepo{payments: paymentsMap}
		typeRepoMulti := &mockTypeRepo{costType: costType}

		svc := asset.NewService(&mockDb{}, &mockAssetRepo{}, &mockPeopleService{owner: owner}, typeRepoMulti, &mockProjectRepo{project: project}, payRepoWithPayments, rateProv)

		// 1. Unauthorized
		_, err := svc.CreateFromPayments(context.Background(), asset.CreateAssetFromPaymentsCommand{ProjectID: project.ProjectID})
		if err == nil {
			t.Errorf("expected unauthorized error")
		}

		// 2. Missing project id
		_, err = svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{Name: "Asset", PaymentIDs: []uuid.UUID{p1ID}})
		if err == nil {
			t.Errorf("expected missing project id error")
		}

		// 3. Empty payment ids
		_, err = svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{ProjectID: project.ProjectID, Name: "Asset"})
		if err == nil {
			t.Errorf("expected empty payment ids error")
		}

		// 4. Empty name
		_, err = svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{ProjectID: project.ProjectID, PaymentIDs: []uuid.UUID{p1ID}})
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

		// 6. Single payment success (inherits type, date, amount, currency)
		a, err := svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{
			ProjectID:  project.ProjectID,
			PaymentIDs: []uuid.UUID{p1ID},
			Name:       "Single Asset",
		})
		if err != nil {
			t.Fatalf("unexpected error creating asset from single payment: %v", err)
		}
		if a.Price().Amount() != 1000 || a.Price().Currency().Code != "USD" {
			t.Errorf("expected price 1000 USD, got %d %s", a.Price().Amount(), a.Price().Currency().Code)
		}
		if a.Type().ID != costType.ID {
			t.Errorf("expected type %s, got %s", costType.ID, a.Type().ID)
		}
		if !a.AcquiredAt().Equal(d1) {
			t.Errorf("expected date %v, got %v", d1, a.AcquiredAt())
		}

		// 7. Multiple payments same currency (sums amounts, latest date)
		a2, err := svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{
			ProjectID:  project.ProjectID,
			PaymentIDs: []uuid.UUID{p1ID, p2ID},
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

		// 8. Differing types without type_id -> error
		_, err = svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{
			ProjectID:  project.ProjectID,
			PaymentIDs: []uuid.UUID{p1ID, p3ID},
			Name:       "Diff Types",
		})
		if err == nil {
			t.Errorf("expected error when payments have different types without type_id")
		}

		// 9. Differing types with type_id -> success
		a3, err := svc.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{
			ProjectID:  project.ProjectID,
			PaymentIDs: []uuid.UUID{p1ID, p3ID},
			TypeID:     costType.ID,
			Name:       "Diff Types With Specified Type",
		})
		if err != nil {
			t.Fatalf("unexpected error with specified type_id: %v", err)
		}
		if a3.Type().ID != costType.ID {
			t.Errorf("expected specified type_id")
		}

		// 10. Mixed currencies (USD + UAH, project MainCurrency is UAH)
		// USD 1000 * 40 = 40000 UAH + 500 UAH = 40500 UAH
		if a3.Price().Currency().Code != "UAH" || a3.Price().Amount() != 40500 {
			t.Errorf("expected price 40500 UAH, got %d %s", a3.Price().Amount(), a3.Price().Currency().Code)
		}

		// 11. Rate provider error
		svcRateErr := asset.NewService(&mockDb{}, &mockAssetRepo{}, &mockPeopleService{owner: owner}, typeRepoMulti, &mockProjectRepo{project: project}, payRepoWithPayments, &mockRateProvider{err: errors.New("rate err")})
		_, err = svcRateErr.CreateFromPayments(authedCtx, asset.CreateAssetFromPaymentsCommand{
			ProjectID:  project.ProjectID,
			PaymentIDs: []uuid.UUID{p1ID, p3ID},
			TypeID:     costType.ID,
			Name:       "Rate Err",
		})
		if err == nil {
			t.Errorf("expected rate provider error")
		}
	})

	t.Run("LinkChild and UnlinkChild tests", func(t *testing.T) {
		pID := uuid.New()
		cID := uuid.New()
		svc := asset.NewService(&mockDb{}, &mockAssetRepo{asset: existingAsset}, &mockPeopleService{owner: owner}, &mockTypeRepo{}, &mockProjectRepo{project: project}, &mockPaymentRepo{})

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
		mockSvcCycle := asset.NewService(&mockDb{}, &mockAssetRepoWithAncestors{mockAssetRepo: mockAssetRepo{asset: existingAsset}, ancestors: []uuid.UUID{cID}}, &mockPeopleService{owner: owner}, &mockTypeRepo{}, &mockProjectRepo{project: project}, &mockPaymentRepo{})
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
}

type mockAssetRepoWithAncestors struct {
	mockAssetRepo
	ancestors []uuid.UUID
}

func (m *mockAssetRepoWithAncestors) FindAncestors(ctx context.Context, assetID uuid.UUID) ([]uuid.UUID, error) {
	return m.ancestors, nil
}

