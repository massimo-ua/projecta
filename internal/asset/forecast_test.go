package asset_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

func mustUTCDate(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

type mockLookupAssetRepo struct {
	mockAssetRepo
	byID map[core.AssetID]*asset.Asset
}

func (m *mockLookupAssetRepo) FindOne(ctx context.Context, filter asset.Filter) (*asset.Asset, error) {
	if m.findOneErr != nil {
		return nil, m.findOneErr
	}
	if m.byID != nil {
		return m.byID[filter.ID], nil
	}
	return m.asset, nil
}

func TestAssetOptionalTargetAndSchedule(t *testing.T) {
	pID := core.NewProjectID()
	owner := asset.OwnerRef{PersonID: core.NewPersonID(), DisplayName: "Max"}
	start := mustUTCDate(2026, time.March, 15)

	// Unbudgeted ongoing asset allowed when target price is nil
	unbudgeted, err := asset.CreateAsset(
		core.NewAssetID(),
		"Coffee Routine",
		"Ongoing tracker",
		pID,
		nil,
		start,
		owner,
		asset.AssetStatusActive,
		nil,
		[]string{"lifestyle"},
	)
	if err != nil {
		t.Fatalf("expected unbudgeted asset creation to succeed, got: %v", err)
	}
	if unbudgeted.HasTargetBudget() {
		t.Errorf("expected HasTargetBudget false for nil target price")
	}
	if unbudgeted.TargetPrice() != nil {
		t.Errorf("expected nil TargetPrice, got %v", unbudgeted.TargetPrice())
	}

	// UpdateSchedule rejects target_completion_date before start_date in UTC
	beforeStart := mustUTCDate(2026, time.March, 10)
	if err := unbudgeted.UpdateSchedule(nil, &beforeStart, false); err == nil {
		t.Errorf("expected validation error when target_completion_date < start_date")
	}

	// Same calendar day across UTC midnight is valid
	sameDayLate := time.Date(2026, time.March, 15, 22, 0, 0, 0, time.UTC)
	if err := unbudgeted.UpdateSchedule(nil, &sameDayLate, false); err != nil {
		t.Errorf("expected same UTC calendar day schedule to succeed, got %v", err)
	}
	if unbudgeted.TargetCompletionDate() == nil {
		t.Fatalf("expected TargetCompletionDate to be set")
	}

	// Clearing target completion date works cleanly
	newStart := mustUTCDate(2026, time.April, 1)
	if err := unbudgeted.UpdateSchedule(&newStart, nil, true); err != nil {
		t.Errorf("expected schedule clear to succeed, got %v", err)
	}
	if unbudgeted.TargetCompletionDate() != nil {
		t.Errorf("expected TargetCompletionDate to be cleared")
	}
	if !unbudgeted.StartDate().Equal(newStart) {
		t.Errorf("expected StartDate %v, got %v", newStart, unbudgeted.StartDate())
	}
}

func TestCalculateProjectCashFlowForecast(t *testing.T) {
	pID := core.NewProjectID()
	owner := asset.OwnerRef{PersonID: core.NewPersonID(), DisplayName: "Max"}
	asOf := mustUTCDate(2026, time.May, 10)

	// Asset 1: Budgeted USD asset with future target completion date (exact cent split May-Jul 2026)
	// $1,000 USD = 40,000 UAH cents via mockRateProvider (40x)
	// Let's use USD 100 cents ($1.00 = 4000 UAH cents), with 1 USD (40 UAH) -> easier: 1000 cents UAH target, 0 spent.
	targetEnd := mustUTCDate(2026, time.July, 10)
	scheduledAsset, err := asset.CreateAsset(
		core.NewAssetID(), "Kitchen Renovation", "", pID,
		money.New(10000, "UAH"), mustUTCDate(2026, time.May, 1),
		owner, asset.AssetStatusActive, nil, []string{"home"},
	)
	if err != nil {
		t.Fatalf("create scheduledAsset: %v", err)
	}
	_ = scheduledAsset.UpdateSchedule(nil, &targetEnd, false)

	// Asset 2: Overdue active target (target date in April 2026 < May 2026)
	overdueDate := mustUTCDate(2026, time.April, 15)
	overdueAsset, err := asset.CreateAsset(
		core.NewAssetID(), "Solar Inverter", "", pID,
		money.New(25000, "UAH"), mustUTCDate(2026, time.February, 1),
		owner, asset.AssetStatusActive, nil, []string{"energy"},
	)
	if err != nil {
		t.Fatalf("create overdueAsset: %v", err)
	}
	_ = overdueAsset.UpdateSchedule(nil, &overdueDate, false)

	// Asset 3: Unbudgeted undated ongoing asset (burn rate contributor)
	ongoingAsset, err := asset.CreateAsset(
		core.NewAssetID(), "Weekly Groceries", "", pID,
		nil, mustUTCDate(2026, time.March, 1),
		owner, asset.AssetStatusActive, nil, []string{"home"},
	)
	if err != nil {
		t.Fatalf("create ongoingAsset: %v", err)
	}

	// Investments:
	// - $100 USD CASH on scheduledAsset in May 2026 -> 4,000 UAH cents (residual = 10,000 - 4,000 = 6,000)
	// - 10h TIME + 1,500 UAH GOODS on scheduledAsset (must NOT reduce cash residual)
	// - 9,000 UAH CASH on ongoingAsset across Mar..May (3 elapsed months => 3,000 UAH/month burn)
	investments := []asset.InvestmentRecordSummary{
		{
			ID:           core.NewInvestmentID(),
			ResourceType: "MONEY",
			Amount:       money.New(100, "USD"),
			Date:         mustUTCDate(2026, time.May, 2),
			AssetAllocations: []asset.InvestmentAllocation{
				{AssetID: scheduledAsset.ID(), SharePercentage: 100},
			},
		},
		{
			ID:           core.NewInvestmentID(),
			ResourceType: "TIME",
			TimeHours:    10.5,
			Date:         mustUTCDate(2026, time.May, 3),
			AssetAllocations: []asset.InvestmentAllocation{
				{AssetID: scheduledAsset.ID(), SharePercentage: 100},
			},
		},
		{
			ID:           core.NewInvestmentID(),
			ResourceType: "GOODS",
			Amount:       money.New(1500, "UAH"),
			Date:         mustUTCDate(2026, time.May, 4),
			AssetAllocations: []asset.InvestmentAllocation{
				{AssetID: scheduledAsset.ID(), SharePercentage: 100},
			},
		},
		{
			ID:           core.NewInvestmentID(),
			ResourceType: "MONEY",
			Amount:       money.New(9000, "UAH"),
			Date:         mustUTCDate(2026, time.April, 5),
			AssetAllocations: []asset.InvestmentAllocation{
				{AssetID: ongoingAsset.ID(), SharePercentage: 100},
			},
		},
	}

	proj := asset.ProjectDescriptor{
		ID:           pID,
		MainCurrency: "UAH",
		StartDate:    mustUTCDate(2026, time.April, 1),
		EndDate:      mustUTCDate(2026, time.July, 31),
	}
	query := asset.GetCashFlowForecastQuery{
		ProjectID: pID,
		FromMonth: "2026-04",
		ToMonth:   "2026-07",
		AsOfDate:  asOf,
	}
	ceilings := map[string]int64{
		"2026-04": 10000, // Actual 9,000 => 90% => AT_RISK
		"2026-05": 20000, // Actual 4,000 + overdue 25,000 + scheduled_May => OVER_CEILING
		"2026-06": 5000,  // HEALTHY
		"2026-07": 0,     // NO_CEILING
	}

	months, _, _, err := asset.ResolveForecastMonths(proj, []*asset.Asset{scheduledAsset, overdueAsset, ongoingAsset}, query)
	if err != nil {
		t.Fatalf("ResolveForecastMonths: %v", err)
	}

	fc := asset.CalculateProjectCashFlowForecast(
		proj,
		[]*asset.Asset{scheduledAsset, overdueAsset, ongoingAsset},
		investments,
		ceilings,
		months,
		query,
		&mockRateProvider{},
	)

	if fc.Currency != "UAH" {
		t.Errorf("expected currency UAH, got %s", fc.Currency)
	}
	if len(fc.Months) != 4 {
		t.Fatalf("expected 4 months, got %d", len(fc.Months))
	}

	// Verify exact-cent conservation for scheduledAsset: residual 6000 must equal sum of demands for scheduledAsset
	var sumScheduledForKitchen int64
	var foundOverdue bool
	for _, m := range fc.Months {
		for _, d := range m.AssetDemands {
			if d.AssetID == scheduledAsset.ID() {
				sumScheduledForKitchen += d.DemandCents
			}
			if d.AssetID == overdueAsset.ID() {
				foundOverdue = true
				if m.Month != "2026-05" || !d.IsOverdue || d.DemandCents != 25000 {
					t.Errorf("unexpected overdue demand placement: month=%s overdue=%v amt=%d", m.Month, d.IsOverdue, d.DemandCents)
				}
			}
		}
	}
	if sumScheduledForKitchen != 6000 {
		t.Errorf("exact cent conservation broken: expected 6000 across May-Jul, got %d", sumScheduledForKitchen)
	}
	if !foundOverdue {
		t.Errorf("expected overdue demand for solar inverter in 2026-05")
	}

	// Ongoing monthly burn: 9000 / 3 months (Mar, Apr, May) = 3000 UAH/month
	if fc.Summary.EstimatedOngoingMonthlyCashCents != 3000 {
		t.Errorf("expected ongoing monthly cash burn 3000, got %d", fc.Summary.EstimatedOngoingMonthlyCashCents)
	}
	// Historical April must have 0 ongoing projection, while May..Jul have 3000
	if fc.Months[0].EstimatedOngoingCashCents != 0 || fc.Months[1].EstimatedOngoingCashCents != 3000 {
		t.Errorf("expected 0 ongoing in Apr and 3000 in May, got Apr=%d May=%d", fc.Months[0].EstimatedOngoingCashCents, fc.Months[1].EstimatedOngoingCashCents)
	}
	// Committed cash in Apr must NOT include ongoing projection (9000/10000 = AT_RISK)
	if fc.Months[0].CommittedCashCents != 9000 || fc.Months[0].AlertStatus != "AT_RISK" {
		t.Errorf("expected Apr committed=9000 AT_RISK, got %d %s", fc.Months[0].CommittedCashCents, fc.Months[0].AlertStatus)
	}
	if fc.Months[1].AlertStatus != "OVER_CEILING" {
		t.Errorf("expected May OVER_CEILING, got %s", fc.Months[1].AlertStatus)
	}
	if fc.Months[3].AlertStatus != "NO_CEILING" {
		t.Errorf("expected Jul NO_CEILING, got %s", fc.Months[3].AlertStatus)
	}
	if fc.Summary.TotalTimeHours != 10.5 || fc.Summary.TotalGoodsCents != 1500 {
		t.Errorf("expected Time=10.5 Goods=1500, got Time=%v Goods=%d", fc.Summary.TotalTimeHours, fc.Summary.TotalGoodsCents)
	}

	// Tag filtering limits assets to "home"
	homeFC := asset.CalculateProjectCashFlowForecast(
		proj,
		[]*asset.Asset{scheduledAsset, overdueAsset, ongoingAsset},
		investments,
		ceilings,
		months,
		asset.GetCashFlowForecastQuery{ProjectID: pID, FromMonth: "2026-04", ToMonth: "2026-07", AsOfDate: asOf, Tag: "home"},
		&mockRateProvider{},
	)
	if len(homeFC.Assets) != 2 {
		t.Errorf("expected 2 home-tagged assets, got %d", len(homeFC.Assets))
	}
}

func TestResolveForecastMonthsErrors(t *testing.T) {
	proj := asset.ProjectDescriptor{ID: core.NewProjectID(), MainCurrency: "UAH"}
	if _, _, _, err := asset.ResolveForecastMonths(proj, nil, asset.GetCashFlowForecastQuery{FromMonth: "bad"}); err == nil {
		t.Errorf("expected error for invalid from_month")
	}
	if _, _, _, err := asset.ResolveForecastMonths(proj, nil, asset.GetCashFlowForecastQuery{ToMonth: "2026/05"}); err == nil {
		t.Errorf("expected error for invalid to_month")
	}
	if _, _, _, err := asset.ResolveForecastMonths(proj, nil, asset.GetCashFlowForecastQuery{FromMonth: "2026-06", ToMonth: "2026-05"}); err == nil {
		t.Errorf("expected error for to_month < from_month")
	}
}

func TestServiceMultiCurrencyScheduleAndForecast(t *testing.T) {
	db := &mockDb{}
	ownerID := core.NewPersonID()
	pID := core.NewProjectID()
	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, ownerID)
	ownerRef := asset.OwnerRef{PersonID: ownerID, DisplayName: "Max"}

	child, _ := asset.CreateAsset(core.NewAssetID(), "Child", "", pID, money.New(500, "UAH"), mustUTCDate(2026, time.January, 1), ownerRef, asset.AssetStatusActive, nil, nil)
	childTarget := mustUTCDate(2026, time.August, 1)
	_ = child.UpdateSchedule(nil, &childTarget, false)

	parent, _ := asset.CreateAsset(core.NewAssetID(), "Parent", "", pID, money.New(2000, "UAH"), mustUTCDate(2026, time.January, 1), ownerRef, asset.AssetStatusActive, nil, nil)
	parent.AssignValuation(nil, nil, []asset.ChildAssetLink{{ChildID: child.ID(), ChildName: "Child", SharePercentage: 100}}, nil)

	col := asset.NewCollection(2)
	col.Add(parent)
	col.Add(child)

	repo := &mockLookupAssetRepo{
		mockAssetRepo: mockAssetRepo{col: col, asset: parent},
		byID: map[core.AssetID]*asset.Asset{
			parent.ID(): parent,
			child.ID():  child,
		},
	}
	peopleSvc := &mockPeopleService{owner: ownerRef}
	projRepo := &mockProjectRepo{
		project: asset.ProjectDescriptor{
			ID:           pID,
			MainCurrency: "UAH",
			StartDate:    mustUTCDate(2026, time.May, 1),
			EndDate:      mustUTCDate(2026, time.August, 31),
		},
		ceilings: map[string]int64{"2026-05": 10000},
	}
	invSrc := &mockInvestmentSource{
		projectSummary: []asset.InvestmentRecordSummary{
			{
				ID:           core.NewInvestmentID(),
				ResourceType: "MONEY",
				Amount:       money.New(10, "USD"), // 400 UAH in child -> rolls up 100% into parent!
				Date:         mustUTCDate(2026, time.May, 2),
				AssetAllocations: []asset.InvestmentAllocation{
					{AssetID: child.ID(), SharePercentage: 100},
				},
			},
		},
	}

	svc := newTestService(db, repo, peopleSvc, projRepo, invSrc, &mockRateProvider{})

	// Find normalizes mixed-currency direct & rolled-up costs in UAH
	foundCol, err := svc.Find(authedCtx, asset.CollectionFilter{ProjectID: pID})
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if foundCol.Elements()[0].TotalCost() == nil || foundCol.Elements()[0].TotalCost().Amount() != 400 || foundCol.Elements()[0].TotalCost().Currency().Code != "UAH" {
		t.Errorf("expected parent TotalCost rolled up to 400 UAH, got %v", foundCol.Elements()[0].TotalCost())
	}

	// FindOne normalizes single asset via project hierarchy
	one, err := svc.FindOne(authedCtx, asset.Filter{ID: parent.ID(), ProjectID: pID})
	if err != nil || one == nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	// Create unbudgeted asset succeeds, while WithPayment without positive price fails
	created, err := svc.Create(authedCtx, asset.CreateAssetCommand{
		Name:      "Unbudgeted",
		ProjectID: pID,
	})
	if err != nil || created.HasTargetBudget() {
		t.Errorf("expected unbudgeted asset create to succeed, got asset=%v err=%v", created, err)
	}
	if _, err := svc.Create(authedCtx, asset.CreateAssetCommand{
		Name:        "Needs Payment",
		ProjectID:   pID,
		WithPayment: true,
	}); err == nil {
		t.Errorf("expected error when WithPayment=true and no target price")
	}

	// UpdateSchedule returns non-blocking warning when parent target precedes child target
	earlierParentTarget := mustUTCDate(2026, time.June, 1)
	_, warnings, err := svc.UpdateSchedule(authedCtx, asset.RescheduleAssetCommand{
		AssetID:              parent.ID(),
		ProjectID:            pID,
		TargetCompletionDate: &earlierParentTarget,
	})
	if err != nil {
		t.Fatalf("UpdateSchedule failed: %v", err)
	}
	if len(warnings) != 1 || !strings.HasPrefix(warnings[0], "child_target_exceeds_parent:") {
		t.Errorf("expected child_target_exceeds_parent warning, got %v", warnings)
	}

	// GetCashFlowForecast validation & error branches
	if _, err := svc.GetCashFlowForecast(context.Background(), asset.GetCashFlowForecastQuery{ProjectID: pID}); err == nil {
		t.Errorf("expected unauthorized error")
	}
	if _, err := svc.GetCashFlowForecast(authedCtx, asset.GetCashFlowForecastQuery{}); err == nil {
		t.Errorf("expected validation error for missing project id")
	}
	invSrc.listErr = errors.New("inv boom")
	if _, err := svc.GetCashFlowForecast(authedCtx, asset.GetCashFlowForecastQuery{ProjectID: pID}); err == nil {
		t.Errorf("expected investment source error propagation")
	}
	invSrc.listErr = nil

	fc, err := svc.GetCashFlowForecast(authedCtx, asset.GetCashFlowForecastQuery{
		ProjectID: pID,
		FromMonth: "2026-05",
		ToMonth:   "2026-08",
		AsOfDate:  mustUTCDate(2026, time.May, 10),
	})
	if err != nil || fc == nil || len(fc.Months) != 4 {
		t.Fatalf("expected 4-month forecast, got fc=%v err=%v", fc, err)
	}

	// Cover additional edge cases cleanly
	if parent.DirectCostHomeAmount() != 0 {
		t.Errorf("expected 0 DirectCostHomeAmount for group parent, got %d", parent.DirectCostHomeAmount())
	}
	if child.DirectCostHomeAmount() != 400 {
		t.Errorf("expected 400 DirectCostHomeAmount for USD child, got %d", child.DirectCostHomeAmount())
	}

	// Child UpdateSchedule triggers parent warning via target.Parents()
	_, childWarnings, err := svc.UpdateSchedule(authedCtx, asset.RescheduleAssetCommand{
		AssetID:   child.ID(),
		ProjectID: pID,
	})
	if err == nil && len(childWarnings) == 0 {
		// Assign parent link on child and verify parent warning fires
		child.AssignValuation(child.DirectCost(), child.TotalCost(), nil, []asset.ParentAssetLink{{ParentID: parent.ID(), ParentName: "Parent", SharePercentage: 100}})
		_, childWarnings, err = svc.UpdateSchedule(authedCtx, asset.RescheduleAssetCommand{
			AssetID:              child.ID(),
			ProjectID:            pID,
			TargetCompletionDate: &childTarget,
		})
		if err != nil || len(childWarnings) == 0 {
			t.Errorf("expected parent warning when updating child target past parent target, got err=%v warnings=%v", err, childWarnings)
		}
	}

	// Unauthorized & not found UpdateSchedule branches
	if _, _, err := svc.UpdateSchedule(context.Background(), asset.RescheduleAssetCommand{AssetID: child.ID(), ProjectID: pID}); err == nil {
		t.Errorf("expected unauthorized error for UpdateSchedule")
	}
	if _, _, err := svc.UpdateSchedule(authedCtx, asset.RescheduleAssetCommand{AssetID: core.NewAssetID(), ProjectID: pID}); err == nil {
		t.Errorf("expected error when UpdateSchedule asset missing")
	}

	// Default month resolution, COMPLETED asset, ONGOING_RUN_RATE with EstimatedCompletion, and unallocated investments
	doneDate := mustUTCDate(2026, time.March, 20)
	completedAsset, _ := asset.CreateAsset(core.NewAssetID(), "Done", "", pID, money.New(1000, "UAH"), mustUTCDate(2026, time.March, 1), ownerRef, asset.AssetStatusCompleted, &doneDate, nil)
	budgetedOngoing, _ := asset.CreateAsset(core.NewAssetID(), "Save Fund", "", pID, money.New(12000, "UAH"), mustUTCDate(2026, time.April, 1), ownerRef, asset.AssetStatusActive, nil, nil)
	farTarget := mustUTCDate(2029, time.January, 1)
	farAsset, _ := asset.CreateAsset(core.NewAssetID(), "Far Goal", "", pID, money.New(1000, "UAH"), mustUTCDate(2026, time.May, 1), ownerRef, asset.AssetStatusActive, nil, nil)
	_ = farAsset.UpdateSchedule(nil, &farTarget, false)

	defProj := asset.ProjectDescriptor{ID: pID, MainCurrency: "UAH", StartDate: mustUTCDate(2026, time.May, 1)}
	defMonths, _, _, err := asset.ResolveForecastMonths(defProj, []*asset.Asset{completedAsset, budgetedOngoing, farAsset}, asset.GetCashFlowForecastQuery{AsOfDate: mustUTCDate(2026, time.May, 10)})
	if err != nil || len(defMonths) != 24 {
		t.Fatalf("expected capped 24 default months, got %d (err=%v)", len(defMonths), err)
	}
	edgeFC := asset.CalculateProjectCashFlowForecast(
		defProj,
		[]*asset.Asset{completedAsset, budgetedOngoing, farAsset},
		[]asset.InvestmentRecordSummary{
			{ID: core.NewInvestmentID(), ResourceType: "MONEY", Amount: money.New(4000, "UAH"), Date: mustUTCDate(2026, time.May, 1), AssetAllocations: []asset.InvestmentAllocation{{AssetID: budgetedOngoing.ID(), SharePercentage: 100}}},
			{ID: core.NewInvestmentID(), ResourceType: "MONEY", Amount: money.New(500, "UAH"), Date: mustUTCDate(2026, time.May, 2)},
		},
		nil,
		defMonths[:6],
		asset.GetCashFlowForecastQuery{ProjectID: pID, AsOfDate: mustUTCDate(2026, time.May, 10)},
		nil,
	)
	var sawCompleted, sawOngoingETA bool
	for _, it := range edgeFC.Assets {
		if it.AssetID == completedAsset.ID() && it.ForecastMode == "COMPLETED" {
			sawCompleted = true
		}
		if it.AssetID == budgetedOngoing.ID() && it.ForecastMode == "ONGOING_RUN_RATE" && it.EstimatedCompletion != nil {
			sawOngoingETA = true
		}
	}
	if !sawCompleted || !sawOngoingETA {
		t.Errorf("expected COMPLETED (%v) and ONGOING_RUN_RATE ETA (%v)", sawCompleted, sawOngoingETA)
	}
}
