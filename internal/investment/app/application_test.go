package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/investment"
	investmentapp "gitlab.com/massimo-ua/projecta/internal/investment/app"
)

type noopTxDB struct{}

func (noopTxDB) Tx(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	return fn(ctx)
}
func (noopTxDB) Close() {}
func (noopTxDB) Ping(_ context.Context) error { return nil }

type invProjectReaderStub struct {
	proj investmentapp.ProjectDescriptor
	err  error
}

func (s invProjectReaderStub) FindProject(_ context.Context, _ core.ProjectID, _ core.PersonID) (investmentapp.ProjectDescriptor, error) {
	if s.err != nil {
		return investmentapp.ProjectDescriptor{}, s.err
	}
	return s.proj, nil
}

type invAssetReaderStub struct {
	assets map[core.AssetID]investmentapp.AssetDescriptor
	err    error
}

func (s invAssetReaderStub) FindAsset(_ context.Context, _ core.ProjectID, assetID core.AssetID) (investmentapp.AssetDescriptor, error) {
	if s.err != nil {
		return investmentapp.AssetDescriptor{}, s.err
	}
	a, ok := s.assets[assetID]
	if !ok {
		return investmentapp.AssetDescriptor{}, errors.New("asset not found")
	}
	return a, nil
}

type invContributorReaderStub struct {
	ref investment.ContributorRef
	err error
}

func (s invContributorReaderStub) FindContributor(_ context.Context, _ core.PersonID) (investment.ContributorRef, error) {
	if s.err != nil {
		return investment.ContributorRef{}, s.err
	}
	return s.ref, nil
}

type invConverterStub struct {
	rate float64
}

func (c invConverterStub) ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error) {
	if amount == nil {
		return money.New(0, targetCurrency), nil
	}
	if amount.Currency().Code == targetCurrency {
		return amount, nil
	}
	return money.New(int64(float64(amount.Amount())*c.rate), targetCurrency), nil
}

func TestNewApplication_FailFastOnMissingDependencies(t *testing.T) {
	repo := investment.NewMemoryInvestmentRepository()
	db := noopTxDB{}
	assets := invAssetReaderStub{}
	projects := invProjectReaderStub{}
	contributors := invContributorReaderStub{}

	cases := []struct {
		name string
		fn   func()
	}{
		{"nil_db", func() { investmentapp.NewApplication(nil, repo, repo, assets, projects, contributors, nil, nil) }},
		{"nil_repo", func() { investmentapp.NewApplication(db, nil, repo, assets, projects, contributors, nil, nil) }},
		{"nil_tags_read_model", func() { investmentapp.NewListProjectTagsHandler(nil, projects, nil) }},
		{"nil_assets", func() { investmentapp.NewApplication(db, repo, repo, nil, projects, contributors, nil, nil) }},
		{"nil_projects", func() { investmentapp.NewApplication(db, repo, repo, assets, nil, contributors, nil, nil) }},
		{"nil_contributors", func() { investmentapp.NewApplication(db, repo, repo, assets, projects, nil, nil, nil) }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected panic for %s", tc.name)
				}
			}()
			tc.fn()
		})
	}
}

func TestInvestmentApplication_FullCQRSFlow(t *testing.T) {
	ctx := context.Background()
	requesterID := core.NewPersonID()
	projectID := core.NewProjectID()
	asset1 := core.NewAssetID()
	asset2 := core.NewAssetID()

	repo := investment.NewMemoryInvestmentRepository()
	app := investmentapp.NewApplication(
		noopTxDB{},
		repo,
		repo,
		invAssetReaderStub{assets: map[core.AssetID]investmentapp.AssetDescriptor{
			asset1: {ID: asset1, ProjectID: projectID, Name: "Solar Panels"},
			asset2: {ID: asset2, ProjectID: projectID, Name: "Inverter"},
		}},
		invProjectReaderStub{proj: investmentapp.ProjectDescriptor{ID: projectID, MainCurrency: "UAH"}},
		invContributorReaderStub{ref: investment.ContributorRef{PersonID: requesterID, DisplayName: "Alice"}},
		invConverterStub{rate: 40},
		nil,
	)

	// 1. CreateInvestment with pre-generated ID
	invID := core.NewInvestmentID()
	err := app.Commands.CreateInvestment.Handle(ctx, investmentapp.CreateInvestment{
		ID:           invID,
		RequesterID:  requesterID,
		ProjectID:    projectID,
		AssetID:      asset1,
		ResourceType: investment.ResourceTypeMoney,
		Amount:       money.New(10000, "USD"),
		Description:  "Buy panels",
		Date:         time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		Tags:         []string{"solar", "hardware"},
	})
	if err != nil {
		t.Fatalf("CreateInvestment failed: %v", err)
	}

	got, err := app.Queries.GetInvestment.Handle(ctx, investmentapp.GetInvestment{
		ID:          invID,
		ProjectID:   projectID,
		RequesterID: requesterID,
	})
	if err != nil {
		t.Fatalf("GetInvestment failed: %v", err)
	}
	if got.HomeAmount() != 400000 || got.Description() != "Buy panels" {
		t.Fatalf("unexpected investment: home=%d desc=%q", got.HomeAmount(), got.Description())
	}

	// 2. UpdateInvestment via locked closure + split allocations
	err = app.Commands.UpdateInvestment.Handle(ctx, investmentapp.UpdateInvestment{
		ID:          invID,
		RequesterID: requesterID,
		ProjectID:   projectID,
		AssetAllocations: []investmentapp.AssetAllocationInput{
			{AssetID: asset1, SharePercentage: 60},
			{AssetID: asset2, SharePercentage: 40},
		},
		ResourceType: investment.ResourceTypeMoney,
		Amount:       money.New(20000, "USD"),
		Description:  "Panels + Inverter",
		Date:         time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
		Tags:         []string{"solar"},
	})
	if err != nil {
		t.Fatalf("UpdateInvestment failed: %v", err)
	}

	updated, err := app.Queries.GetInvestment.Handle(ctx, investmentapp.GetInvestment{
		ID:          invID,
		ProjectID:   projectID,
		RequesterID: requesterID,
	})
	if err != nil || len(updated.AssetAllocations()) != 2 || updated.HomeAmount() != 800000 {
		t.Fatalf("unexpected updated investment: allocs=%d home=%d err=%v", len(updated.AssetAllocations()), updated.HomeAmount(), err)
	}

	// 3. Invalid update rolls back cleanly in MemoryInvestmentRepository
	if err := app.Commands.UpdateInvestment.Handle(ctx, investmentapp.UpdateInvestment{
		ID:          invID,
		RequesterID: requesterID,
		ProjectID:   projectID,
		AssetAllocations: []investmentapp.AssetAllocationInput{
			{AssetID: asset1, SharePercentage: 50},
			{AssetID: asset2, SharePercentage: 10}, // invalid: sums to 60 != 100
		},
		Description: "Should rollback",
	}); err == nil {
		t.Fatal("expected validation error for invalid allocation sum")
	}
	afterErr, _ := app.Queries.GetInvestment.Handle(ctx, investmentapp.GetInvestment{
		ID:          invID,
		ProjectID:   projectID,
		RequesterID: requesterID,
	})
	if afterErr.Description() != "Panels + Inverter" {
		t.Fatalf("rollback leaked description: %q", afterErr.Description())
	}

	// 4. CreateBatchInvestments + AssignInvestmentsToAsset + ListProjectTags
	batchID1 := core.NewInvestmentID()
	batchID2 := core.NewInvestmentID()
	err = app.Commands.CreateBatchInvestments.Handle(ctx, investmentapp.CreateBatchInvestments{
		RequesterID:  requesterID,
		ProjectID:    projectID,
		DefaultAsset: asset1,
		Items: []investmentapp.BatchCreateInvestmentItem{
			{ID: batchID1, Amount: money.New(5000, "UAH"), Description: "Cable", Tags: []string{"wiring"}},
			{ID: batchID2, Amount: money.New(3000, "UAH"), Description: "Mount", Tags: []string{"mounting"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateBatchInvestments failed: %v", err)
	}

	if err := app.Commands.AssignInvestmentsToAsset.Handle(ctx, investmentapp.AssignInvestmentsToAsset{
		RequesterID:   requesterID,
		AssetID:       asset2,
		InvestmentIDs: []core.InvestmentID{batchID1},
	}); err != nil {
		t.Fatalf("AssignInvestmentsToAsset failed: %v", err)
	}
	reassigned, _ := app.Queries.GetInvestment.Handle(ctx, investmentapp.GetInvestment{
		ID:          batchID1,
		ProjectID:   projectID,
		RequesterID: requesterID,
	})
	if reassigned.PrimaryAssetID() != asset2 {
		t.Fatalf("expected reassigned primary asset %s, got %s", asset2, reassigned.PrimaryAssetID())
	}

	tags, err := app.Queries.ListProjectTags.Handle(ctx, investmentapp.ListProjectTags{
		ProjectID:   projectID,
		RequesterID: requesterID,
	})
	if err != nil || len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %v err=%v", tags, err)
	}

	// 5. List & Remove
	col, err := app.Queries.ListInvestments.Handle(ctx, investmentapp.ListInvestments{
		RequesterID: requesterID,
		ProjectID:   projectID,
		Pagination:  core.Pagination{Limit: 10},
	})
	if err != nil || col.Total() != 3 {
		t.Fatalf("expected 3 investments, got total=%d err=%v", col.Total(), err)
	}

	if err := app.Commands.RemoveInvestment.Handle(ctx, investmentapp.RemoveInvestment{
		ID:          batchID2,
		ProjectID:   projectID,
		RequesterID: requesterID,
	}); err != nil {
		t.Fatalf("RemoveInvestment failed: %v", err)
	}
}
