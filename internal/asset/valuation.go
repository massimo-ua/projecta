package asset

import (
	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
)

// CurrencyConverter is a narrow domain interface for currency conversions.
type CurrencyConverter interface {
	Convert(source, target currency.Currency) (currency.Currency, error)
}

// BudgetProgress represents calculated progress toward an asset's target budget.
type BudgetProgress struct {
	TargetHomeAmount   *int64
	ProgressPercentage *float64
	IsOverBudget       bool
}

// CalculateBudgetProgress computes the budget progress and home currency equivalent for an asset.
func CalculateBudgetProgress(
	totalCost *money.Money,
	targetPrice *money.Money,
	homeCurrency string,
	converter CurrencyConverter,
) BudgetProgress {
	if targetPrice == nil || targetPrice.Amount() <= 0 {
		return BudgetProgress{}
	}

	targetAmt := targetPrice.Amount()
	targetCurr := targetPrice.Currency().Code
	targetHomeVal := targetAmt

	if converter != nil && targetCurr != homeCurrency && targetAmt > 0 {
		converted, err := converter.Convert(
			currency.NewCurrency(targetAmt, targetCurr),
			currency.NewCurrency(0, homeCurrency),
		)
		if err == nil {
			targetHomeVal = converted.Amount
		}
	}

	if targetHomeVal <= 0 {
		return BudgetProgress{TargetHomeAmount: &targetHomeVal}
	}

	var totalHomeVal int64
	if totalCost != nil && totalCost.Amount() > 0 {
		totalHomeVal = totalCost.Amount()
		totalCurr := totalCost.Currency().Code
		if converter != nil && totalCurr != homeCurrency {
			converted, err := converter.Convert(
				currency.NewCurrency(totalHomeVal, totalCurr),
				currency.NewCurrency(0, homeCurrency),
			)
			if err == nil {
				totalHomeVal = converted.Amount
			}
		}
	}

	pct := (float64(totalHomeVal) / float64(targetHomeVal)) * 100.0
	isOver := pct > 100.0

	return BudgetProgress{
		TargetHomeAmount:   &targetHomeVal,
		ProgressPercentage: &pct,
		IsOverBudget:       isOver,
	}
}

// ProjectTotals represents aggregated financial totals for a project.
type ProjectTotals struct {
	TotalInvested  int64
	TotalValuation int64
	Currency       string
	HasActivity    bool
}

// Balance returns the remaining unallocated project balance (TotalInvested - TotalValuation).
func (t ProjectTotals) Balance() int64 {
	return t.TotalInvested - t.TotalValuation
}

// CalculateProjectTotals aggregates raw investments and root assets into project totals.
func CalculateProjectTotals(
	investmentAmounts []*money.Money,
	rootAssets []*Asset,
	homeCurrency string,
	converter CurrencyConverter,
) (ProjectTotals, error) {
	if homeCurrency == "" {
		homeCurrency = "UAH"
	}

	var totalInvested int64
	for _, inv := range investmentAmounts {
		if inv == nil || inv.Amount() == 0 {
			continue
		}
		amt := inv.Amount()
		if converter != nil && inv.Currency().Code != homeCurrency {
			converted, err := converter.Convert(
				currency.NewCurrency(inv.Amount(), inv.Currency().Code),
				currency.NewCurrency(0, homeCurrency),
			)
			if err != nil {
				return ProjectTotals{}, err
			}
			amt = converted.Amount
		}
		totalInvested += amt
	}

	var totalValuation int64
	for _, a := range rootAssets {
		if a == nil {
			continue
		}
		tc := a.TotalCost()
		if tc == nil || tc.Amount() == 0 {
			continue
		}
		amt := tc.Amount()
		if converter != nil && tc.Currency().Code != homeCurrency {
			converted, err := converter.Convert(
				currency.NewCurrency(tc.Amount(), tc.Currency().Code),
				currency.NewCurrency(0, homeCurrency),
			)
			if err != nil {
				return ProjectTotals{}, err
			}
			amt = converted.Amount
		}
		totalValuation += amt
	}

	return ProjectTotals{
		TotalInvested:  totalInvested,
		TotalValuation: totalValuation,
		Currency:       homeCurrency,
	}, nil
}
