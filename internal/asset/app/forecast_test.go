package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	assetapp "gitlab.com/massimo-ua/projecta/internal/asset/app"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

type fixedAppConverter struct{}

func (fixedAppConverter) ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error) {
	if amount == nil {
		return money.New(0, targetCurrency), nil
	}
	if amount.Currency().Code == targetCurrency {
		return amount, nil
	}
	return money.New(amount.Amount()*40, targetCurrency), nil
}

type trackingProjectReader struct {
	project        asset.ProjectDescriptor
	ceilings       map[string]int64
	lastCeilingTag string
}

func (r *trackingProjectReader) FindProject(_ context.Context, _ core.ProjectID, _ core.PersonID) (asset.ProjectDescriptor, error) {
	return r.project, nil
}

func (r *trackingProjectReader) FindMonthlyCashCeilings(_ context.Context, _ core.ProjectID, _ core.PersonID, _ []string, tag string) (map[string]int64, error) {
	r.lastCeilingTag = tag
	return r.ceilings, nil
}

type controllableInvestmentSource struct {
	summaries []asset.InvestmentRecordSummary
	listErr   error
}

func (s *controllableInvestmentSource) FindInvestments(_ context.Context, _ core.ProjectID, _ core.PersonID, _ []core.InvestmentID) ([]asset.InvestmentItem, error) {
	return nil, nil
}

func (s *controllableInvestmentSource) ListProjectInvestments(_ context.Context, _ core.ProjectID, _ core.PersonID) ([]asset.InvestmentRecordSummary, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.summaries, nil
}

func (s *controllableInvestmentSource) AssignToAsset(_ context.Context, _ core.AssetID, _ core.PersonID, _ []core.InvestmentID) error {
	return nil
}

func (s *controllableInvestmentSource) CreateInitialInvestment(_ context.Context, _ asset.InitialInvestment) error {
	return nil
}

type noopTxDBForecast struct{}

func (noopTxDBForecast) Tx(ctx context.Context, fn func(context.Context) (any, error)) (any, error) {
	return fn(ctx)
}
func (noopTxDBForecast) Close()                       {}
func (noopTxDBForecast) Ping(_ context.Context) error { return nil }

type staticOwnerReader struct {
	owner asset.OwnerRef
}

func (r staticOwnerReader) FindOwner(_ context.Context, _ core.PersonID) (asset.OwnerRef, error) {
	return r.owner, nil
}

func TestApplicationMultiCurrencyScheduleAndForecast(t *testing.T) {
	ctx := context.Background()
	ownerID := core.NewPersonID()
	pID := core.NewProjectID()
	ownerRef := asset.OwnerRef{PersonID: ownerID, DisplayName: "Max"}

	repo := asset.NewMemoryAssetRepository()
	child, _ := asset.CreateAsset(core.NewAssetID(), "Child", "", pID, money.New(500, "UAH"), time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), ownerRef, asset.AssetStatusActive, nil, nil)
	childTarget := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	_ = child.UpdateSchedule(nil, &childTarget, false)

	parent, _ := asset.CreateAsset(core.NewAssetID(), "Parent", "", pID, money.New(2000, "UAH"), time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), ownerRef, asset.AssetStatusActive, nil, nil)
	_ = repo.Save(ctx, parent)
	_ = repo.Save(ctx, child)
	_ = repo.AddChild(ctx, parent.ID(), child.ID(), 100)

	projReader := &trackingProjectReader{
		project: asset.ProjectDescriptor{
			ID:           pID,
			MainCurrency: "UAH",
			StartDate:    time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC),
			EndDate:      time.Date(2026, time.August, 31, 0, 0, 0, 0, time.UTC),
		},
		ceilings: map[string]int64{"2026-05": 10000},
	}
	invSrc := &controllableInvestmentSource{
		summaries: []asset.InvestmentRecordSummary{
			{
				ID:           core.NewInvestmentID(),
				ResourceType: "MONEY",
				Amount:       money.New(10, "USD"), // 400 UAH in child -> rolls up 100% into parent
				Date:         time.Date(2026, time.May, 2, 0, 0, 0, 0, time.UTC),
				AssetAllocations: []asset.InvestmentAllocation{
					{AssetID: child.ID(), SharePercentage: 100},
				},
			},
		},
	}

	app := assetapp.NewApplication(noopTxDBForecast{}, repo, staticOwnerReader{owner: ownerRef}, projReader, invSrc, fixedAppConverter{}, nil)

	// ListAssets normalizes mixed-currency direct & rolled-up costs in UAH
	foundCol, err := app.Queries.ListAssets.Handle(ctx, assetapp.ListAssets{RequesterID: ownerID, ProjectID: pID})
	if err != nil {
		t.Fatalf("ListAssets failed: %v", err)
	}
	var gotParent, gotChild *asset.Asset
	for _, a := range foundCol.Elements() {
		if a.ID() == parent.ID() {
			gotParent = a
		}
		if a.ID() == child.ID() {
			gotChild = a
		}
	}
	if gotParent == nil || gotParent.TotalCost() == nil || gotParent.TotalCost().Amount() != 400 || gotParent.TotalCost().Currency().Code != "UAH" {
		t.Fatalf("expected parent TotalCost rolled up to 400 UAH, got %v", gotParent)
	}
	if gotParent.DirectCostHomeAmount() != 0 {
		t.Errorf("expected 0 DirectCostHomeAmount for group parent, got %d", gotParent.DirectCostHomeAmount())
	}
	if gotChild == nil || gotChild.DirectCostHomeAmount() != 400 {
		t.Errorf("expected 400 DirectCostHomeAmount for USD child, got %v", gotChild)
	}

	// CreateAsset unbudgeted succeeds, while WithPayment without positive price fails
	unbID := core.NewAssetID()
	if err := app.Commands.CreateAsset.Handle(ctx, assetapp.CreateAsset{
		AssetID:     unbID,
		RequesterID: ownerID,
		Name:        "Unbudgeted",
		ProjectID:   pID,
	}); err != nil {
		t.Fatalf("expected unbudgeted asset create to succeed, got: %v", err)
	}
	unb, err := app.Queries.GetAsset.Handle(ctx, assetapp.GetAsset{ID: unbID, ProjectID: pID, RequesterID: ownerID})
	if err != nil || unb.HasTargetBudget() {
		t.Errorf("expected unbudgeted asset HasTargetBudget=false, got asset=%v err=%v", unb, err)
	}
	if err := app.Commands.CreateAsset.Handle(ctx, assetapp.CreateAsset{
		AssetID:     core.NewAssetID(),
		RequesterID: ownerID,
		Name:        "Needs Payment",
		ProjectID:   pID,
		WithPayment: true,
	}); err == nil {
		t.Errorf("expected error when WithPayment=true and no target price")
	}

	// RescheduleAsset + GetAssetScheduleView returns non-blocking warning when parent target precedes child target
	earlierParentTarget := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	if err := app.Commands.RescheduleAsset.Handle(ctx, assetapp.RescheduleAsset{
		AssetID:              parent.ID(),
		ProjectID:            pID,
		RequesterID:          ownerID,
		TargetCompletionDate: &earlierParentTarget,
	}); err != nil {
		t.Fatalf("RescheduleAsset parent failed: %v", err)
	}
	view, err := app.Queries.GetAssetScheduleView.Handle(ctx, assetapp.GetAssetScheduleView{AssetID: parent.ID(), ProjectID: pID, RequesterID: ownerID})
	if err != nil {
		t.Fatalf("GetAssetScheduleView failed: %v", err)
	}
	if len(view.Warnings) != 1 || !strings.HasPrefix(view.Warnings[0], "child_target_exceeds_parent:") {
		t.Errorf("expected child_target_exceeds_parent warning, got %v", view.Warnings)
	}

	// Child schedule view also surfaces parent warning
	childView, err := app.Queries.GetAssetScheduleView.Handle(ctx, assetapp.GetAssetScheduleView{AssetID: child.ID(), ProjectID: pID, RequesterID: ownerID})
	if err != nil || len(childView.Warnings) == 0 {
		t.Errorf("expected parent warning on child schedule view, got view=%+v err=%v", childView, err)
	}

	// GetCashFlowForecast validation, error propagation, and tag ceiling forwarding
	if _, err := app.Queries.GetCashFlowForecast.Handle(ctx, assetapp.GetCashFlowForecast{ProjectID: pID}); err == nil {
		t.Errorf("expected unauthorized error for nil requesterID")
	}
	if _, err := app.Queries.GetCashFlowForecast.Handle(ctx, assetapp.GetCashFlowForecast{RequesterID: ownerID}); err == nil {
		t.Errorf("expected validation error for missing project id")
	}
	if _, err := app.Queries.GetCashFlowForecast.Handle(ctx, assetapp.GetCashFlowForecast{
		RequesterID: ownerID,
		ProjectID:   pID,
		Tag:         strings.Repeat("a", 101),
	}); err == nil {
		t.Errorf("expected validation error for oversized tag filter")
	}

	invSrc.listErr = errors.New("inv boom")
	if _, err := app.Queries.GetCashFlowForecast.Handle(ctx, assetapp.GetCashFlowForecast{RequesterID: ownerID, ProjectID: pID}); err == nil {
		t.Errorf("expected investment source error propagation")
	}
	invSrc.listErr = nil

	fc, err := app.Queries.GetCashFlowForecast.Handle(ctx, assetapp.GetCashFlowForecast{
		RequesterID: ownerID,
		ProjectID:   pID,
		FromMonth:   "2026-05",
		ToMonth:     "2026-08",
		Tag:         "edu",
		AsOfDate:    time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC),
	})
	if err != nil || fc == nil || len(fc.Months) != 4 {
		t.Fatalf("expected 4-month forecast, got fc=%v err=%v", fc, err)
	}
	if projReader.lastCeilingTag != "edu" {
		t.Errorf("expected FindMonthlyCashCeilings to receive tag 'edu', got %q", projReader.lastCeilingTag)
	}
}
