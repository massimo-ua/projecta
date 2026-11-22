package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	assetapp "gitlab.com/massimo-ua/projecta/internal/asset/app"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type noopAssetTxDB struct{}

func (noopAssetTxDB) Tx(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	return fn(ctx)
}
func (noopAssetTxDB) Close()                       {}
func (noopAssetTxDB) Ping(_ context.Context) error { return nil }

type assetOwnerReaderStub struct {
	owner asset.OwnerRef
	err   error
}

func (s assetOwnerReaderStub) FindOwner(_ context.Context, _ core.PersonID) (asset.OwnerRef, error) {
	return s.owner, s.err
}

type assetProjectReaderStub struct {
	proj     asset.ProjectDescriptor
	ceilings map[string]int64
	err      error
}

func (s assetProjectReaderStub) FindProject(_ context.Context, _ core.ProjectID, _ core.PersonID) (asset.ProjectDescriptor, error) {
	return s.proj, s.err
}

func (s assetProjectReaderStub) FindMonthlyCashCeilings(_ context.Context, _ core.ProjectID, _ core.PersonID, _ []string, _ string) (map[string]int64, error) {
	if s.ceilings == nil {
		return map[string]int64{}, s.err
	}
	return s.ceilings, s.err
}

type assetInvestmentSourceStub struct {
	items     []asset.InvestmentItem
	summaries []asset.InvestmentRecordSummary
	created   []asset.InitialInvestment
	assigned  map[core.AssetID][]core.InvestmentID
}

func (s *assetInvestmentSourceStub) FindInvestments(_ context.Context, _ core.ProjectID, _ core.PersonID, _ []core.InvestmentID) ([]asset.InvestmentItem, error) {
	return s.items, nil
}
func (s *assetInvestmentSourceStub) ListProjectInvestments(_ context.Context, _ core.ProjectID, _ core.PersonID) ([]asset.InvestmentRecordSummary, error) {
	return s.summaries, nil
}
func (s *assetInvestmentSourceStub) AssignToAsset(_ context.Context, assetID core.AssetID, _ core.PersonID, ids []core.InvestmentID) error {
	if s.assigned == nil {
		s.assigned = make(map[core.AssetID][]core.InvestmentID)
	}
	s.assigned[assetID] = append(s.assigned[assetID], ids...)
	return nil
}
func (s *assetInvestmentSourceStub) CreateInitialInvestment(_ context.Context, item asset.InitialInvestment) error {
	s.created = append(s.created, item)
	return nil
}

func TestNewApplication_FailFastOnMissingDependencies(t *testing.T) {
	repo := asset.NewMemoryAssetRepository()
	db := noopAssetTxDB{}
	owners := assetOwnerReaderStub{}
	projects := assetProjectReaderStub{}

	cases := []struct {
		name string
		fn   func()
	}{
		{"nil_db", func() { assetapp.NewApplication(nil, repo, owners, projects, nil, nil, nil) }},
		{"nil_repo", func() { assetapp.NewApplication(db, nil, owners, projects, nil, nil, nil) }},
		{"nil_owners", func() { assetapp.NewApplication(db, repo, nil, projects, nil, nil, nil) }},
		{"nil_projects", func() { assetapp.NewApplication(db, repo, owners, nil, nil, nil, nil) }},
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

func TestAssetApplication_FullCQRSFlow(t *testing.T) {
	ctx := context.Background()
	requesterID := core.NewPersonID()
	projectID := core.NewProjectID()

	repo := asset.NewMemoryAssetRepository()
	invSource := &assetInvestmentSourceStub{}
	app := assetapp.NewApplication(
		noopAssetTxDB{},
		repo,
		assetOwnerReaderStub{owner: asset.OwnerRef{PersonID: requesterID, DisplayName: "Alice"}},
		assetProjectReaderStub{proj: asset.ProjectDescriptor{ID: projectID, MainCurrency: "UAH", StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}},
		invSource,
		nil,
		nil,
	)

	// 1. CreateAsset with pre-generated AssetID and WithPayment
	parentID := core.NewAssetID()
	parentTarget := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	if err := app.Commands.CreateAsset.Handle(ctx, assetapp.CreateAsset{
		AssetID:              parentID,
		RequesterID:          requesterID,
		ProjectID:            projectID,
		Name:                 "Roof Solar",
		Description:          "Primary installation",
		Status:               asset.AssetStatusActive,
		TargetPrice:          money.New(100000, "UAH"),
		StartDate:            time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		TargetCompletionDate: &parentTarget,
		WithPayment:          true,
		Tags:                 []string{"solar"},
	}); err != nil {
		t.Fatalf("CreateAsset failed: %v", err)
	}
	if len(invSource.created) != 1 || invSource.created[0].AssetID != parentID {
		t.Fatalf("expected 1 initial investment created for %s, got %+v", parentID, invSource.created)
	}

	// 2. Create child asset with target date exceeding parent to test DAG schedule warnings
	childID := core.NewAssetID()
	childTarget := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	if err := app.Commands.CreateAsset.Handle(ctx, assetapp.CreateAsset{
		AssetID:              childID,
		RequesterID:          requesterID,
		ProjectID:            projectID,
		Name:                 "Battery Bank",
		Status:               asset.AssetStatusActive,
		TargetPrice:          money.New(50000, "UAH"),
		StartDate:            time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		TargetCompletionDate: &childTarget,
	}); err != nil {
		t.Fatalf("CreateAsset child failed: %v", err)
	}

	if err := app.Commands.LinkChildAsset.Handle(ctx, assetapp.LinkChildAsset{
		RequesterID:     requesterID,
		ProjectID:       projectID,
		ParentID:        parentID,
		ChildID:         childID,
		SharePercentage: 100,
	}); err != nil {
		t.Fatalf("LinkChildAsset failed: %v", err)
	}

	// 3. RescheduleAsset + GetAssetScheduleView detects child target exceeding parent
	earlierParent := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	if err := app.Commands.RescheduleAsset.Handle(ctx, assetapp.RescheduleAsset{
		RequesterID:          requesterID,
		ProjectID:            projectID,
		AssetID:              parentID,
		TargetCompletionDate: &earlierParent,
	}); err != nil {
		t.Fatalf("RescheduleAsset failed: %v", err)
	}

	view, err := app.Queries.GetAssetScheduleView.Handle(ctx, assetapp.GetAssetScheduleView{
		AssetID:     parentID,
		ProjectID:   projectID,
		RequesterID: requesterID,
	})
	if err != nil {
		t.Fatalf("GetAssetScheduleView failed: %v", err)
	}
	if len(view.Warnings) != 1 {
		t.Fatalf("expected 1 schedule warning, got %v", view.Warnings)
	}

	// 4. UpdateAsset via locked closure + rollback on invalid details
	if err := app.Commands.UpdateAsset.Handle(ctx, assetapp.UpdateAsset{
		AssetID:     parentID,
		RequesterID: requesterID,
		ProjectID:   projectID,
		Name:        "Roof Solar Pro",
		Description: "Updated",
		Status:      asset.AssetStatusActive,
		TargetPrice: money.New(120000, "UAH"),
	}); err != nil {
		t.Fatalf("UpdateAsset failed: %v", err)
	}

	if err := app.Commands.UpdateAsset.Handle(ctx, assetapp.UpdateAsset{
		AssetID:     parentID,
		RequesterID: requesterID,
		ProjectID:   projectID,
		Name:        "", // invalid
	}); err == nil {
		t.Fatal("expected validation error on empty name")
	}

	reloaded, err := app.Queries.GetAsset.Handle(ctx, assetapp.GetAsset{
		ID:          parentID,
		ProjectID:   projectID,
		RequesterID: requesterID,
	})
	if err != nil || reloaded.Name() != "Roof Solar Pro" {
		t.Fatalf("unexpected reloaded asset: name=%q err=%v", reloaded.Name(), err)
	}

	// 5. GroupAssets + ListAssets + GetCashFlowForecast + RemoveAsset
	groupID := core.NewAssetID()
	if err := app.Commands.GroupAssets.Handle(ctx, assetapp.GroupAssets{
		GroupAssetID:  groupID,
		RequesterID:   requesterID,
		ProjectID:     projectID,
		Name:          "Energy System",
		ChildAssetIDs: []core.AssetID{parentID},
	}); err != nil {
		t.Fatalf("GroupAssets failed: %v", err)
	}

	col, err := app.Queries.ListAssets.Handle(ctx, assetapp.ListAssets{
		RequesterID: requesterID,
		ProjectID:   projectID,
		Pagination:  core.Pagination{Limit: 10},
	})
	if err != nil || col.Total() != 3 {
		t.Fatalf("expected 3 assets, got total=%d err=%v", col.Total(), err)
	}

	fc, err := app.Queries.GetCashFlowForecast.Handle(ctx, assetapp.GetCashFlowForecast{
		RequesterID: requesterID,
		ProjectID:   projectID,
		FromMonth:   "2026-01",
		ToMonth:     "2026-06",
	})
	if err != nil || fc == nil || len(fc.Months) != 6 {
		t.Fatalf("unexpected forecast: %+v err=%v", fc, err)
	}

	if err := app.Commands.RemoveAsset.Handle(ctx, assetapp.RemoveAsset{
		AssetID:     groupID,
		ProjectID:   projectID,
		RequesterID: requesterID,
	}); err != nil {
		t.Fatalf("RemoveAsset failed: %v", err)
	}
}
