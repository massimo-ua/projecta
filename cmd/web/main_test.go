package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/investment"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/project"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
	"gitlab.com/massimo-ua/projecta/pkg/dal"
)

func TestSetupAppHandler(t *testing.T) {
	cfg := &core.AppConfig{
		DbUri:              "postgres://user:pass@localhost:5432/db",
		HttpUri:            ":8080",
		GoogleClientID:     "client_id",
		GoogleClientSecret: "client_secret",
		JwtSecret:          "secret",
		TokenTTL:           300,
		GoogleCertTTL:      3600,
		HttpReadTimeout:    30 * time.Second,
		HttpWriteTimeout:   45 * time.Second,
		ShutdownTimeout:    10 * time.Second,
	}

	db, _ := dal.NewPgDbConnection("postgres://user:pass@localhost:5432/db")

	handler, err := setupAppHandler(cfg, db)
	if err != nil || handler == nil {
		t.Fatalf("unexpected error setting up app handler: %v", err)
	}
}

type mockPersonReader struct {
	person *people.Person
	err    error
}

func (m *mockPersonReader) FindByID(ctx context.Context, personID core.PersonID) (*people.Person, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.person, nil
}

type mockAdapterProjectRepo struct {
	proj *project.Project
	err  error
}

func (m *mockAdapterProjectRepo) Find(ctx context.Context, filter project.ProjectCollectionFilter) ([]*project.Project, error) {
	return nil, nil
}
func (m *mockAdapterProjectRepo) FindOne(ctx context.Context, filter project.ProjectFilter) (*project.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.proj, nil
}
func (m *mockAdapterProjectRepo) FindByShareToken(ctx context.Context, token uuid.UUID) (*project.Project, error) {
	return m.proj, m.err
}
func (m *mockAdapterProjectRepo) Create(ctx context.Context, p *project.Project) error { return nil }
func (m *mockAdapterProjectRepo) Update(ctx context.Context, p *project.Project) error { return nil }
func (m *mockAdapterProjectRepo) Remove(ctx context.Context, p *project.Project) error { return nil }
func (m *mockAdapterProjectRepo) CreateShareRecord(ctx context.Context, projectID core.ProjectID, personID core.PersonID) error {
	return nil
}

type mockAdapterAssetRepo struct {
	ast *asset.Asset
	err error
}

func (m *mockAdapterAssetRepo) Save(ctx context.Context, a *asset.Asset) error   { return nil }
func (m *mockAdapterAssetRepo) Remove(ctx context.Context, a *asset.Asset) error { return nil }
func (m *mockAdapterAssetRepo) FindOne(ctx context.Context, filter asset.Filter) (*asset.Asset, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.ast, nil
}
func (m *mockAdapterAssetRepo) Find(ctx context.Context, filter asset.CollectionFilter) (*asset.Collection, error) {
	return nil, nil
}
func (m *mockAdapterAssetRepo) AddChild(ctx context.Context, parentID core.AssetID, childID core.AssetID, sharePercentage float64) error {
	return nil
}
func (m *mockAdapterAssetRepo) RemoveChild(ctx context.Context, parentID core.AssetID, childID core.AssetID) error {
	return nil
}
func (m *mockAdapterAssetRepo) FindChildren(ctx context.Context, parentID core.AssetID) ([]asset.ChildAssetLink, error) {
	return nil, nil
}
func (m *mockAdapterAssetRepo) FindParents(ctx context.Context, childID core.AssetID) ([]asset.ParentAssetLink, error) {
	return nil, nil
}
func (m *mockAdapterAssetRepo) FindAncestors(ctx context.Context, assetID core.AssetID) ([]core.AssetID, error) {
	return nil, nil
}

func TestCrossContextAdapters(t *testing.T) {
	ctx := context.Background()
	pID := core.NewPersonID()
	cred, _ := people.NewCredentials("LOCAL", "user@example.com", "secret")
	person, _ := people.NewPerson(pID, "John", "Doe", "J.D.", []people.Credentials{cred})

	peopleAdapter := &projectPeopleAdapter{repo: &mockPersonReader{person: person}}
	owner, err := peopleAdapter.FindOwner(ctx, pID)
	if err != nil || owner == nil || owner.PersonID != pID || owner.DisplayName != "J.D." {
		t.Fatalf("unexpected FindOwner result: %+v, err: %v", owner, err)
	}

	peopleAdapterErr := &projectPeopleAdapter{repo: &mockPersonReader{err: errors.New("not found")}}
	if _, err := peopleAdapterErr.FindOwner(ctx, pID); err == nil {
		t.Errorf("expected error from projectPeopleAdapter")
	}

	assetOwnerAdapter := &assetOwnerReaderAdapter{people: peopleAdapter}
	ownerRef, err := assetOwnerAdapter.FindOwner(ctx, pID)
	if err != nil || ownerRef.PersonID != pID || ownerRef.DisplayName != "J.D." {
		t.Fatalf("unexpected assetOwnerReaderAdapter result: %+v, err: %v", ownerRef, err)
	}

	invContributorAdapter := &investmentContributorReaderAdapter{people: peopleAdapter}
	contribRef, err := invContributorAdapter.FindContributor(ctx, pID)
	if err != nil || contribRef.PersonID != pID || contribRef.DisplayName != "J.D." {
		t.Fatalf("unexpected investmentContributorReaderAdapter result: %+v, err: %v", contribRef, err)
	}

	now := time.Now()
	proj, _ := project.NewProject(core.NewProjectID(), "Project 1", "Desc", owner, now, now, "USD")
	projRepo := &mockAdapterProjectRepo{proj: proj}

	assetProjAdapter := &assetProjectReaderAdapter{projects: projRepo}
	assetProjDesc, err := assetProjAdapter.FindProject(ctx, proj.ID())
	if err != nil || assetProjDesc.ID != proj.ID() || assetProjDesc.MainCurrency != "USD" {
		t.Fatalf("unexpected assetProjectReaderAdapter result: %+v, err: %v", assetProjDesc, err)
	}

	invProjAdapter := &investmentProjectReaderAdapter{projects: projRepo}
	invProjDesc, err := invProjAdapter.FindProject(ctx, proj.ID())
	if err != nil || invProjDesc.ID != proj.ID() || invProjDesc.MainCurrency != "USD" {
		t.Fatalf("unexpected investmentProjectReaderAdapter result: %+v, err: %v", invProjDesc, err)
	}

	ast, _ := asset.CreateAsset(core.NewAssetID(), "Truck", "Desc", proj.ID(), money.New(1000, "USD"), now, ownerRef, asset.AssetStatusActive, nil, nil)
	invAssetAdapter := &investmentAssetReaderAdapter{assets: &mockAdapterAssetRepo{ast: ast}}
	invAssetDesc, err := invAssetAdapter.FindAsset(ctx, proj.ID(), ast.ID())
	if err != nil || invAssetDesc.ID != ast.ID() || invAssetDesc.Name != "Truck" {
		t.Fatalf("unexpected investmentAssetReaderAdapter result: %+v, err: %v", invAssetDesc, err)
	}

	// currencyConverterAdapter
	convNilAdapter := &currencyConverterAdapter{provider: nil}
	nilMoneyRes, err := convNilAdapter.ConvertMoney(nil, "USD")
	if err != nil || nilMoneyRes.Amount() != 0 || nilMoneyRes.Currency().Code != "USD" {
		t.Fatalf("unexpected nil money conversion: %+v, %v", nilMoneyRes, err)
	}

	sameCurrRes, err := convNilAdapter.ConvertMoney(money.New(500, "USD"), "USD")
	if err != nil || sameCurrRes.Amount() != 500 {
		t.Fatalf("unexpected same currency conversion: %+v, %v", sameCurrRes, err)
	}

	if _, err := convNilAdapter.ConvertMoney(money.New(500, "USD"), "UAH"); err == nil {
		t.Errorf("expected error when provider is nil")
	}

	mockRateProv := &mockRateProvider{}
	convAdapter := &currencyConverterAdapter{provider: mockRateProv}
	convRes, err := convAdapter.ConvertMoney(money.New(100, "USD"), "UAH")
	if err != nil || convRes.Amount() != 4000 || convRes.Currency().Code != "UAH" {
		t.Fatalf("unexpected conversion result: %+v, %v", convRes, err)
	}

	mockRateProvErr := &mockRateProvider{err: errors.New("rate failed")}
	convAdapterErr := &currencyConverterAdapter{provider: mockRateProvErr}
	if _, err := convAdapterErr.ConvertMoney(money.New(100, "USD"), "UAH"); err == nil {
		t.Errorf("expected error when provider fails")
	}
}

type mockRateProvider struct {
	err error
}

func (m *mockRateProvider) Convert(a currency.Currency, b currency.Currency) (currency.Currency, error) {
	if m.err != nil {
		return currency.Currency{}, m.err
	}
	return currency.Currency{Amount: a.Amount * 40, Code: b.Code}, nil
}

type mockInvestmentService struct {
	findFn func(ctx context.Context, filter investment.CollectionFilter) (*investment.Collection, error)
	col    *investment.Collection
	inv    *investment.Investment
	err    error
}

func (m *mockInvestmentService) Create(context.Context, investment.CreateInvestmentCommand) (*investment.Investment, error) {
	return m.inv, m.err
}
func (m *mockInvestmentService) CreateBatch(context.Context, investment.CreateBatchInvestmentsCommand) ([]*investment.Investment, error) {
	return nil, m.err
}
func (m *mockInvestmentService) Update(context.Context, investment.UpdateInvestmentCommand) error {
	return m.err
}
func (m *mockInvestmentService) Remove(context.Context, investment.RemoveInvestmentCommand) error {
	return m.err
}
func (m *mockInvestmentService) FindOne(context.Context, investment.Filter) (*investment.Investment, error) {
	return m.inv, m.err
}
func (m *mockInvestmentService) Find(ctx context.Context, filter investment.CollectionFilter) (*investment.Collection, error) {
	if m.findFn != nil {
		return m.findFn(ctx, filter)
	}
	if m.err != nil {
		return nil, m.err
	}
	if m.col != nil {
		return m.col, nil
	}
	return investment.NewCollection(0), nil
}
func (m *mockInvestmentService) FindTags(context.Context, core.ProjectID) ([]string, error) {
	return nil, m.err
}
func (m *mockInvestmentService) AssignToAsset(context.Context, core.AssetID, []core.InvestmentID) error {
	return m.err
}

type mockAssetService struct {
	findFn func(ctx context.Context, filter asset.CollectionFilter) (*asset.Collection, error)
	col    *asset.Collection
	ast    *asset.Asset
	err    error
}

func (m *mockAssetService) Create(context.Context, asset.CreateAssetCommand) (*asset.Asset, error) {
	return m.ast, m.err
}
func (m *mockAssetService) CreateFromPayments(context.Context, asset.CreateAssetFromPaymentsCommand) (*asset.Asset, error) {
	return m.ast, m.err
}
func (m *mockAssetService) CreateFromInvestments(context.Context, asset.CreateAssetFromInvestmentsCommand) (*asset.Asset, error) {
	return m.ast, m.err
}
func (m *mockAssetService) Update(context.Context, asset.UpdateAssetCommand) error {
	return m.err
}
func (m *mockAssetService) Remove(context.Context, asset.RemoveAssetCommand) error {
	return m.err
}
func (m *mockAssetService) FindOne(context.Context, asset.Filter) (*asset.Asset, error) {
	return m.ast, m.err
}
func (m *mockAssetService) Find(ctx context.Context, filter asset.CollectionFilter) (*asset.Collection, error) {
	if m.findFn != nil {
		return m.findFn(ctx, filter)
	}
	if m.err != nil {
		return nil, m.err
	}
	if m.col != nil {
		return m.col, nil
	}
	return asset.NewCollection(0), nil
}
func (m *mockAssetService) LinkChild(context.Context, asset.LinkChildCommand) error {
	return m.err
}
func (m *mockAssetService) UnlinkChild(context.Context, asset.UnlinkChildCommand) error {
	return m.err
}
func (m *mockAssetService) AssignInvestments(context.Context, asset.AssignInvestmentsCommand) error {
	return m.err
}
func (m *mockAssetService) Group(context.Context, asset.GroupAssetsCommand) (*asset.Asset, error) {
	return m.ast, m.err
}
func (m *mockAssetService) LinkChildren(context.Context, asset.LinkChildrenCommand) error {
	return m.err
}

type mockProjectServiceForAdapter struct {
	proj *project.Project
	err  error
}

func (m *mockProjectServiceForAdapter) Find(context.Context, project.ProjectCollectionFilter) ([]*project.Project, error) {
	return nil, nil
}
func (m *mockProjectServiceForAdapter) FindOne(context.Context, project.ProjectFilter) (*project.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.proj, nil
}
func (m *mockProjectServiceForAdapter) Create(context.Context, project.CreateProjectCommand) (*project.Project, error) {
	return m.proj, m.err
}
func (m *mockProjectServiceForAdapter) Remove(context.Context, project.RemoveProjectCommand) error {
	return m.err
}
func (m *mockProjectServiceForAdapter) Update(context.Context, project.UpdateProjectCommand) (*project.Project, error) {
	return m.proj, m.err
}
func (m *mockProjectServiceForAdapter) AcceptShare(context.Context, uuid.UUID, core.PersonID) (*project.Project, error) {
	return m.proj, m.err
}
func (m *mockProjectServiceForAdapter) GetTotals(context.Context, core.ProjectID) (project.ProjectTotals, error) {
	return project.ProjectTotals{}, m.err
}

func TestAdapters_Step7(t *testing.T) {
	ctx := context.Background()
	projID := core.NewProjectID()
	contribRef := investment.ContributorRef{PersonID: core.NewPersonID(), DisplayName: "Contributor"}

	t.Run("investmentTransactionFinder pagination", func(t *testing.T) {
		inv := investment.ReconstituteInvestment(
			core.NewInvestmentID(), projID, "UAH", core.NewAssetID(), "Asset", nil,
			contribRef, investment.ResourceTypeMoney, money.New(100, "UAH"),
			nil, nil, nil, "", "", "desc", time.Now(), nil,
		)

		invSvc := &mockInvestmentService{
			findFn: func(ctx context.Context, filter investment.CollectionFilter) (*investment.Collection, error) {
				if filter.Offset == 0 {
					col := investment.NewCollection(600)
					for i := 0; i < 500; i++ {
						col.Add(inv)
					}
					return col, nil
				}
				col := investment.NewCollection(600)
				for i := 0; i < 100; i++ {
					col.Add(inv)
				}
				return col, nil
			},
		}

		finder := &investmentTransactionFinder{svc: invSvc}
		txs, err := finder.FindExisting(ctx, projID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(txs) != 600 {
			t.Fatalf("expected 600 transactions from pagination, got %d", len(txs))
		}
	})

	t.Run("investmentSourceAdapter FindInvestments batching", func(t *testing.T) {
		id1 := core.NewInvestmentID()
		id2 := core.NewInvestmentID()
		inv1 := investment.ReconstituteInvestment(
			id1, projID, "UAH", core.NewAssetID(), "Asset 1", nil,
			contribRef, investment.ResourceTypeMoney, money.New(200, "UAH"),
			nil, nil, nil, "", "", "desc1", time.Now(), nil,
		)
		inv2 := investment.ReconstituteInvestment(
			id2, projID, "UAH", core.NewAssetID(), "Asset 2", nil,
			contribRef, investment.ResourceTypeMoney, money.New(300, "UAH"),
			nil, nil, nil, "", "", "desc2", time.Now(), nil,
		)
		col := investment.NewCollection(2)
		col.Add(inv1, inv2)

		var capturedFilter investment.CollectionFilter
		invSvc := &mockInvestmentService{
			findFn: func(ctx context.Context, filter investment.CollectionFilter) (*investment.Collection, error) {
				capturedFilter = filter
				return col, nil
			},
		}

		adapter := &investmentSourceAdapter{svc: invSvc}
		items, err := adapter.FindInvestments(ctx, projID, []core.InvestmentID{id1, id2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(items))
		}
		if len(capturedFilter.IDs) != 2 || capturedFilter.Limit != 2 {
			t.Errorf("expected IDs filter and Limit 2, got %+v", capturedFilter)
		}

		// Empty IDs
		emptyItems, err := adapter.FindInvestments(ctx, projID, nil)
		if err != nil || emptyItems != nil {
			t.Fatalf("expected nil for empty ids, got: %v, %v", emptyItems, err)
		}

		// AssignToAsset
		if err := adapter.AssignToAsset(ctx, core.NewAssetID(), []core.InvestmentID{id1}); err != nil {
			t.Errorf("unexpected error in AssignToAsset: %v", err)
		}

		// CreateInitialInvestment
		if err := adapter.CreateInitialInvestment(ctx, asset.InitialInvestment{
			ProjectID: projID,
			AssetID:   core.NewAssetID(),
			Amount:    money.New(500, "UAH"),
		}); err != nil {
			t.Errorf("unexpected error in CreateInitialInvestment: %v", err)
		}
	})

	t.Run("projectTotalsCalculatorAdapter", func(t *testing.T) {
		inv := investment.ReconstituteInvestment(
			core.NewInvestmentID(), projID, "UAH", core.NewAssetID(), "Asset", nil,
			contribRef, investment.ResourceTypeMoney, money.New(50000, "UAH"),
			nil, nil, nil, "", "", "desc", time.Now(), nil,
		)
		invCol := investment.NewCollection(1)
		invCol.Add(inv)
		invSvc := &mockInvestmentService{col: invCol}

		ownerRef := asset.OwnerRef{PersonID: core.NewPersonID(), DisplayName: "Owner"}
		ast, _ := asset.CreateAsset(
			core.NewAssetID(), "Root Asset", "desc", projID, money.New(20000, "UAH"),
			time.Now(), ownerRef, asset.AssetStatusActive, nil, nil,
		)
		astCol := asset.NewCollection(1)
		astCol.Add(ast)
		astSvc := &mockAssetService{col: astCol}

		calc := &projectTotalsCalculatorAdapter{
			assets:       astSvc,
			investments:  invSvc,
			rateProvider: &mockRateProvider{},
		}

		totals, err := calc.CalculateTotals(ctx, projID, "UAH")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !totals.HasActivity {
			t.Errorf("expected HasActivity to be true")
		}
		if totals.TotalInvested != 50000 {
			t.Errorf("expected TotalInvested 50000, got %d", totals.TotalInvested)
		}
		if totals.Balance != 30000 {
			t.Errorf("expected Balance 30000, got %d", totals.Balance)
		}

		// Error in investments
		calcErr := &projectTotalsCalculatorAdapter{
			assets:      astSvc,
			investments: &mockInvestmentService{err: errors.New("inv error")},
		}
		if _, err := calcErr.CalculateTotals(ctx, projID, "UAH"); err == nil {
			t.Error("expected investment error")
		}

		// Error in assets
		calcAstErr := &projectTotalsCalculatorAdapter{
			assets:      &mockAssetService{err: errors.New("ast error")},
			investments: invSvc,
		}
		if _, err := calcAstErr.CalculateTotals(ctx, projID, "UAH"); err == nil {
			t.Error("expected asset error")
		}
	})

	t.Run("projectValidator", func(t *testing.T) {
		pvSuccess := &projectValidator{svc: &mockProjectServiceForAdapter{}}
		if err := pvSuccess.ValidateProject(ctx, projID); err != nil {
			t.Errorf("expected nil error, got: %v", err)
		}

		pvErr := &projectValidator{svc: &mockProjectServiceForAdapter{err: errors.New("not found")}}
		if err := pvErr.ValidateProject(ctx, projID); err == nil {
			t.Error("expected error from projectValidator")
		}
	})
}


