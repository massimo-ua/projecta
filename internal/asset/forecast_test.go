package asset_test

import (
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
)

func mustUTCDate(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

type fixedRateConverter struct{}

func (fixedRateConverter) ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error) {
	if amount == nil {
		return money.New(0, targetCurrency), nil
	}
	if amount.Currency().Code == targetCurrency {
		return amount, nil
	}
	return money.New(amount.Amount()*40, targetCurrency), nil
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

	months, _, _, err := asset.ResolveForecastMonths(proj, []*asset.Asset{scheduledAsset, overdueAsset, ongoingAsset}, investments, query)
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
		fixedRateConverter{},
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
		fixedRateConverter{},
	)
	if len(homeFC.Assets) != 2 {
		t.Errorf("expected 2 home-tagged assets, got %d", len(homeFC.Assets))
	}
}

func TestResolveForecastMonthsErrors(t *testing.T) {
	proj := asset.ProjectDescriptor{ID: core.NewProjectID(), MainCurrency: "UAH"}
	if _, _, _, err := asset.ResolveForecastMonths(proj, nil, nil, asset.GetCashFlowForecastQuery{FromMonth: "bad"}); err == nil {
		t.Errorf("expected error for invalid from_month")
	}
	if _, _, _, err := asset.ResolveForecastMonths(proj, nil, nil, asset.GetCashFlowForecastQuery{ToMonth: "2026/05"}); err == nil {
		t.Errorf("expected error for invalid to_month")
	}
	if _, _, _, err := asset.ResolveForecastMonths(proj, nil, nil, asset.GetCashFlowForecastQuery{FromMonth: "2026-06", ToMonth: "2026-05"}); err == nil {
		t.Errorf("expected error for to_month < from_month")
	}
}

func TestResolveForecastMonthsIncludesPreStartAndMultiYearInvestments(t *testing.T) {
	pID := core.NewProjectID()
	proj := asset.ProjectDescriptor{
		ID:           pID,
		MainCurrency: "UAH",
		StartDate:    mustUTCDate(2024, time.December, 27),
		EndDate:      mustUTCDate(2029, time.March, 1),
	}
	invs := []asset.InvestmentRecordSummary{
		{
			ID:           core.NewInvestmentID(),
			ResourceType: "MONEY",
			Amount:       money.New(5000, "UAH"),
			Date:         mustUTCDate(2024, time.October, 15), // Before project.StartDate
		},
		{
			ID:           core.NewInvestmentID(),
			ResourceType: "MONEY",
			Amount:       money.New(12000, "UAH"),
			Date:         mustUTCDate(2027, time.January, 10), // Past 24 months from start
		},
	}

	months, _, _, err := asset.ResolveForecastMonths(proj, nil, invs, asset.GetCashFlowForecastQuery{
		ProjectID: pID,
		AsOfDate:  mustUTCDate(2026, time.November, 22),
	})
	if err != nil {
		t.Fatalf("ResolveForecastMonths failed: %v", err)
	}
	if len(months) == 0 {
		t.Fatalf("expected non-empty months")
	}
	if months[0] != "2024-10" {
		t.Errorf("expected from_month 2024-10 to include pre-start investment, got %s", months[0])
	}
	if months[len(months)-1] != "2029-03" {
		t.Errorf("expected to_month 2029-03 to cover multi-year project, got %s", months[len(months)-1])
	}

	// Safety upper bound still caps absurdly long spans at 120 months
	veryLongProj := asset.ProjectDescriptor{
		ID:           pID,
		MainCurrency: "UAH",
		StartDate:    mustUTCDate(2020, time.January, 1),
		EndDate:      mustUTCDate(2050, time.January, 1),
	}
	capped, _, _, err := asset.ResolveForecastMonths(veryLongProj, nil, nil, asset.GetCashFlowForecastQuery{
		ProjectID: pID,
		AsOfDate:  mustUTCDate(2026, time.November, 22),
	})
	if err != nil {
		t.Fatalf("ResolveForecastMonths failed: %v", err)
	}
	if len(capped) != 120 {
		t.Errorf("expected safety cap of 120 months, got %d", len(capped))
	}

	// Ancient outlier investment must not push current/upcoming months out of the 120-month cap
	outlierInvs := []asset.InvestmentRecordSummary{
		{
			ID:           core.NewInvestmentID(),
			ResourceType: "MONEY",
			Amount:       money.New(100, "UAH"),
			Date:         mustUTCDate(2014, time.January, 10),
		},
		{
			ID:           core.NewInvestmentID(),
			ResourceType: "MONEY",
			Amount:       money.New(5000, "UAH"),
			Date:         mustUTCDate(2026, time.October, 10),
		},
	}
	anchored, _, _, err := asset.ResolveForecastMonths(proj, nil, outlierInvs, asset.GetCashFlowForecastQuery{
		ProjectID: pID,
		AsOfDate:  mustUTCDate(2026, time.November, 22),
	})
	if err != nil {
		t.Fatalf("ResolveForecastMonths with outlier failed: %v", err)
	}
	if len(anchored) != 120 {
		t.Fatalf("expected 120 capped months for outlier query, got %d", len(anchored))
	}
	if anchored[len(anchored)-1] != "2027-04" {
		t.Errorf("expected outlier-anchored window to preserve currentMonth+5 (2027-04), got %s", anchored[len(anchored)-1])
	}
}

func TestCalculateProjectCashFlowForecast_TagFilterMatchesInvestmentTags(t *testing.T) {
	pID := core.NewProjectID()
	owner := asset.OwnerRef{PersonID: core.NewPersonID(), DisplayName: "Max"}
	asOf := mustUTCDate(2026, time.May, 10)

	untaggedAsset, _ := asset.CreateAsset(
		core.NewAssetID(), "General", "", pID,
		nil, mustUTCDate(2026, time.January, 1),
		owner, asset.AssetStatusActive, nil, []string{"other"},
	)

	invs := []asset.InvestmentRecordSummary{
		{
			ID:           core.NewInvestmentID(),
			ResourceType: "MONEY",
			Amount:       money.New(10000, "UAH"),
			Date:         mustUTCDate(2026, time.April, 10),
			Tags:         []string{"освіта"},
			AssetAllocations: []asset.InvestmentAllocation{
				{AssetID: untaggedAsset.ID(), SharePercentage: 100},
			},
		},
		{
			ID:           core.NewInvestmentID(),
			ResourceType: "MONEY",
			Amount:       money.New(5000, "UAH"),
			Date:         mustUTCDate(2026, time.April, 15),
			Tags:         []string{"освіта"},
		},
		{
			ID:           core.NewInvestmentID(),
			ResourceType: "MONEY",
			Amount:       money.New(9999, "UAH"),
			Date:         mustUTCDate(2026, time.April, 20),
			Tags:         []string{"unrelated"},
			AssetAllocations: []asset.InvestmentAllocation{
				{AssetID: untaggedAsset.ID(), SharePercentage: 100},
			},
		},
	}

	proj := asset.ProjectDescriptor{ID: pID, MainCurrency: "UAH", StartDate: mustUTCDate(2026, time.April, 1)}
	fc := asset.CalculateProjectCashFlowForecast(
		proj,
		[]*asset.Asset{untaggedAsset},
		invs,
		map[string]int64{"2026-04": 20000},
		[]string{"2026-04"},
		asset.GetCashFlowForecastQuery{ProjectID: pID, AsOfDate: asOf, Tag: "#освіта"},
		nil,
	)

	if fc.Months[0].Actual.CashCents != 15000 {
		t.Errorf("expected tag-filtered actual cash 15000 (allocated + unallocated investment tags), got %d", fc.Months[0].Actual.CashCents)
	}
	if fc.Summary.TotalActualCashCents != 15000 {
		t.Errorf("expected summary actual cash 15000, got %d", fc.Summary.TotalActualCashCents)
	}
}

func TestCalculateProjectCashFlowForecast_CompletedAndOngoingETA(t *testing.T) {
	pID := core.NewProjectID()
	ownerRef := asset.OwnerRef{PersonID: core.NewPersonID(), DisplayName: "Max"}

	doneDate := mustUTCDate(2026, time.March, 20)
	completedAsset, _ := asset.CreateAsset(core.NewAssetID(), "Done", "", pID, money.New(1000, "UAH"), mustUTCDate(2026, time.March, 1), ownerRef, asset.AssetStatusCompleted, &doneDate, nil)
	budgetedOngoing, _ := asset.CreateAsset(core.NewAssetID(), "Save Fund", "", pID, money.New(12000, "UAH"), mustUTCDate(2026, time.April, 1), ownerRef, asset.AssetStatusActive, nil, nil)
	farTarget := mustUTCDate(2029, time.January, 1)
	farAsset, _ := asset.CreateAsset(core.NewAssetID(), "Far Goal", "", pID, money.New(1000, "UAH"), mustUTCDate(2026, time.May, 1), ownerRef, asset.AssetStatusActive, nil, nil)
	_ = farAsset.UpdateSchedule(nil, &farTarget, false)

	defProj := asset.ProjectDescriptor{ID: pID, MainCurrency: "UAH", StartDate: mustUTCDate(2026, time.May, 1)}
	defMonths, _, _, err := asset.ResolveForecastMonths(defProj, []*asset.Asset{completedAsset, budgetedOngoing, farAsset}, nil, asset.GetCashFlowForecastQuery{AsOfDate: mustUTCDate(2026, time.May, 10)})
	if err != nil || len(defMonths) != 33 {
		t.Fatalf("expected 33 default months (2026-05..2029-01), got %d (err=%v)", len(defMonths), err)
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
