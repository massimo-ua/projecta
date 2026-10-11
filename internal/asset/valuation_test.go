package asset_test

import (
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
)

type mockConverter struct {
	rate float64
}

func (m *mockConverter) Convert(source, target currency.Currency) (currency.Currency, error) {
	// Simple mock: if source != target, multiply by m.rate
	amt := int64(float64(source.Amount) * m.rate)
	return currency.NewCurrency(amt, target.Code), nil
}

func TestCalculateBudgetProgress_SameCurrency(t *testing.T) {
	totalCost := money.New(5000, "EUR")
	targetPrice := money.New(10000, "EUR")

	bp := asset.CalculateBudgetProgress(totalCost, targetPrice, "EUR", nil)
	if bp.TargetHomeAmount == nil || *bp.TargetHomeAmount != 10000 {
		t.Errorf("expected TargetHomeAmount 10000, got %v", bp.TargetHomeAmount)
	}
	if bp.ProgressPercentage == nil || *bp.ProgressPercentage != 50.0 {
		t.Errorf("expected ProgressPercentage 50.0, got %v", bp.ProgressPercentage)
	}
	if bp.IsOverBudget {
		t.Errorf("expected IsOverBudget false")
	}
}

func TestCalculateBudgetProgress_OverbudgetWithConversion(t *testing.T) {
	converter := &mockConverter{rate: 40.0} // 1 EUR = 40 UAH
	totalCost := money.New(450000, "UAH")  // 4500.00 UAH
	targetPrice := money.New(10000, "EUR") // 100.00 EUR -> 4000.00 UAH

	bp := asset.CalculateBudgetProgress(totalCost, targetPrice, "UAH", converter)
	if bp.TargetHomeAmount == nil || *bp.TargetHomeAmount != 400000 {
		t.Errorf("expected TargetHomeAmount 400000, got %v", bp.TargetHomeAmount)
	}
	if bp.ProgressPercentage == nil || *bp.ProgressPercentage != 112.5 {
		t.Errorf("expected ProgressPercentage 112.5, got %v", bp.ProgressPercentage)
	}
	if !bp.IsOverBudget {
		t.Errorf("expected IsOverBudget true")
	}
}

func TestCalculateProjectTotals(t *testing.T) {
	converter := &mockConverter{rate: 40.0}
	inv1 := money.New(1000, "UAH")
	inv2 := money.New(100, "EUR") // converted: 4000 UAH

	owner := &projecta.Owner{PersonID: uuid.New(), DisplayName: "User"}
	proj, _ := projecta.NewProject(uuid.New(), "P", "", owner, time.Now(), time.Now())
	rootA := asset.NewAsset(uuid.New(), "RootA", "", proj, nil, money.New(0, "UAH"), time.Now(), owner)
	rootA.AssignValuation(money.New(5000, "UAH"), money.New(5000, "UAH"), nil, nil)

	totals, err := asset.CalculateProjectTotals(
		[]*money.Money{inv1, inv2},
		[]*asset.Asset{rootA},
		"UAH",
		converter,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if totals.TotalInvested != 5000 {
		t.Errorf("expected TotalInvested 5000 (1000 + 4000), got %d", totals.TotalInvested)
	}
	if totals.TotalValuation != 5000 {
		t.Errorf("expected TotalValuation 5000, got %d", totals.TotalValuation)
	}
}
