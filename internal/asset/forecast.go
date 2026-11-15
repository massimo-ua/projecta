package asset

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

func errInvalidForecastMonth(field string) error {
	return exceptions.NewValidationException(fmt.Sprintf("%s must be formatted as YYYY-MM", field), nil)
}

func errInvalidForecastRange() error {
	return exceptions.NewValidationException("to_month must be greater than or equal to from_month", nil)
}

const (
	defaultTrailingWindowMonths = 6
	maxForecastWindowMonths     = 24
	atRiskThresholdRatio        = 0.85
)

// ResourceBreakdown captures money, time, and goods contributions.
type ResourceBreakdown struct {
	CashCents  int64
	TimeHours  float64
	GoodsCents int64
}

// MonthlyBurnRate represents trailing average monthly consumption rates for an asset.
type MonthlyBurnRate struct {
	CashCentsPerMonth  int64
	TimeHoursPerMonth  float64
	GoodsCentsPerMonth int64
}

// AssetDemandItem attributes a scheduled cash demand slice in a month to a specific asset.
type AssetDemandItem struct {
	AssetID     core.AssetID
	AssetName   string
	DemandCents int64
	IsOverdue   bool
}

// MonthCashFlow represents historical actuals, scheduled capital demand, and ongoing burn for a calendar month.
type MonthCashFlow struct {
	Month                     string
	TagBudgetCeilingCents     int64
	Actual                    ResourceBreakdown
	ScheduledDemandCents      int64
	EstimatedOngoingCashCents int64
	CommittedCashCents        int64
	CashUtilizationPct        float64
	AlertStatus               string
	AssetDemands              []AssetDemandItem
}

// AssetForecastItem summarizes per-asset true cost, residual target, burn rate, and forecast status.
type AssetForecastItem struct {
	AssetID                core.AssetID
	AssetName              string
	Status                 AssetStatus
	StartDate              time.Time
	TargetCompletionDate   *time.Time
	CompletedDate          *time.Time
	HasTargetBudget        bool
	TargetHomeCents        int64
	ActualToDate           ResourceBreakdown
	RolledUpTotalCashCents int64
	ResidualTargetCents    int64
	BurnRateMonthly        MonthlyBurnRate
	ForecastMode           string
	EstimatedCompletion    *time.Time
}

// CashFlowSummary aggregates project-level commitments, actuals, and risk indicators.
type CashFlowSummary struct {
	TotalTargetCents                 int64
	TotalActualCashCents             int64
	TotalRemainingScheduledCents     int64
	TotalTimeHours                   float64
	TotalGoodsCents                  int64
	EstimatedOngoingMonthlyCashCents int64
	MonthlyBudgetCeilingAvgCents     int64
	PeakMonth                        string
	PeakCommittedCashCents           int64
	AlertMonthsCount                 int
}

// CashFlowForecast is the complete cash flow and true-cost plan for a project.
type CashFlowForecast struct {
	ProjectID            core.ProjectID
	Currency             string
	FromMonth            string
	ToMonth              string
	TrailingWindowMonths int
	Summary              CashFlowSummary
	Months               []MonthCashFlow
	Assets               []AssetForecastItem
}

// NormalizeProjectInvestments converts project investments into homeCurrency and aggregates
// per-asset direct monetary totals (MONEY + GOODS) as well as resource-typed cash-only totals.
func NormalizeProjectInvestments(
	investments []InvestmentRecordSummary,
	homeCurrency string,
	converter CurrencyConverter,
) (directAll map[core.AssetID]*money.Money, directCash map[core.AssetID]*money.Money) {
	if homeCurrency == "" {
		homeCurrency = "UAH"
	}
	directAllCents := make(map[core.AssetID]int64)
	directCashCents := make(map[core.AssetID]int64)

	for _, inv := range investments {
		homeAmt := convertToHomeCents(inv.Amount, homeCurrency, converter)
		if homeAmt == 0 || len(inv.AssetAllocations) == 0 {
			continue
		}
		resType := normalizeResourceType(inv.ResourceType)
		for _, alloc := range inv.AssetAllocations {
			if alloc.AssetID.IsNil() || alloc.SharePercentage <= 0 {
				continue
			}
			shareAmt := int64(math.Round(float64(homeAmt) * alloc.SharePercentage / 100.0))
			directAllCents[alloc.AssetID] += shareAmt
			if resType == "MONEY" {
				directCashCents[alloc.AssetID] += shareAmt
			}
		}
	}

	directAll = make(map[core.AssetID]*money.Money, len(directAllCents))
	for id, amt := range directAllCents {
		directAll[id] = money.New(amt, homeCurrency)
	}
	directCash = make(map[core.AssetID]*money.Money, len(directCashCents))
	for id, amt := range directCashCents {
		directCash[id] = money.New(amt, homeCurrency)
	}
	return directAll, directCash
}

func convertToHomeCents(amount *money.Money, homeCurrency string, converter CurrencyConverter) int64 {
	if amount == nil || amount.Amount() == 0 {
		return 0
	}
	amt := amount.Amount()
	if amount.Currency() != nil && amount.Currency().Code != "" && amount.Currency().Code != homeCurrency && converter != nil {
		if converted, err := converter.ConvertMoney(amount, homeCurrency); err == nil && converted != nil {
			amt = converted.Amount()
		}
	}
	return amt
}

func normalizeResourceType(rt string) string {
	switch strings.ToUpper(strings.TrimSpace(rt)) {
	case "TIME":
		return "TIME"
	case "GOODS":
		return "GOODS"
	default:
		return "MONEY"
	}
}

func utcMonthStart(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func parseMonthStart(monthStr string) (time.Time, bool) {
	if strings.TrimSpace(monthStr) == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse("2006-01", strings.TrimSpace(monthStr))
	if err != nil {
		return time.Time{}, false
	}
	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC), true
}

func formatMonthKey(t time.Time) string {
	return t.UTC().Format("2006-01")
}

func monthsBetweenInclusive(startMonth, endMonth time.Time) int {
	if endMonth.Before(startMonth) {
		return 0
	}
	return (endMonth.Year()-startMonth.Year())*12 + int(endMonth.Month()-startMonth.Month()) + 1
}

// ResolveForecastMonths determines the ordered list of YYYY-MM strings for a cash flow query.
func ResolveForecastMonths(
	project ProjectDescriptor,
	assets []*Asset,
	query GetCashFlowForecastQuery,
) ([]string, time.Time, time.Time, error) {
	asOf := query.AsOfDate
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	} else {
		asOf = asOf.UTC()
	}
	currentMonth := utcMonthStart(asOf)

	fromTime, hasFrom := parseMonthStart(query.FromMonth)
	if query.FromMonth != "" && !hasFrom {
		return nil, time.Time{}, time.Time{}, errInvalidForecastMonth("from_month")
	}
	if !hasFrom {
		if !project.StartDate.IsZero() {
			fromTime = utcMonthStart(project.StartDate)
		} else {
			fromTime = currentMonth
		}
	}

	toTime, hasTo := parseMonthStart(query.ToMonth)
	if query.ToMonth != "" && !hasTo {
		return nil, time.Time{}, time.Time{}, errInvalidForecastMonth("to_month")
	}
	if !hasTo {
		toTime = fromTime.AddDate(0, 11, 0)
		if toTime.Before(currentMonth.AddDate(0, 5, 0)) {
			toTime = currentMonth.AddDate(0, 5, 0)
		}
		if !project.EndDate.IsZero() {
			projEnd := utcMonthStart(project.EndDate)
			if projEnd.After(toTime) {
				toTime = projEnd
			}
		}
		for _, a := range assets {
			if a == nil || a.TargetCompletionDate() == nil {
				continue
			}
			td := utcMonthStart(*a.TargetCompletionDate())
			if td.After(toTime) {
				toTime = td
			}
		}
		maxTo := fromTime.AddDate(0, maxForecastWindowMonths-1, 0)
		if toTime.After(maxTo) {
			toTime = maxTo
		}
	}

	if toTime.Before(fromTime) {
		return nil, time.Time{}, time.Time{}, errInvalidForecastRange()
	}

	count := monthsBetweenInclusive(fromTime, toTime)
	if count > maxForecastWindowMonths*2 {
		count = maxForecastWindowMonths * 2
		toTime = fromTime.AddDate(0, count-1, 0)
	}

	months := make([]string, 0, count)
	for cur := fromTime; !cur.After(toTime); cur = cur.AddDate(0, 1, 0) {
		months = append(months, formatMonthKey(cur))
	}
	return months, fromTime, toTime, nil
}

// CalculateProjectCashFlowForecast computes the deterministic true-cost & cash-flow forecast.
func CalculateProjectCashFlowForecast(
	project ProjectDescriptor,
	assets []*Asset,
	investments []InvestmentRecordSummary,
	ceilings map[string]int64,
	months []string,
	query GetCashFlowForecastQuery,
	converter CurrencyConverter,
) *CashFlowForecast {
	homeCurrency := project.MainCurrency
	if homeCurrency == "" {
		homeCurrency = "UAH"
	}
	asOf := query.AsOfDate
	if asOf.IsZero() {
		asOf = time.Now().UTC()
	} else {
		asOf = asOf.UTC()
	}
	asOfDay := utcCalendarDay(asOf)
	currentMonth := utcMonthStart(asOf)
	currentMonthKey := formatMonthKey(currentMonth)
	trailingStartMonth := currentMonth.AddDate(0, -(defaultTrailingWindowMonths - 1), 0)

	// Filter assets if tag filter provided
	filteredAssets, includedAssetIDs := filterAssetsByTag(assets, query.Tag)

	// Build hierarchy across ALL project assets so parent/child shares roll up accurately.
	hierarchy := NewAssetHierarchy()
	for _, a := range assets {
		if a == nil {
			continue
		}
		hierarchy.AddAsset(a)
	}
	for _, a := range assets {
		if a == nil {
			continue
		}
		for _, ch := range a.Children() {
			hierarchy.AddEdge(a.ID(), ch.ChildID, ch.SharePercentage)
		}
	}

	_, directCashMap := NormalizeProjectInvestments(investments, homeCurrency, converter)
	rolledUpCashMap := hierarchy.CalculateRollups(directCashMap, homeCurrency)

	// Aggregate per-month actuals across the project and per-asset resource totals + trailing 6m sums.
	monthActuals := make(map[string]*ResourceBreakdown, len(months))
	for _, m := range months {
		monthActuals[m] = &ResourceBreakdown{}
	}
	assetTotals := make(map[core.AssetID]*ResourceBreakdown, len(assets))
	assetTrailing := make(map[core.AssetID]*ResourceBreakdown, len(assets))

	for _, inv := range investments {
		mKey := formatMonthKey(inv.Date)
		resType := normalizeResourceType(inv.ResourceType)
		homeAmt := convertToHomeCents(inv.Amount, homeCurrency, converter)
		inTrailing := !utcMonthStart(inv.Date).Before(trailingStartMonth) && !utcMonthStart(inv.Date).After(currentMonth)

		if len(inv.AssetAllocations) == 0 {
			if query.Tag == "" {
				if ma, ok := monthActuals[mKey]; ok {
					addResourceToBreakdown(ma, resType, homeAmt, inv.TimeHours)
				}
			}
			continue
		}

		for _, alloc := range inv.AssetAllocations {
			if alloc.AssetID.IsNil() || alloc.SharePercentage <= 0 {
				continue
			}
			if query.Tag != "" && !includedAssetIDs[alloc.AssetID] {
				continue
			}
			factor := alloc.SharePercentage / 100.0
			shareAmt := int64(math.Round(float64(homeAmt) * factor))
			shareHours := inv.TimeHours * factor

			if ma, ok := monthActuals[mKey]; ok {
				addResourceToBreakdown(ma, resType, shareAmt, shareHours)
			}

			at, ok := assetTotals[alloc.AssetID]
			if !ok {
				at = &ResourceBreakdown{}
				assetTotals[alloc.AssetID] = at
			}
			addResourceToBreakdown(at, resType, shareAmt, shareHours)

			if inTrailing {
				tr, ok := assetTrailing[alloc.AssetID]
				if !ok {
					tr = &ResourceBreakdown{}
					assetTrailing[alloc.AssetID] = tr
				}
				addResourceToBreakdown(tr, resType, shareAmt, shareHours)
			}
		}
	}

	monthScheduled := make(map[string]int64, len(months))
	monthDemands := make(map[string][]AssetDemandItem, len(months))
	monthSet := make(map[string]bool, len(months))
	for _, m := range months {
		monthSet[m] = true
	}

	var summary CashFlowSummary
	assetItems := make([]AssetForecastItem, 0, len(filteredAssets))

	for _, a := range filteredAssets {
		if a == nil {
			continue
		}
		targetHome := int64(0)
		hasTarget := a.HasTargetBudget()
		if hasTarget {
			targetHome = convertToHomeCents(a.TargetPrice(), homeCurrency, converter)
			summary.TotalTargetCents += targetHome
		}

		actual := ResourceBreakdown{}
		if at, ok := assetTotals[a.ID()]; ok {
			actual = *at
		}

		rolledUpCash := int64(0)
		if rc, ok := rolledUpCashMap[a.ID()]; ok && rc != nil {
			rolledUpCash = rc.Amount()
		}

		residual := int64(0)
		if hasTarget {
			residual = targetHome - rolledUpCash
			if residual < 0 {
				residual = 0
			}
		}

		burnRate := computeAssetBurnRate(a, assetTrailing[a.ID()], currentMonth)
		status := a.Status()

		item := AssetForecastItem{
			AssetID:                a.ID(),
			AssetName:              a.Name(),
			Status:                 status,
			StartDate:              a.StartDate(),
			TargetCompletionDate:   a.TargetCompletionDate(),
			CompletedDate:          a.CompletedDate(),
			HasTargetBudget:        hasTarget,
			TargetHomeCents:        targetHome,
			ActualToDate:           actual,
			RolledUpTotalCashCents: rolledUpCash,
			ResidualTargetCents:    residual,
			BurnRateMonthly:        burnRate,
		}

		if status != AssetStatusActive {
			item.ForecastMode = string(status)
		} else if a.TargetCompletionDate() != nil {
			targetDay := utcCalendarDay(*a.TargetCompletionDate())
			if targetDay.Before(currentMonth) {
				item.ForecastMode = "OVERDUE_TARGET"
				if residual > 0 {
					summary.TotalRemainingScheduledCents += residual
					if monthSet[currentMonthKey] {
						monthScheduled[currentMonthKey] += residual
						monthDemands[currentMonthKey] = append(monthDemands[currentMonthKey], AssetDemandItem{
							AssetID:     a.ID(),
							AssetName:   a.Name(),
							DemandCents: residual,
							IsOverdue:   true,
						})
					}
				}
			} else {
				item.ForecastMode = "SCHEDULED_TARGET"
				if residual > 0 {
					summary.TotalRemainingScheduledCents += residual
					startDay := utcCalendarDay(a.StartDate())
					effectiveStart := startDay
					if asOfDay.After(effectiveStart) {
						effectiveStart = asOfDay
					}
					if effectiveStart.After(targetDay) {
						effectiveStart = targetDay
					}
					allocated := distributeExactCentsByCalendarDays(residual, effectiveStart, targetDay)
					for mKey, demand := range allocated {
						if demand <= 0 || !monthSet[mKey] {
							continue
						}
						monthScheduled[mKey] += demand
						monthDemands[mKey] = append(monthDemands[mKey], AssetDemandItem{
							AssetID:     a.ID(),
							AssetName:   a.Name(),
							DemandCents: demand,
							IsOverdue:   false,
						})
					}
				}
			}
		} else {
			if hasTarget {
				item.ForecastMode = "ONGOING_RUN_RATE"
				if residual > 0 && burnRate.CashCentsPerMonth > 0 {
					monthsNeeded := int(math.Ceil(float64(residual) / float64(burnRate.CashCentsPerMonth)))
					est := currentMonth.AddDate(0, monthsNeeded, 0)
					item.EstimatedCompletion = &est
				}
			} else {
				item.ForecastMode = "UNBUDGETED"
			}
			summary.EstimatedOngoingMonthlyCashCents += burnRate.CashCentsPerMonth
		}

		assetItems = append(assetItems, item)
	}

	sort.SliceStable(assetItems, func(i, j int) bool {
		return assetItems[i].AssetName < assetItems[j].AssetName
	})

	monthRows := make([]MonthCashFlow, 0, len(months))
	var ceilingSum int64
	var ceilingMonths int64

	for _, mKey := range months {
		mTime, _ := parseMonthStart(mKey)
		actual := *monthActuals[mKey]
		scheduled := monthScheduled[mKey]

		ongoing := int64(0)
		if !mTime.Before(currentMonth) {
			ongoing = summary.EstimatedOngoingMonthlyCashCents
		}

		committed := actual.CashCents + scheduled
		ceiling := int64(0)
		if ceilings != nil {
			ceiling = ceilings[mKey]
		}
		if ceiling > 0 {
			ceilingSum += ceiling
			ceilingMonths++
		}

		alertStatus, utilPct := evaluateCeilingAlert(committed, ceiling)
		if alertStatus == "AT_RISK" || alertStatus == "OVER_CEILING" {
			summary.AlertMonthsCount++
		}
		if committed > summary.PeakCommittedCashCents {
			summary.PeakCommittedCashCents = committed
			summary.PeakMonth = mKey
		}

		summary.TotalActualCashCents += actual.CashCents
		summary.TotalTimeHours += actual.TimeHours
		summary.TotalGoodsCents += actual.GoodsCents

		demands := monthDemands[mKey]
		if demands == nil {
			demands = make([]AssetDemandItem, 0)
		}

		monthRows = append(monthRows, MonthCashFlow{
			Month:                     mKey,
			TagBudgetCeilingCents:     ceiling,
			Actual:                    actual,
			ScheduledDemandCents:      scheduled,
			EstimatedOngoingCashCents: ongoing,
			CommittedCashCents:        committed,
			CashUtilizationPct:        utilPct,
			AlertStatus:               alertStatus,
			AssetDemands:              demands,
		})
	}

	if ceilingMonths > 0 {
		summary.MonthlyBudgetCeilingAvgCents = ceilingSum / ceilingMonths
	}

	fromStr := ""
	toStr := ""
	if len(months) > 0 {
		fromStr = months[0]
		toStr = months[len(months)-1]
	}

	return &CashFlowForecast{
		ProjectID:            project.ID,
		Currency:             homeCurrency,
		FromMonth:            fromStr,
		ToMonth:              toStr,
		TrailingWindowMonths: defaultTrailingWindowMonths,
		Summary:              summary,
		Months:               monthRows,
		Assets:               assetItems,
	}
}

func filterAssetsByTag(assets []*Asset, tag string) ([]*Asset, map[core.AssetID]bool) {
	trimmed := strings.TrimSpace(tag)
	included := make(map[core.AssetID]bool, len(assets))
	if trimmed == "" {
		out := make([]*Asset, 0, len(assets))
		for _, a := range assets {
			if a != nil {
				out = append(out, a)
				included[a.ID()] = true
			}
		}
		return out, included
	}
	out := make([]*Asset, 0)
	for _, a := range assets {
		if a == nil {
			continue
		}
		for _, t := range a.Tags() {
			if strings.EqualFold(t, trimmed) {
				out = append(out, a)
				included[a.ID()] = true
				break
			}
		}
	}
	return out, included
}

func addResourceToBreakdown(dst *ResourceBreakdown, resType string, amt int64, hours float64) {
	switch resType {
	case "TIME":
		dst.TimeHours += hours
	case "GOODS":
		dst.GoodsCents += amt
	default:
		dst.CashCents += amt
	}
}

func computeAssetBurnRate(a *Asset, trailing *ResourceBreakdown, currentMonth time.Time) MonthlyBurnRate {
	if trailing == nil {
		return MonthlyBurnRate{}
	}
	startMonth := utcMonthStart(a.StartDate())
	elapsed := monthsBetweenInclusive(startMonth, currentMonth)
	if elapsed < 1 {
		elapsed = 1
	}
	if elapsed > defaultTrailingWindowMonths {
		elapsed = defaultTrailingWindowMonths
	}
	denom := float64(elapsed)
	return MonthlyBurnRate{
		CashCentsPerMonth:  int64(math.Round(float64(trailing.CashCents) / denom)),
		TimeHoursPerMonth:  math.Round((trailing.TimeHours/denom)*100) / 100,
		GoodsCentsPerMonth: int64(math.Round(float64(trailing.GoodsCents) / denom)),
	}
}

func evaluateCeilingAlert(committed, ceiling int64) (string, float64) {
	if ceiling <= 0 {
		return "NO_CEILING", 0
	}
	util := (float64(committed) / float64(ceiling)) * 100.0
	roundedUtil := math.Round(util*10) / 10
	if committed > ceiling {
		return "OVER_CEILING", roundedUtil
	}
	if float64(committed) >= float64(ceiling)*atRiskThresholdRatio {
		return "AT_RISK", roundedUtil
	}
	return "HEALTHY", roundedUtil
}

type monthDaySlice struct {
	monthKey  string
	days      int64
	remainder int64
	alloc     int64
}

// distributeExactCentsByCalendarDays spreads totalCents proportionally across inclusive UTC
// calendar days in each overlapping month of [startDay, endDay], preserving exact cent conservation.
func distributeExactCentsByCalendarDays(totalCents int64, startDay, endDay time.Time) map[string]int64 {
	res := make(map[string]int64)
	if totalCents <= 0 || endDay.Before(startDay) {
		return res
	}

	totalDays := int64(endDay.Sub(startDay).Hours()/24) + 1
	if totalDays <= 0 {
		totalDays = 1
	}

	slices := make([]monthDaySlice, 0)
	curMonth := utcMonthStart(startDay)
	endMonth := utcMonthStart(endDay)

	for !curMonth.After(endMonth) {
		monthLastDay := curMonth.AddDate(0, 1, -1)
		overlapStart := startDay
		if curMonth.After(overlapStart) {
			overlapStart = curMonth
		}
		overlapEnd := endDay
		if monthLastDay.Before(overlapEnd) {
			overlapEnd = monthLastDay
		}
		if !overlapEnd.Before(overlapStart) {
			d := int64(overlapEnd.Sub(overlapStart).Hours()/24) + 1
			numer := totalCents * d
			slices = append(slices, monthDaySlice{
				monthKey:  formatMonthKey(curMonth),
				days:      d,
				alloc:     numer / totalDays,
				remainder: numer % totalDays,
			})
		}
		curMonth = curMonth.AddDate(0, 1, 0)
	}

	var sumAllocated int64
	for _, s := range slices {
		sumAllocated += s.alloc
	}
	leftover := totalCents - sumAllocated
	if leftover > 0 {
		order := make([]int, len(slices))
		for i := range order {
			order[i] = i
		}
		sort.SliceStable(order, func(i, j int) bool {
			if slices[order[i]].remainder != slices[order[j]].remainder {
				return slices[order[i]].remainder > slices[order[j]].remainder
			}
			return order[i] < order[j]
		})
		for i := int64(0); i < leftover && i < int64(len(order)); i++ {
			slices[order[i]].alloc++
		}
	}

	for _, s := range slices {
		res[s.monthKey] = s.alloc
	}
	return res
}
