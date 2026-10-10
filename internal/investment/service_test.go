package investment_test

import (
	"context"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/investment"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

type mockDb struct{}

func (m *mockDb) Tx(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
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
func (m *mockInvestmentRepo) FindTags(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	return m.tags, nil
}

type mockAssetRepo struct {
	asset *asset.Asset
	err   error
}

func (m *mockAssetRepo) Save(ctx context.Context, a *asset.Asset) error   { return nil }
func (m *mockAssetRepo) Remove(ctx context.Context, a *asset.Asset) error { return nil }
func (m *mockAssetRepo) FindOne(ctx context.Context, filter asset.Filter) (*asset.Asset, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.asset, nil
}
func (m *mockAssetRepo) Find(ctx context.Context, filter asset.CollectionFilter) (*asset.Collection, error) {
	return nil, nil
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

type mockProjectRepo struct {
	project *projecta.Project
	err     error
}

func (m *mockProjectRepo) Create(ctx context.Context, p *projecta.Project) error { return nil }
func (m *mockProjectRepo) Update(ctx context.Context, p *projecta.Project) error { return nil }
func (m *mockProjectRepo) Remove(ctx context.Context, p *projecta.Project) error { return nil }
func (m *mockProjectRepo) FindByShareToken(ctx context.Context, token uuid.UUID) (*projecta.Project, error) {
	return m.project, nil
}
func (m *mockProjectRepo) CreateShareRecord(ctx context.Context, projectID uuid.UUID, personID uuid.UUID) error {
	return nil
}
func (m *mockProjectRepo) Find(ctx context.Context, filter projecta.ProjectCollectionFilter) ([]*projecta.Project, error) {
	return nil, nil
}
func (m *mockProjectRepo) FindOne(ctx context.Context, filter projecta.ProjectFilter) (*projecta.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.project, nil
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

func TestInvestmentService(t *testing.T) {
	personID := uuid.New()
	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, personID)

	owner := &projecta.Owner{PersonID: personID, DisplayName: "Alex"}
	project, _ := projecta.NewProject(uuid.New(), "Project Alpha", "Description", owner, time.Now(), time.Now())
	existingAsset := asset.NewAsset(uuid.New(), "Vehicle", "Car", project, nil, money.New(10000, "USD"), time.Now(), owner)

	t.Run("Create success (Money, Time, Goods)", func(t *testing.T) {
		repo := &mockInvestmentRepo{}
		svc := investment.NewService(
			&mockDb{},
			repo,
			&mockAssetRepo{asset: existingAsset},
			&mockProjectRepo{project: project},
			&mockPeopleService{owner: owner},
		)

		// Money investment
		inv, err := svc.Create(authedCtx, investment.CreateInvestmentCommand{
			ProjectID:    project.ProjectID,
			AssetID:      existingAsset.ID(),
			ResourceType: investment.ResourceTypeMoney,
			Amount:       money.New(500, "USD"),
			Description:  "Fuel purchase",
			Tags:         []string{"fuel", "travel"},
		})
		if err != nil {
			t.Fatalf("unexpected error creating money investment: %v", err)
		}
		if inv.Amount.Amount() != 500 || inv.ResourceType != investment.ResourceTypeMoney {
			t.Errorf("unexpected money investment details: %+v", inv)
		}

		// Time investment with calculated rate
		hours := 4.5
		rate := money.New(4000, "USD") // $40/hr
		invTime, err := svc.Create(authedCtx, investment.CreateInvestmentCommand{
			ProjectID:      project.ProjectID,
			AssetID:        existingAsset.ID(),
			ResourceType:   investment.ResourceTypeTime,
			TimeHours:      &hours,
			TimeHourlyRate: rate,
			Description:    "Brake repair",
			Tags:           []string{"labor", "maintenance"},
		})
		if err != nil {
			t.Fatalf("unexpected error creating time investment: %v", err)
		}
		if invTime.Amount.Amount() != 18000 { // 4.5 * 4000 = 18000
			t.Errorf("expected 18000 amount for time, got %d", invTime.Amount.Amount())
		}

		// Goods investment
		qty := 4.0
		invGoods, err := svc.Create(authedCtx, investment.CreateInvestmentCommand{
			ProjectID:     project.ProjectID,
			AssetID:       existingAsset.ID(),
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
		if invGoods.GoodsItemName != "Winter Tires" {
			t.Errorf("expected GoodsItemName 'Winter Tires', got %s", invGoods.GoodsItemName)
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
			AssetID:     existingAsset.ID(),
			Amount:      money.New(100, "USD"),
			Description: "Test",
		})
		if err == nil {
			t.Errorf("expected validation error for missing project id")
		}

		// Missing AssetID
		_, err = svc.Create(authedCtx, investment.CreateInvestmentCommand{
			ProjectID:   project.ProjectID,
			Amount:      money.New(100, "USD"),
			Description: "Test",
		})
		if err == nil {
			t.Errorf("expected validation error for missing asset id")
		}

		// Missing Description
		_, err = svc.Create(authedCtx, investment.CreateInvestmentCommand{
			ProjectID: project.ProjectID,
			AssetID:   existingAsset.ID(),
			Amount:    money.New(100, "USD"),
		})
		if err == nil {
			t.Errorf("expected validation error for missing description")
		}

		// Amount <= 0
		_, err = svc.Create(authedCtx, investment.CreateInvestmentCommand{
			ProjectID:   project.ProjectID,
			AssetID:     existingAsset.ID(),
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
				AssetID:      existingAsset.ID(),
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
			ProjectID:    project.ProjectID,
			DefaultAsset: existingAsset.ID(),
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
		inv := investment.NewInvestment(
			uuid.New(),
			project,
			existingAsset,
			owner,
			investment.ResourceTypeMoney,
			money.New(100, "USD"),
			nil, nil, nil, "", "",
			"Desc",
			time.Now(),
			[]string{"tag1"},
		)

		repo := &mockInvestmentRepo{inv: inv}
		svc := investment.NewService(
			&mockDb{},
			repo,
			&mockAssetRepo{asset: existingAsset},
			&mockProjectRepo{project: project},
			&mockPeopleService{owner: owner},
		)

		err := svc.Update(authedCtx, investment.UpdateInvestmentCommand{
			ID:          inv.ID,
			ProjectID:   project.ProjectID,
			Amount:      money.New(250, "USD"),
			Description: "Updated desc",
			Tags:        []string{"tag2"},
		})
		if err != nil {
			t.Errorf("unexpected update error: %v", err)
		}
		if inv.Amount.Amount() != 250 || inv.Description != "Updated desc" {
			t.Errorf("update did not apply properly: %+v", inv)
		}

		err = svc.Remove(authedCtx, investment.RemoveInvestmentCommand{
			ID:        inv.ID,
			ProjectID: project.ProjectID,
		})
		if err != nil {
			t.Errorf("unexpected remove error: %v", err)
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

		tags, err := svc.FindTags(authedCtx, project.ProjectID)
		if err != nil {
			t.Fatalf("unexpected error finding tags: %v", err)
		}
		if len(tags) != 3 {
			t.Errorf("expected 3 tags, got %d", len(tags))
		}
	})
}
