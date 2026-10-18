package investment_test

import (
	"context"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/investment"
)

type mockDb struct {
	txCalled bool
}

func (m *mockDb) Tx(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	m.txCalled = true
	return fn(ctx)
}
func (m *mockDb) Close()                        {}
func (m *mockDb) Ping(ctx context.Context) error { return nil }

type mockInvestmentRepo struct {
	saveErr    error
	removeErr  error
	findErr    error
	findOneErr error
	inv        *investment.Investment
	col        *investment.Collection
	tags       []string
}

func (m *mockInvestmentRepo) Save(ctx context.Context, inv *investment.Investment) error {
	return m.saveErr
}
func (m *mockInvestmentRepo) Remove(ctx context.Context, inv *investment.Investment) error {
	return m.removeErr
}
func (m *mockInvestmentRepo) FindOne(ctx context.Context, filter investment.Filter) (*investment.Investment, error) {
	if m.findOneErr != nil {
		return nil, m.findOneErr
	}
	return m.inv, nil
}
func (m *mockInvestmentRepo) Find(ctx context.Context, filter investment.CollectionFilter) (*investment.Collection, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.col != nil {
		return m.col, nil
	}
	return investment.NewCollection(0), nil
}
func (m *mockInvestmentRepo) FindTags(ctx context.Context, projectID core.ProjectID) ([]string, error) {
	return m.tags, nil
}

type mockAssetRepo struct {
	asset  investment.AssetDescriptor
	assets map[core.AssetID]investment.AssetDescriptor
	err    error
}

func (m *mockAssetRepo) FindAsset(ctx context.Context, projectID core.ProjectID, assetID core.AssetID) (investment.AssetDescriptor, error) {
	if m.err != nil {
		return investment.AssetDescriptor{}, m.err
	}
	if m.assets != nil {
		if a, ok := m.assets[assetID]; ok {
			return a, nil
		}
	}
	return m.asset, nil
}

type mockProjectRepo struct {
	project investment.ProjectDescriptor
	err     error
}

func (m *mockProjectRepo) FindProject(ctx context.Context, projectID core.ProjectID) (investment.ProjectDescriptor, error) {
	if m.err != nil {
		return investment.ProjectDescriptor{}, m.err
	}
	return m.project, nil
}

type mockPeopleService struct {
	owner investment.ContributorRef
	err   error
}

func (m *mockPeopleService) FindContributor(ctx context.Context, id core.PersonID) (investment.ContributorRef, error) {
	if m.err != nil {
		return investment.ContributorRef{}, m.err
	}
	return m.owner, nil
}

func TestInvestmentService(t *testing.T) {
	personID := core.NewPersonID()
	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, personID)

	owner := investment.ContributorRef{PersonID: personID, DisplayName: "Alex"}
	project := investment.ProjectDescriptor{ID: core.NewProjectID(), MainCurrency: "UAH"}
	existingAsset := investment.AssetDescriptor{ID: core.NewAssetID(), Name: "Vehicle"}

	t.Run("Create success (Money, Time, Goods)", func(t *testing.T) {
		db := &mockDb{}
		repo := &mockInvestmentRepo{}
		svc := investment.NewService(
			db,
			repo,
			&mockAssetRepo{asset: existingAsset},
			&mockProjectRepo{project: project},
			&mockPeopleService{owner: owner},
		)

		// Money investment
		inv, err := svc.Create(authedCtx, investment.CreateInvestmentCommand{
			ProjectID:    project.ID,
			AssetID:      existingAsset.ID,
			ResourceType: investment.ResourceTypeMoney,
			Amount:       money.New(500, "USD"),
			Description:  "Fuel purchase",
			Tags:         []string{"fuel", "travel"},
		})
		if err != nil {
			t.Fatalf("unexpected error creating money investment: %v", err)
		}
		if inv.Amount().Amount() != 500 || inv.ResourceType() != investment.ResourceTypeMoney {
			t.Errorf("unexpected money investment details: %+v", inv)
		}
		if !db.txCalled {
			t.Errorf("expected db.Tx to be called on Create")
		}

		// Time investment with calculated rate
		hours := 4.5
		rate := money.New(4000, "USD") // $40/hr
		invTime, err := svc.Create(authedCtx, investment.CreateInvestmentCommand{
			ProjectID:      project.ID,
			AssetID:        existingAsset.ID,
			ResourceType:   investment.ResourceTypeTime,
			TimeHours:      &hours,
			TimeHourlyRate: rate,
			Description:    "Brake repair",
			Tags:           []string{"labor", "maintenance"},
		})
		if err != nil {
			t.Fatalf("unexpected error creating time investment: %v", err)
		}
		if invTime.Amount().Amount() != 18000 { // 4.5 * 4000 = 18000
			t.Errorf("expected 18000 amount for time, got %d", invTime.Amount().Amount())
		}

		// Goods investment
		qty := 4.0
		invGoods, err := svc.Create(authedCtx, investment.CreateInvestmentCommand{
			ProjectID:     project.ID,
			AssetID:       existingAsset.ID,
			ResourceType:  investment.ResourceTypeGoods,
			Amount:        money.New(30000, "USD"),
			GoodsQuantity: &qty,
			GoodsUnit:     "pcs",
			GoodsItemName: "Winter Tires",
			Description:   "Installed tires",
			Tags:          []string{"parts"},
		})
		if err != nil {
			t.Fatalf("unexpected error creating goods investment: %v", err)
		}
		if invGoods.GoodsItemName() != "Winter Tires" {
			t.Errorf("expected GoodsItemName 'Winter Tires', got %s", invGoods.GoodsItemName())
		}
	})

	t.Run("Create validation errors", func(t *testing.T) {
		svc := investment.NewService(
			&mockDb{},
			&mockInvestmentRepo{},
			&mockAssetRepo{asset: existingAsset},
			&mockProjectRepo{project: project},
			&mockPeopleService{owner: owner},
		)

		// Unauthorized
		_, err := svc.Create(context.Background(), investment.CreateInvestmentCommand{})
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		// Missing ProjectID
		_, err = svc.Create(authedCtx, investment.CreateInvestmentCommand{
			AssetID:     existingAsset.ID,
			Amount:      money.New(100, "USD"),
			Description: "Test",
		})
		if err == nil {
			t.Errorf("expected validation error for missing project id")
		}

		// Missing AssetID
		_, err = svc.Create(authedCtx, investment.CreateInvestmentCommand{
			ProjectID:   project.ID,
			Amount:      money.New(100, "USD"),
			Description: "Test",
		})
		if err == nil {
			t.Errorf("expected validation error for missing asset id")
		}

		// Missing Description
		_, err = svc.Create(authedCtx, investment.CreateInvestmentCommand{
			ProjectID: project.ID,
			AssetID:   existingAsset.ID,
			Amount:    money.New(100, "USD"),
		})
		if err == nil {
			t.Errorf("expected validation error for missing description")
		}

		// Amount <= 0
		_, err = svc.Create(authedCtx, investment.CreateInvestmentCommand{
			ProjectID:   project.ID,
			AssetID:     existingAsset.ID,
			Amount:      money.New(0, "USD"),
			Description: "Test",
		})
		if err == nil {
			t.Errorf("expected validation error for 0 amount")
		}
	})

	t.Run("CreateBatch", func(t *testing.T) {
		svc := investment.NewService(
			&mockDb{},
			&mockInvestmentRepo{},
			&mockAssetRepo{asset: existingAsset},
			&mockProjectRepo{project: project},
			&mockPeopleService{owner: owner},
		)

		items := []investment.BatchCreateInvestmentItem{
			{
				AssetID:      existingAsset.ID,
				ResourceType: investment.ResourceTypeMoney,
				Amount:       money.New(200, "USD"),
				Description:  "Batch 1",
			},
			{
				ResourceType: investment.ResourceTypeMoney,
				Amount:       money.New(300, "USD"),
				Description:  "Batch 2",
			},
		}

		created, err := svc.CreateBatch(authedCtx, investment.CreateBatchInvestmentsCommand{
			ProjectID:    project.ID,
			DefaultAsset: existingAsset.ID,
			Items:        items,
		})
		if err != nil {
			t.Fatalf("unexpected error in CreateBatch: %v", err)
		}
		if len(created) != 2 {
			t.Errorf("expected 2 created investments, got %d", len(created))
		}
	})

	t.Run("Update and Remove", func(t *testing.T) {
		inv, err := investment.NewInvestment(
			core.NewInvestmentID(),
			project.ID,
			"USD",
			existingAsset.ID,
			existingAsset.Name,
			investment.ContributorRef{PersonID: owner.PersonID, DisplayName: owner.DisplayName},
			investment.ResourceTypeMoney,
			money.New(100, "USD"),
			nil, nil, nil, "", "",
			"Desc",
			time.Now(),
			[]string{"tag1"},
		)
		if err != nil {
			t.Fatalf("unexpected error creating investment: %v", err)
		}

		db := &mockDb{}
		repo := &mockInvestmentRepo{inv: inv}
		svc := investment.NewService(
			db,
			repo,
			&mockAssetRepo{asset: existingAsset},
			&mockProjectRepo{project: project},
			&mockPeopleService{owner: owner},
		)

		err = svc.Update(authedCtx, investment.UpdateInvestmentCommand{
			ID:          inv.ID(),
			ProjectID:   project.ID,
			Amount:      money.New(250, "USD"),
			Description: "Updated desc",
			Tags:        []string{"tag2"},
		})
		if err != nil {
			t.Errorf("unexpected update error: %v", err)
		}
		if inv.Amount().Amount() != 250 || inv.Description() != "Updated desc" {
			t.Errorf("update did not apply properly: %+v", inv)
		}
		if !db.txCalled {
			t.Errorf("expected db.Tx on update")
		}

		db.txCalled = false
		err = svc.Remove(authedCtx, investment.RemoveInvestmentCommand{
			ID:        inv.ID(),
			ProjectID: project.ID,
		})
		if err != nil {
			t.Errorf("unexpected remove error: %v", err)
		}
		if !db.txCalled {
			t.Errorf("expected db.Tx on remove")
		}
	})

	t.Run("FindTags", func(t *testing.T) {
		repo := &mockInvestmentRepo{tags: []string{"fuel", "labor", "parts"}}
		svc := investment.NewService(
			&mockDb{},
			repo,
			&mockAssetRepo{asset: existingAsset},
			&mockProjectRepo{project: project},
			&mockPeopleService{owner: owner},
		)

		tags, err := svc.FindTags(authedCtx, project.ID)
		if err != nil {
			t.Fatalf("unexpected error finding tags: %v", err)
		}
		if len(tags) != 3 {
			t.Errorf("expected 3 tags, got %d", len(tags))
		}
	})

	t.Run("MultiAssetInvestment", func(t *testing.T) {
		asset2 := investment.AssetDescriptor{ID: core.NewAssetID(), Name: "Second Asset"}
		assetsMap := map[core.AssetID]investment.AssetDescriptor{
			existingAsset.ID: existingAsset,
			asset2.ID:        asset2,
		}

		svc := investment.NewService(
			&mockDb{},
			&mockInvestmentRepo{},
			&mockAssetRepo{assets: assetsMap},
			&mockProjectRepo{project: project},
			&mockPeopleService{owner: owner},
		)

		t.Run("creates successfully when shares equal 100", func(t *testing.T) {
			inv, err := svc.Create(authedCtx, investment.CreateInvestmentCommand{
				ProjectID: project.ID,
				AssetAllocations: []investment.InvestmentAssetInput{
					{AssetID: existingAsset.ID, SharePercentage: 60},
					{AssetID: asset2.ID, SharePercentage: 40},
				},
				ResourceType: investment.ResourceTypeMoney,
				Amount:       money.New(1000, "USD"),
				Description:  "Multi-asset investment",
				Date:         time.Now(),
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(inv.AssetAllocations()) != 2 {
				t.Fatalf("expected 2 allocations, got %d", len(inv.AssetAllocations()))
			}
			if inv.AssetAllocations()[0].SharePercentage != 60 || inv.AssetAllocations()[1].SharePercentage != 40 {
				t.Errorf("unexpected allocations: %+v", inv.AssetAllocations())
			}
			if inv.PrimaryAssetID() != existingAsset.ID {
				t.Errorf("expected primary asset ID to be asset1, got %v", inv.PrimaryAssetID())
			}
		})

		t.Run("fails when shares do not equal 100", func(t *testing.T) {
			_, err := svc.Create(authedCtx, investment.CreateInvestmentCommand{
				ProjectID: project.ID,
				AssetAllocations: []investment.InvestmentAssetInput{
					{AssetID: existingAsset.ID, SharePercentage: 60},
					{AssetID: asset2.ID, SharePercentage: 30},
				},
				ResourceType: investment.ResourceTypeMoney,
				Amount:       money.New(1000, "USD"),
				Description:  "Invalid shares",
				Date:         time.Now(),
			})
			if err == nil {
				t.Fatal("expected error for invalid total shares, got nil")
			}
		})

		t.Run("fails on duplicate asset allocations", func(t *testing.T) {
			_, err := svc.Create(authedCtx, investment.CreateInvestmentCommand{
				ProjectID: project.ID,
				AssetAllocations: []investment.InvestmentAssetInput{
					{AssetID: existingAsset.ID, SharePercentage: 50},
					{AssetID: existingAsset.ID, SharePercentage: 50},
				},
				ResourceType: investment.ResourceTypeMoney,
				Amount:       money.New(1000, "USD"),
				Description:  "Duplicate asset",
				Date:         time.Now(),
			})
			if err == nil {
				t.Fatal("expected error for duplicate asset, got nil")
			}
		})

		t.Run("fails on negative or zero share", func(t *testing.T) {
			_, err := svc.Create(authedCtx, investment.CreateInvestmentCommand{
				ProjectID: project.ID,
				AssetAllocations: []investment.InvestmentAssetInput{
					{AssetID: existingAsset.ID, SharePercentage: 100},
					{AssetID: asset2.ID, SharePercentage: 0},
				},
				ResourceType: investment.ResourceTypeMoney,
				Amount:       money.New(1000, "USD"),
				Description:  "Zero share",
				Date:         time.Now(),
			})
			if err == nil {
				t.Fatal("expected error for zero share, got nil")
			}
		})

		t.Run("defaults to 100% share for single asset investment without user defining share", func(t *testing.T) {
			inv, err := svc.Create(authedCtx, investment.CreateInvestmentCommand{
				ProjectID:    project.ID,
				AssetID:      existingAsset.ID,
				ResourceType: investment.ResourceTypeMoney,
				Amount:       money.New(500, "USD"),
				Description:  "Single asset investment",
				Date:         time.Now(),
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(inv.AssetAllocations()) != 1 {
				t.Fatalf("expected 1 allocation, got %d", len(inv.AssetAllocations()))
			}
			if inv.AssetAllocations()[0].SharePercentage != 100.0 {
				t.Errorf("expected default 100%% share, got %v", inv.AssetAllocations()[0].SharePercentage)
			}
		})

		t.Run("defaults to 100% share when single allocation passed with zero or undefined share", func(t *testing.T) {
			inv, err := svc.Create(authedCtx, investment.CreateInvestmentCommand{
				ProjectID: project.ID,
				AssetAllocations: []investment.InvestmentAssetInput{
					{AssetID: existingAsset.ID, SharePercentage: 0},
				},
				ResourceType: investment.ResourceTypeMoney,
				Amount:       money.New(500, "USD"),
				Description:  "Single asset allocation without share",
				Date:         time.Now(),
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(inv.AssetAllocations()) != 1 {
				t.Fatalf("expected 1 allocation, got %d", len(inv.AssetAllocations()))
			}
			if inv.AssetAllocations()[0].SharePercentage != 100.0 {
				t.Errorf("expected default 100%% share, got %v", inv.AssetAllocations()[0].SharePercentage)
			}
		})
	})
}
